package goadmin

import (
	"embed"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"github.com/CloudyKit/jet"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/partyzanex/go-admin-bootstrap/assets"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
	"github.com/partyzanex/go-admin-bootstrap/views"
)

type App struct {
	config *Config

	echo   *echo.Echo
	static *echo.Group
	admin  *echo.Group

	baseURL *url.URL
}

func New(config Config) (*App, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	if err := migrations.Up(config.DBConfig.DB, config.DBConfig.MigrationsTable); err != nil {
		return nil, errors.Wrap(err, "cannot up migrations")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse %q as base url", config.BaseURL)
	}

	app := new(App)
	app.config = &config
	app.echo = echo.New()
	app.baseURL = baseURL

	app.setStaticGroup()
	app.setDefaultRoutes()
	app.setDefaultMiddleware()
	app.setDefaultRenderer()

	err = app.CreateAssets()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sources")
	}

	return app, nil
}

func (app *App) Static() *echo.Group {
	return app.static
}

func (app *App) Admin() *echo.Group {
	return app.admin
}

func (app *App) Echo() *echo.Echo {
	return app.echo
}

func (app *App) Serve() error {
	return app.echo.Start(app.getAddr())
}

func (app *App) CreateAssets() error {
	assetsByKind := make(map[AssetKind][]*Asset)

	for _, source := range app.config.Assets {
		_, ok := assetsByKind[source.Kind]
		if !ok {
			assetsByKind[source.Kind] = []*Asset{source}

			continue
		}

		assetsByKind[source.Kind] = append(assetsByKind[source.Kind], source)
	}

	var (
		javascriptSources = JS
		stylesheetSources = CSS
		viewSources       = Views
	)

	if sources, ok := assetsByKind[JavaScript]; ok {
		javascriptSources = append(javascriptSources, sources...)
	}

	sort.Slice(javascriptSources, func(i, j int) bool {
		return javascriptSources[i].SortOrder < javascriptSources[j].SortOrder
	})

	for _, source := range javascriptSources {
		err := app.createSource(app.config.AssetsPath, source, &assets.JS)
		if err != nil {
			return errors.Wrapf(err, "cannot create source %s", source.Path)
		}
	}

	if sources, ok := assetsByKind[Stylesheet]; ok {
		stylesheetSources = append(stylesheetSources, sources...)
	}

	sort.Slice(stylesheetSources, func(i, j int) bool {
		return stylesheetSources[i].SortOrder < stylesheetSources[j].SortOrder
	})

	for _, source := range stylesheetSources {
		err := app.createSource(app.config.AssetsPath, source, &assets.CSS)
		if err != nil {
			return errors.Wrapf(err, "cannot create source %q", source.Path)
		}
	}

	if sources, ok := assetsByKind[View]; ok {
		viewSources = append(viewSources, sources...)
	}

	for _, source := range viewSources {
		err := app.createSource(app.config.ViewsPath, source, &views.Sources)
		if err != nil {
			// todo: wrap error
			return err
		}
	}

	return nil
}

func (*App) createSource(path string, source *Asset, fs *embed.FS) error {
	sourcePath := filepath.Join(path, source.Path)

	_, err := os.Stat(sourcePath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "loading asset %s source failed", source.Path)
	}

	if os.IsNotExist(err) {
		sourceDir := filepath.Dir(sourcePath)

		b, err := fs.ReadFile(source.Path)
		if err != nil {
			return errors.Wrapf(err, "cannot read file %q", source.Path)
		}

		err = os.MkdirAll(sourceDir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "make assets dir %s failed", sourceDir)
		}

		err = ioutil.WriteFile(sourcePath, b, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "cannot write file %q", sourcePath)
		}
	}

	return nil
}

func (app *App) getAddr() string {
	return fmt.Sprintf("%s:%d", app.config.Host, app.config.Port)
}

func (app *App) setStaticGroup() {
	if app.config.AssetsPath == "" {
		app.config.AssetsPath = DefaultAssetsPath
	}

	if app.config.ViewsPath == "" {
		app.config.ViewsPath = DefaultViewsPath
	}

	app.static = app.echo.Group(app.baseURL.Path + "/assets")
	app.static.Static("/", app.config.AssetsPath)
}

func (app *App) setDefaultRoutes() {
	app.admin = app.echo.Group(app.baseURL.Path, withViewData)
	app.admin.GET(LoginURL, WrapHandler(Login))
	app.admin.POST(LoginURL, WrapHandler(Login))

	auth := AuthByCookie

	app.admin.Any(LogoutURL, WrapHandler(Logout), auth)
	app.admin.GET(DashboardURL, WrapHandler(Dashboard), auth)
	app.admin.GET(UserListURL, WrapHandler(UserList), auth)
	app.admin.GET(UserCreateURL, WrapHandler(UserCreate), auth)
	app.admin.POST(UserCreateURL, WrapHandler(UserCreate), auth)
	app.admin.GET(UserDeleteURL, WrapHandler(UserDelete), auth)
	app.admin.GET(UserUpdateURL, WrapHandler(UserUpdate), auth)
	app.admin.POST(UserUpdateURL, WrapHandler(UserUpdate), auth)
}

func (app *App) setDefaultMiddleware() {
	for _, mw := range app.config.Middleware {
		app.echo.Use(mw)
	}

	app.echo.Use(withAppContext(app))
}

func (app *App) setDefaultRenderer() {
	renderer := &Renderer{
		Views: jet.NewHTMLSet(app.config.ViewsPath),
	}

	renderer.Views.SetDevelopmentMode(app.config.DevMode)
	renderer.Views.AddGlobal("adminPath", app.baseURL.Path)
	renderer.Views.AddGlobal("loginURL", LoginURL)
	renderer.Views.AddGlobal("logoutURL", LogoutURL)
	renderer.Views.AddGlobal("userListURL", UserListURL)

	app.echo.Renderer = renderer
}
