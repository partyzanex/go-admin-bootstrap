package goadmin

import (
	"cmp"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/CloudyKit/jet/v6"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/partyzanex/go-admin-bootstrap/assets"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
	"github.com/partyzanex/go-admin-bootstrap/views"
)

type App struct {
	config *Config
	logger *slog.Logger

	echo   *echo.Echo
	static *echo.Group
	admin  *echo.Group

	baseURL *url.URL
}

func New(config *Config) (*App, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if err := migrations.Up(config.DBConfig.DB.DB, config.DBConfig.MigrationsTable); err != nil {
		return nil, fmt.Errorf("cannot up migrations: %w", err)
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q as base url: %w", config.BaseURL, err)
	}

	e := echo.New()
	e.HTTPErrorHandler = errorHandler

	app := new(App)
	app.config = config.Clone()
	app.echo = e
	app.baseURL = baseURL

	if app.config.AccessCookieName == "" {
		app.config.AccessCookieName = DefaultAccessCookieName
	}

	if app.config.DBConfig.MigrationsTable == "" {
		app.config.DBConfig.MigrationsTable = DefaultMigrationsTable
	}

	if config.Logger != nil {
		app.logger = config.Logger
	} else {
		app.logger = slog.Default()
	}

	app.setStaticGroup()
	app.setDefaultRoutes()
	app.setDefaultMiddleware()
	app.setDefaultRenderer()

	err = app.CreateAssets()
	if err != nil {
		return nil, fmt.Errorf("cannot create sources: %w", err)
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

func (app *App) Close() error {
	return app.echo.Close()
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

	slices.SortFunc(javascriptSources, func(a, b *Asset) int {
		return cmp.Compare(a.SortOrder, b.SortOrder)
	})

	for _, source := range javascriptSources {
		err := app.createSource(app.config.AssetsPath, source, &assets.JS)
		if err != nil {
			return fmt.Errorf("cannot create source %s: %w", source.Path, err)
		}
	}

	if sources, ok := assetsByKind[Stylesheet]; ok {
		stylesheetSources = append(stylesheetSources, sources...)
	}

	slices.SortFunc(stylesheetSources, func(a, b *Asset) int {
		return cmp.Compare(a.SortOrder, b.SortOrder)
	})

	for _, source := range stylesheetSources {
		err := app.createSource(app.config.AssetsPath, source, &assets.CSS)
		if err != nil {
			return fmt.Errorf("cannot create source %q: %w", source.Path, err)
		}
	}

	if sources, ok := assetsByKind[View]; ok {
		viewSources = append(viewSources, sources...)
	}

	for _, source := range viewSources {
		err := app.createSource(app.config.ViewsPath, source, &views.Sources)
		if err != nil {
			return fmt.Errorf("cannot create view source %s: %w", source.Path, err)
		}
	}

	return nil
}

func (*App) createSource(path string, source *Asset, fs *embed.FS) (err error) {
	sourcePath := filepath.Join(path, source.Path)

	stat, err := os.Stat(sourcePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading asset %s source failed: %w", source.Path, err)
	}

	if stat != nil {
		return nil
	}

	var (
		b         []byte
		sourceDir = filepath.Dir(sourcePath)
	)

	b, err = fs.ReadFile(source.Path)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %w", source.Path, err)
	}

	err = os.MkdirAll(sourceDir, 0o750) //nolint:mnd // standard directory permission
	if err != nil {
		return fmt.Errorf("make assets dir %s failed: %w", sourceDir, err)
	}

	err = os.WriteFile(sourcePath, b, 0o600) //nolint:mnd // standard file permission
	if err != nil {
		return fmt.Errorf("cannot write file %q: %w", sourcePath, err)
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

	app.static = app.echo.Group(app.baseURL.Path + assetsRelativePath)
	app.static.Static("/", app.config.AssetsPath)
}

func (app *App) setDefaultRoutes() {
	app.admin = app.echo.Group(app.baseURL.Path, withViewData)

	// Rate-limited login routes
	loginGroup := app.admin.Group(LoginURL, echomiddleware.RateLimiter(
		echomiddleware.NewRateLimiterMemoryStore(LoginRateLimitPerSec),
	))
	loginGroup.GET("", WrapHandler(Login))
	loginGroup.POST("", WrapHandler(Login))

	app.admin.Any(LogoutURL, WrapHandler(Logout), AuthByCookie)
	app.admin.GET(DashboardURL, WrapHandler(Dashboard), AuthByCookie)
	app.admin.GET(UserListURL, WrapHandler(UserList), AuthByCookie)
	app.admin.GET(UserCreateURL, WrapHandler(UserCreate), AuthByCookie)
	app.admin.POST(UserCreateURL, WrapHandler(UserCreate), AuthByCookie)
	app.admin.GET(UserDeleteURL, WrapHandler(UserDelete), AuthByCookie)
	app.admin.GET(UserUpdateURL, WrapHandler(UserUpdate), AuthByCookie)
	app.admin.POST(UserUpdateURL, WrapHandler(UserUpdate), AuthByCookie)
	app.admin.GET(FaviconPrefix, Favicon)
}

func (app *App) setDefaultMiddleware() {
	for _, mw := range app.config.Middleware {
		app.echo.Use(mw)
	}

	app.echo.Use(withAppContext(app))
	app.echo.Use(withRequestLogger(app.logger))

	// CSRF protection for all admin routes
	app.admin.Use(echomiddleware.CSRFWithConfig(echomiddleware.CSRFConfig{
		TokenLength:    32,
		TokenLookup:    "form:_csrf,header:X-CSRF-Token",
		CookiePath:     "/",
		CookieSecure:   !app.config.DevMode,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
	}))
}

func (app *App) setDefaultRenderer() {
	opts := make([]jet.Option, 0)

	if app.config.DevMode {
		opts = append(opts, jet.InDevelopmentMode())
	}

	renderer := &Renderer{
		Views: jet.NewSet(
			jet.NewOSFileSystemLoader(app.config.ViewsPath),
			opts...,
		),
	}

	renderer.Views.AddGlobal(adminPathVar, app.baseURL.Path)
	renderer.Views.AddGlobal(loginURLVar, LoginURL)
	renderer.Views.AddGlobal(logoutURLVar, LogoutURL)
	renderer.Views.AddGlobal(userListURLVar, UserListURL)

	app.echo.Renderer = renderer
}
