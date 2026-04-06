package goadmin

import (
	"cmp"
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"time"

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

func New(config *Config, opts ...Option) (*App, error) {
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

	app := &App{
		config:  config.Clone(),
		echo:    e,
		baseURL: baseURL,
	}

	for _, opt := range opts {
		if optErr := opt(app); optErr != nil {
			return nil, fmt.Errorf("applying option: %w", optErr)
		}
	}

	app.applyDefaults()
	app.setStaticGroup()
	app.setDefaultMiddleware()
	app.setDefaultRoutes()
	app.setDefaultRenderer()

	err = app.CreateAssets()
	if err != nil {
		return nil, fmt.Errorf("cannot create sources: %w", err)
	}

	return app, nil
}

func (app *App) applyDefaults() {
	if app.config.AccessCookieName == "" {
		app.config.AccessCookieName = DefaultAccessCookieName
	}

	if app.config.DBConfig.MigrationsTable == "" {
		app.config.DBConfig.MigrationsTable = DefaultMigrationsTable
	}

	if app.config.AccessTokenTTL == 0 {
		app.config.AccessTokenTTL = DefaultAccessTokenTTL
	}

	if app.config.RefreshTokenTTL == 0 {
		app.config.RefreshTokenTTL = DefaultRefreshTokenTTL
	}

	if app.config.Logger != nil {
		app.logger = app.config.Logger
	} else {
		app.logger = slog.Default()
	}
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

const defaultShutdownTimeout = 10 * time.Second

func (app *App) Serve(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := app.echo.Start(app.getAddr()); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}

		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		return app.echo.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (app *App) Close() error {
	return app.echo.Close()
}

func (app *App) CreateAssets() error {
	assetsByKind := make(map[AssetKind][]*Asset)

	for _, source := range app.config.assets {
		assetsByKind[source.Kind] = append(assetsByKind[source.Kind], source)
	}

	// Views always need to be on disk (Jet uses OSFileSystemLoader)
	viewSources := Views
	if sources, ok := assetsByKind[View]; ok {
		viewSources = append(viewSources, sources...)
	}

	for _, source := range viewSources {
		err := app.createSource(app.config.ViewsPath, source, &views.Sources)
		if err != nil {
			return fmt.Errorf("cannot create view source %s: %w", source.Path, err)
		}
	}

	// JS/CSS: copy to disk only in dev mode (prod serves from embed.FS)
	if app.config.DevMode {
		if err := app.createStaticAssets(assetsByKind); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) createStaticAssets(assetsByKind map[AssetKind][]*Asset) error {
	javascriptSources := JS
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

	stylesheetSources := CSS
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

	return nil
}

func (*App) createSource(path string, source *Asset, embedFS *embed.FS) (err error) {
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

	b, err = embedFS.ReadFile(source.Path)
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

	if app.config.DevMode {
		app.static.Static("/", app.config.AssetsPath)
	} else {
		app.static.StaticFS("/", &mergedFS{filesystems: []embed.FS{assets.CSS, assets.JS}})
	}
}

type mergedFS struct {
	filesystems []embed.FS
}

func (m *mergedFS) Open(name string) (fs.File, error) {
	for _, fsys := range m.filesystems {
		f, err := fsys.Open(name)
		if err == nil {
			return f, nil
		}
	}

	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
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

	adminOnly := RequireRole(RoleOwner, RoleRoot)
	app.admin.GET(UserListURL, WrapHandler(UserList), AuthByCookie, adminOnly)
	app.admin.GET(UserCreateURL, WrapHandler(UserCreate), AuthByCookie, adminOnly)
	app.admin.POST(UserCreateURL, WrapHandler(UserCreate), AuthByCookie, adminOnly)
	app.admin.GET(UserDeleteURL, WrapHandler(UserDelete), AuthByCookie, adminOnly)
	app.admin.GET(UserUpdateURL, WrapHandler(UserUpdate), AuthByCookie, adminOnly)
	app.admin.POST(UserUpdateURL, WrapHandler(UserUpdate), AuthByCookie, adminOnly)
	app.admin.GET(FaviconPrefix, Favicon)

	app.echo.GET(app.baseURL.Path+"/health", app.healthCheck)
}

const healthCheckTimeout = 2 * time.Second

func (app *App) healthCheck(ctx echo.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx.Request().Context(), healthCheckTimeout)
	defer cancel()

	if err := app.config.DBConfig.DB.PingContext(checkCtx); err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error", "db": "down"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (app *App) setDefaultMiddleware() {
	for _, mw := range app.config.middleware {
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
