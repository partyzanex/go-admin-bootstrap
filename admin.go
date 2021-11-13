package goadmin

import (
	"fmt"
	"net/url"

	"github.com/CloudyKit/jet"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	migration "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
)

type Admin struct {
	*Config

	e       *echo.Echo
	static  *echo.Group
	admin   *echo.Group
	baseURL *url.URL
}

func (a *Admin) Serve() error {
	if err := a.hasEcho(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", a.Host, a.Port)

	return a.e.Start(addr)
}

func (a *Admin) Echo() *echo.Echo {
	return a.e
}

func (a *Admin) Static() *echo.Group {
	return a.static
}

func (a *Admin) Admin() *echo.Group {
	return a.admin
}

func (a *Admin) configure() error {
	a.configureMiddleware()
	a.configureRenderer()
	a.configureErrorHandler()
	a.configureAssets()
	a.configureRoutes()

	if err := a.configureDatabase(); err != nil {
		return errors.Wrap(err, "configure database failed")
	}

	return nil
}

func (a *Admin) configureDatabase() error {
	return migration.Up(a.DBConfig.DB, a.DBConfig.MigrationsTable)
}

func (a *Admin) configureRoutes() {
	a.admin = a.e.Group(a.baseURL.Path, withViewData)
	a.admin.GET(LoginURL, WrapHandler(Login))
	a.admin.POST(LoginURL, WrapHandler(Login))

	auth := AuthByCookie

	a.admin.Any(LogoutURL, WrapHandler(Logout), auth)
	a.admin.GET(DashboardURL, WrapHandler(Dashboard), auth)
	a.admin.GET(UserListURL, WrapHandler(UserList), auth)
	a.admin.GET(UserCreateURL, WrapHandler(UserCreate), auth)
	a.admin.POST(UserCreateURL, WrapHandler(UserCreate), auth)
	a.admin.GET(UserDeleteURL, WrapHandler(UserDelete), auth)
	a.admin.GET(UserUpdateURL, WrapHandler(UserUpdate), auth)
	a.admin.POST(UserUpdateURL, WrapHandler(UserUpdate), auth)
}

func (a *Admin) configureMiddleware() {
	for _, mw := range a.Middleware {
		a.e.Use(mw)
	}

	a.e.Use(withAdminContext(a))
}

func (a *Admin) configureErrorHandler() {
	a.e.HTTPErrorHandler = errorHandler
}

func (a *Admin) configureRenderer() {
	renderer := &Renderer{
		Views: jet.NewHTMLSet(a.ViewsPath),
	}

	renderer.Views.SetDevelopmentMode(a.DevMode)
	renderer.Views.AddGlobal("adminPath", a.baseURL.Path)
	renderer.Views.AddGlobal("loginURL", LoginURL)
	renderer.Views.AddGlobal("logoutURL", LogoutURL)
	renderer.Views.AddGlobal("userListURL", UserListURL)

	a.e.Renderer = renderer
}

func (a *Admin) configureAssets() {
	if a.AssetsPath == "" {
		a.AssetsPath = DefaultAssetsPath
	}

	if a.ViewsPath == "" {
		a.ViewsPath = DefaultViewsPath
	}

	a.static = a.e.Group(a.baseURL.Path + "/assets")
	a.static.Static("/", a.AssetsPath)
}

func (a *Admin) hasEcho() error {
	if a.e == nil {
		return errors.New("please use goadmin.New when creating Admin")
	}

	return nil
}

func New(config Config) (*Admin, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "validation failed")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse base url failed")
	}

	a := &Admin{
		Config:  &config,
		e:       echo.New(),
		baseURL: baseURL,
	}

	if err := a.configure(); err != nil {
		return nil, err
	}

	if err := a.LoadSources(); err != nil {
		return nil, err
	}

	return a, nil
}
