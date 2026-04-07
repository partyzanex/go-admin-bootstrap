package goadmin

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/partyzanex/go-admin-bootstrap/assets"
	"github.com/partyzanex/go-admin-bootstrap/views"
)

// --- applyDefaults ---

func TestApplyDefaults_FillsEmptyValues(t *testing.T) {
	app := &App{
		config: &Config{},
		echo:   echo.New(),
	}

	app.applyDefaults()

	assert.Equal(t, DefaultAccessCookieName, app.config.AccessCookieName)
	assert.Equal(t, DefaultAccessTokenTTL, app.config.AccessTokenTTL)
	assert.Equal(t, DefaultRefreshTokenTTL, app.config.RefreshTokenTTL)
	assert.NotNil(t, app.logger)
}

func TestApplyDefaults_PreservesExistingValues(t *testing.T) {
	customLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	app := &App{
		config: &Config{
			AccessCookieName: "my_cookie",
			AccessTokenTTL:   5 * time.Minute,
			RefreshTokenTTL:  7 * 24 * time.Hour,
			Logger:           customLogger,
		},
		echo: echo.New(),
	}

	app.applyDefaults()

	assert.Equal(t, "my_cookie", app.config.AccessCookieName)
	assert.Equal(t, 5*time.Minute, app.config.AccessTokenTTL)
	assert.Equal(t, 7*24*time.Hour, app.config.RefreshTokenTTL)
	assert.Equal(t, customLogger, app.logger)
}

// --- Getters ---

func TestApp_Getters(t *testing.T) {
	app := newTestApp(t)

	app.setStaticGroup()
	app.setDefaultRoutes()

	assert.Equal(t, app.echo, app.Echo())
	assert.NotNil(t, app.Static())
	assert.NotNil(t, app.Admin())
}

// --- getAddr ---

func TestGetAddr(t *testing.T) {
	app := &App{
		config: &Config{Host: "0.0.0.0", Port: 8080},
	}
	assert.Equal(t, "0.0.0.0:8080", app.getAddr())
}

func TestGetAddr_DefaultHost(t *testing.T) {
	app := &App{
		config: &Config{Port: 9090},
	}
	assert.Equal(t, ":9090", app.getAddr())
}

// --- Close ---

func TestApp_Close(t *testing.T) {
	app := newTestApp(t)
	assert.NoError(t, app.Close())
}

// --- mergedFS ---

func TestMergedFS_Open_FoundInFirst(t *testing.T) {
	mfs := &mergedFS{filesystems: []embed.FS{assets.JS, assets.CSS}}

	f, err := mfs.Open("plugins/jquery/jquery-3.4.1.min.js")
	require.NoError(t, err)
	assert.NotNil(t, f)
	require.NoError(t, f.Close())
}

func TestMergedFS_Open_FoundInSecond(t *testing.T) {
	mfs := &mergedFS{filesystems: []embed.FS{assets.JS, assets.CSS}}

	// css/style.css is in assets.CSS, not assets.JS
	f, err := mfs.Open("css/style.css")
	require.NoError(t, err)
	assert.NotNil(t, f)
	require.NoError(t, f.Close())
}

func TestMergedFS_Open_NotFound(t *testing.T) {
	mfs := &mergedFS{filesystems: []embed.FS{assets.JS, assets.CSS}}

	_, err := mfs.Open("nonexistent/file.txt")
	assert.Error(t, err)
}

// --- createSource ---

func TestCreateSource_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	app := &App{}

	asset := &Asset{Path: "auth/login.jet", Kind: View}
	err := app.createSource(dir, asset, &views.Sources)

	require.NoError(t, err)
	_, statErr := os.Stat(filepath.Join(dir, "auth/login.jet"))
	assert.NoError(t, statErr)
}

func TestCreateSource_FileAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	app := &App{}

	asset := &Asset{Path: "auth/login.jet", Kind: View}

	// First call — creates the file
	require.NoError(t, app.createSource(dir, asset, &views.Sources))

	// Second call — file already exists, should return nil without overwriting
	err := app.createSource(dir, asset, &views.Sources)
	assert.NoError(t, err)
}

func TestCreateSource_EmbedReadError(t *testing.T) {
	dir := t.TempDir()
	app := &App{}

	// Non-existent path in embed.FS
	asset := &Asset{Path: "nonexistent/file.jet", Kind: View}
	err := app.createSource(dir, asset, &views.Sources)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read file")
}

// --- CreateAssets ---

func TestCreateAssets_DevModeOff(t *testing.T) {
	dir := t.TempDir()
	app := newTestApp(t)
	app.config.ViewsPath = dir
	app.config.DevMode = false

	err := app.CreateAssets()
	require.NoError(t, err)

	// In non-dev mode only view files are copied
	_, err = os.Stat(filepath.Join(dir, "auth/login.jet"))
	assert.NoError(t, err)
}

func TestCreateAssets_DevModeOn(t *testing.T) {
	dir := t.TempDir()
	app := newTestApp(t)
	app.config.ViewsPath = dir
	app.config.AssetsPath = dir
	app.config.DevMode = true

	err := app.CreateAssets()
	require.NoError(t, err)

	// In dev mode JS/CSS are also copied
	_, err = os.Stat(filepath.Join(dir, "css/style.css"))
	assert.NoError(t, err)
}

// --- runTokenCleanup ---

func TestRunTokenCleanup_ContextCancel(t *testing.T) {
	app := newTestApp(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	done := make(chan struct{})

	go func() {
		app.runTokenCleanup(ctx)
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Fatal("runTokenCleanup did not finish in time")
	}
}

// --- Serve ---

func TestServe_ContextCancel(t *testing.T) {
	app := newTestApp(t)

	port := getFreePort(t)
	app.config.Port = uint16(port)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)

	go func() {
		done <- app.Serve(ctx)
	}()

	// Give the server time to start
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		// Shutdown on a closed server or nil
		assert.True(t, err == nil || errors.Is(err, http.ErrServerClosed))
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not finish after context cancellation")
	}
}

func TestServe_StartError(t *testing.T) {
	// Occupy the port with a persistent listener so that Start returns an error
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, ln.Close()) })

	port := ln.Addr().(*net.TCPAddr).Port

	app := newTestApp(t)
	app.config.Port = uint16(port)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = app.Serve(ctx)
	assert.Error(t, err)
}

// --- setDefaultMiddleware / setDefaultRoutes ---

func TestSetDefaultMiddleware(t *testing.T) {
	app := newTestApp(t)
	// Should not panic
	app.setDefaultMiddleware()
}

func TestSetDefaultRenderer(t *testing.T) {
	app := newTestApp(t)
	app.config.ViewsPath = t.TempDir()
	// Should not panic
	app.setDefaultRenderer()
	assert.NotNil(t, app.echo.Renderer)
}

// errPinger is a DBPinger stub that always returns an error.
type errPinger struct{}

func (e *errPinger) PingContext(_ context.Context) error {
	return errors.New("db down")
}

func TestHealthCheck_DBDown(t *testing.T) {
	app := newTestApp(t)
	app.config.DBConfig.DB = &errPinger{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)

	err := app.healthCheck(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestHealthCheck_NoDB(t *testing.T) {
	app := newTestApp(t)
	// DB is nil — health check skips DB ping and returns 200

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)

	err := app.healthCheck(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- helpers ---

func getFreePort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())

	return port
}
