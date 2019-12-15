package goadmin

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"os"
	"path/filepath"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/CloudyKit/jet"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

const (
	DefaultAssetsPath               = "./assets"
	DefaultViewsPath                = "./views"
	DefaultCacheTTL   time.Duration = 30 * time.Minute
	DefaultCacheClean time.Duration = 5 * time.Second
	DefaultLimit                    = 20
	Version                         = "v0.0.1"
)

type AdminHandler func(ctx *AdminContext) error

type Admin struct {
	*Config

	e      *echo.Echo
	static *echo.Group
	group  *echo.Group
	cache  *cache.Cache
}

func (admin *Admin) Serve() error {
	if err := admin.hasEcho(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", admin.Host, admin.Port)
	return admin.e.Start(addr)
}

func (admin *Admin) Echo() *echo.Echo {
	return admin.e
}

func (admin *Admin) configure() error {
	admin.configureCache()
	admin.configureMiddleware()
	admin.configureRenderer()
	admin.configureErrorHandler()
	admin.configureAssets()
	admin.configureRoutes()

	if err := admin.configureDatabase(); err != nil {
		return errors.Wrap(err, "configure database failed")
	}

	return nil
}

func (admin *Admin) configureDatabase() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(admin.DBConfig.DB, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, "creating postgres instance failed")
	}

	migrationsPath := filepath.Join(dir, admin.DBConfig.MigrationsPath)

	m, err := migrate.NewWithDatabaseInstance(
		"file:///"+migrationsPath,
		admin.DBConfig.DriverName,
		driver,
	)

	if err != nil {
		return errors.Wrap(err, "creating database instance failed")
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "to migrate up failed")
	}

	return nil
}

func (admin *Admin) configureRoutes() {
	admin.group = admin.e.Group(admin.baseURL.Path, withViewData)
	admin.group.GET(LoginURL, WrapHandler(Login))
	admin.group.POST(LoginURL, WrapHandler(Login))
	admin.group.Any(LogoutURL, WrapHandler(Logout), AuthByCookie)
	admin.group.GET(DashboardURL, WrapHandler(Dashboard), AuthByCookie)
	admin.group.GET(UserListURL, WrapHandler(UserList), AuthByCookie)
}

func (admin *Admin) configureMiddleware() {
	for _, mw := range admin.Middleware {
		admin.e.Use(mw)
	}

	admin.e.Use(middleware.Recover())
	admin.e.Use(middleware.Logger())
	admin.e.Use(withAdminContext(admin))
}

func (admin *Admin) configureErrorHandler() {
	admin.e.HTTPErrorHandler = errorHandler
}

func (admin *Admin) configureRenderer() {
	renderer := &Renderer{
		Views: jet.NewHTMLSet(admin.ViewsPath),
	}
	renderer.Views.SetDevelopmentMode(admin.DevMode)
	renderer.Views.AddGlobal("adminPath", admin.baseURL.Path)
	renderer.Views.AddGlobal("loginURL", LoginURL)
	renderer.Views.AddGlobal("logoutURL", LogoutURL)
	renderer.Views.AddGlobal("userListURL", UserListURL)

	admin.e.Renderer = renderer
}

func (admin *Admin) configureAssets() {
	if admin.AssetsPath == "" {
		admin.AssetsPath = DefaultAssetsPath
	}

	admin.static = admin.e.Group(admin.baseURL.Path + "/assets")
	admin.static.Static("/", admin.AssetsPath)
}

func (admin *Admin) configureCache() {
	if admin.CacheTTL == 0 {
		admin.CacheTTL = DefaultCacheTTL
	}

	if admin.CacheClean == 0 {
		admin.CacheClean = DefaultCacheClean
	}

	admin.cache = cache.New(admin.CacheTTL, admin.CacheClean)
}

func (admin *Admin) hasEcho() error {
	if admin.e == nil {
		return errors.New("please use goadmin.New when creating Admin")
	}

	return nil
}

func New(config Config) (*Admin, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "validation failed")
	}

	admin := &Admin{
		Config: &config,
		e:      echo.New(),
	}

	if err := admin.configure(); err != nil {
		return nil, err
	}

	if err := admin.LoadSources(); err != nil {
		return nil, err
	}

	return admin, nil
}

func WrapHandler(h AdminHandler) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac, ok := ctx.(*AdminContext)
		if !ok {
			return ErrContextNotConfigured
		}

		return h(ac)
	}
}
