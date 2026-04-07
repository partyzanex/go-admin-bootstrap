package goadmin

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	baseURL, _ := url.Parse("http://localhost/admin")

	return &App{
		config: &Config{
			AccessCookieName: "test",
			JWTSecret:        []byte("test-secret-32-bytes-long!!!!!!!!"),
			AccessTokenTTL:   15 * time.Minute,
			RefreshTokenTTL:  24 * time.Hour,
			UserCase:         &stubUseCase{},
		},
		baseURL: baseURL,
		echo:    echo.New(),
		logger:  slog.Default(),
	}
}

func TestWrapHandler(t *testing.T) {
	app := newTestApp(t)

	handler := WrapHandler(func(ctx *AppContext) error {
		return ctx.String(http.StatusOK, "ok")
	})

	t.Run("with AppContext", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := app.echo.NewContext(req, rec)
		ac := &AppContext{Context: ctx, app: app}

		err := handler(ac)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("without AppContext", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := app.echo.NewContext(req, rec)

		err := handler(ctx)
		assert.ErrorIs(t, err, ErrContextNotConfigured)
	})
}

func TestRequireRole(t *testing.T) {
	app := newTestApp(t)
	mw := RequireRole(RoleOwner, RoleRoot)

	handler := mw(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	t.Run("allowed role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := app.echo.NewContext(req, rec)
		ctx.Set(UserContextKey, &User{ID: 1, Role: RoleOwner})

		err := handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("denied role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := app.echo.NewContext(req, rec)
		ctx.Set(UserContextKey, &User{ID: 1, Role: RoleUser})

		err := handler(ctx)

		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		assert.Equal(t, http.StatusForbidden, he.Code)
	})

	t.Run("no user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := app.echo.NewContext(req, rec)

		err := handler(ctx)

		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})
}

func TestWithAppContext(t *testing.T) {
	app := newTestApp(t)
	mw := withAppContext(app)

	var capturedCtx echo.Context

	handler := mw(func(ctx echo.Context) error {
		capturedCtx = ctx

		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)

	err := handler(ctx)
	assert.NoError(t, err)

	ac, ok := capturedCtx.(*AppContext)
	require.True(t, ok)
	assert.Equal(t, app, ac.app)
}

func TestWithRequestLogger(t *testing.T) {
	logger := slog.Default()
	mw := withRequestLogger(logger)

	handler := mw(func(ctx echo.Context) error {
		l := ctx.Get(LoggerContextKey)
		assert.NotNil(t, l)

		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	ctx := e.NewContext(req, rec)

	err := handler(ctx)
	assert.NoError(t, err)
}

func TestWithViewData(t *testing.T) {
	app := newTestApp(t)

	handler := withViewData(func(ctx echo.Context) error {
		ac := ctx.(*AppContext)
		data := ac.Data()
		assert.NotNil(t, data)

		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	err := handler(ac)
	assert.NoError(t, err)
}

func TestAuthByCookie_NonAppContext(t *testing.T) {
	e := echo.New()
	handler := AuthByCookie(func(_ echo.Context) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := handler(ctx)
	assert.ErrorIs(t, err, ErrContextNotConfigured)
}

func TestAuthByCookie_Redirect_On401(t *testing.T) {
	app := newTestApp(t)
	// No cookies → authByCookie returns 401 → should redirect to login
	handler := AuthByCookie(func(_ echo.Context) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	err := handler(ac)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
}

func TestAuthByCookie_NonUnauthorizedError(t *testing.T) {
	app := newTestApp(t)

	// Blocked user with valid JWT → authByCookie returns 403 (not 401)
	user := &User{ID: 1, Role: RoleOwner, Status: UserBlocked}
	app.config.UserCase = &mockUseCaseForAuth{user: user}

	tokenStr, _, err := createAccessToken(user, app.config.JWTSecret, app.config.AccessTokenTTL, "", nil)
	require.NoError(t, err)

	handler := AuthByCookie(func(_ echo.Context) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: tokenStr})

	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	returnedErr := handler(ac)
	require.Error(t, returnedErr)

	var he *echo.HTTPError
	require.ErrorAs(t, returnedErr, &he)
	assert.Equal(t, http.StatusForbidden, he.Code)
}

func TestAuthByCookie_Success(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 7, Role: RoleOwner, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{user: user}

	tokenStr, _, err := createAccessToken(user, app.config.JWTSecret, app.config.AccessTokenTTL, "", nil)
	require.NoError(t, err)

	var capturedUser *User
	handler := AuthByCookie(func(ctx echo.Context) error {
		capturedUser, _ = ctx.Get(UserContextKey).(*User)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: tokenStr})

	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	require.NoError(t, handler(ac))
	require.NotNil(t, capturedUser)
	assert.Equal(t, int64(7), capturedUser.ID)
	assert.True(t, capturedUser.Current)
}

func TestWithRequestLogger_WithRequestID(t *testing.T) {
	logger := slog.Default()
	mw := withRequestLogger(logger)

	handler := mw(func(ctx echo.Context) error {
		l := ctx.Get(LoggerContextKey)
		assert.NotNil(t, l)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	rec.Header().Set(echo.HeaderXRequestID, "req-123")

	e := echo.New()
	ctx := e.NewContext(req, rec)

	assert.NoError(t, handler(ctx))
}

func TestWithViewData_WithUser(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 3, Role: RoleUser}

	handler := withViewData(func(ctx echo.Context) error {
		ac := ctx.(*AppContext)
		data := ac.Data()
		assert.NotNil(t, data)
		assert.Equal(t, user, data.User)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}
	ac.Set(UserContextKey, user)

	assert.NoError(t, handler(ac))
}

func TestWithViewData_WithCSRF(t *testing.T) {
	app := newTestApp(t)

	handler := withViewData(func(ctx echo.Context) error {
		ac := ctx.(*AppContext)
		data := ac.Data()
		assert.NotNil(t, data)
		assert.True(t, data.Has("csrf_token"))
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}
	ac.Set("csrf", "csrf-token-value")

	assert.NoError(t, handler(ac))
}

// stubUseCase from config_test.go is reused here.
// Ensure it satisfies the interface — it's defined in config_test.go.

// Minimal usecase for middleware tests that need SearchByLogin.
type mockUseCaseForAuth struct {
	stubUseCase
	user             *User
	searchErr        error
	compareOk        bool
	compareErr       error
	token            *Token
	tokenErr         error
	revokeErr        error
	setLastLoggedErr error
	createTokenErr   error
}

func (m *mockUseCaseForAuth) SearchByLogin(_ context.Context, _ string) (*User, error) {
	return m.user, m.searchErr
}

func (m *mockUseCaseForAuth) SearchByID(_ context.Context, _ int64) (*User, error) {
	return m.user, m.searchErr
}

func (m *mockUseCaseForAuth) ComparePassword(_ *User, _ string) (bool, error) {
	return m.compareOk, m.compareErr
}

func (m *mockUseCaseForAuth) SearchToken(_ context.Context, _ string) (*Token, error) {
	return m.token, m.tokenErr
}

func (m *mockUseCaseForAuth) RevokeUserTokens(_ context.Context, _ int64) error {
	return m.revokeErr
}

func (m *mockUseCaseForAuth) CreateAuthToken(_ context.Context, _ *User, _ string, _ time.Duration) (*Token, error) {
	if m.createTokenErr != nil {
		return nil, m.createTokenErr
	}

	return &Token{DTExpired: time.Now().Add(time.Hour)}, nil
}

func (m *mockUseCaseForAuth) SetLastLogged(_ context.Context, _ *User) error {
	return m.setLastLoggedErr
}
