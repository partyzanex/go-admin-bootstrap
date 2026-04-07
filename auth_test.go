package goadmin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecureToken(t *testing.T) {
	token, err := generateSecureToken(32)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	token2, err := generateSecureToken(32)
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestClearAuthCookies(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	clearAuthCookies(ac)

	cookies := rec.Result().Cookies()
	assert.Len(t, cookies, 2)

	for _, c := range cookies {
		assert.Equal(t, -1, c.MaxAge)
		assert.Empty(t, c.Value)
	}
}

func TestSetAuthCookies(t *testing.T) {
	app := newTestApp(t)
	app.config.UserCase = &mockUseCaseForAuth{
		user: &User{ID: 1, Role: RoleOwner, Status: UserActive},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	user := &User{ID: 1, Role: RoleOwner}

	err := setAuthCookies(ac, user)
	require.NoError(t, err)

	cookies := rec.Result().Cookies()
	assert.GreaterOrEqual(t, len(cookies), 2)

	var hasAccess, hasRefresh bool

	for _, c := range cookies {
		if c.Name == "test" {
			hasAccess = true
			assert.NotEmpty(t, c.Value)
			assert.True(t, c.HttpOnly)
		}

		if c.Name == "test_refresh" {
			hasRefresh = true
			assert.NotEmpty(t, c.Value)
			assert.True(t, c.HttpOnly)
		}
	}

	assert.True(t, hasAccess, "access cookie must be set")
	assert.True(t, hasRefresh, "refresh cookie must be set")
}

func TestAuthByCookie_ValidJWT(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 42, Role: RoleOwner, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{user: user}

	tokenStr, _, err := createAccessToken(user, app.config.JWTSecret, app.config.AccessTokenTTL, "", nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: tokenStr})

	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	result, authErr := authByCookie(ac)
	require.NoError(t, authErr)
	assert.Equal(t, int64(42), result.ID)
	assert.True(t, result.Current)
}

func TestAuthByCookie_BlockedUserJWT(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleOwner, Status: UserBlocked}
	app.config.UserCase = &mockUseCaseForAuth{user: user}

	tokenStr, _, err := createAccessToken(user, app.config.JWTSecret, app.config.AccessTokenTTL, "", nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: tokenStr})

	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, authErr := authByCookie(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, authErr, &he)
	assert.Equal(t, http.StatusForbidden, he.Code)
}

func TestAuthByCookie_NoCookies(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, authErr := authByCookie(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, authErr, &he)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestAuthByCookie_InvalidJWT_FallsToRefresh(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleOwner, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user: user,
		token: &Token{
			UserID:    1,
			Token:     "refresh-val",
			DTExpired: time.Now().Add(time.Hour),
			User:      user,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: "invalid-jwt"})
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "refresh-val"})

	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	// Invalid JWT (not expired, just bad) → returns 401, doesn't fall to refresh
	_, authErr := authByCookie(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, authErr, &he)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestAuthByCookie_ExpiredJWT_RefreshWorks(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleOwner, Status: UserActive}

	// Create expired JWT
	expiredToken, _, err := createAccessToken(user, app.config.JWTSecret, -time.Hour, "", nil)
	require.NoError(t, err)

	app.config.UserCase = &mockUseCaseForAuth{
		user: user,
		token: &Token{
			UserID:    1,
			Token:     "refresh-val",
			DTExpired: time.Now().Add(time.Hour),
			User:      user,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: expiredToken})
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "refresh-val"})

	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	result, authErr := authByCookie(ac)
	require.NoError(t, authErr)
	assert.Equal(t, int64(1), result.ID)
}

func TestAuth_Success(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 10, Login: "admin@example.com", Role: RoleOwner, Status: UserActive, Password: "hashed"}
	app.config.UserCase = &mockUseCaseForAuth{
		user:      user,
		compareOk: true,
	}

	body := "login=admin%40example.com&password=Admin123"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	result, err := auth(ac)
	require.NoError(t, err)
	assert.Equal(t, int64(10), result.ID)
	assert.Equal(t, "admin@example.com", result.Login)
}

func TestAuth_SearchByLoginError(t *testing.T) {
	app := newTestApp(t)

	app.config.UserCase = &mockUseCaseForAuth{
		searchErr: errors.New("db error"),
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := auth(ac)
	assert.Error(t, err)
}

func TestAuth_BlockedUser(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleUser, Status: UserBlocked}
	app.config.UserCase = &mockUseCaseForAuth{user: user}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := auth(ac)
	assert.ErrorIs(t, err, ErrUserBlocked)
}

func TestAuth_ComparePasswordError(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleUser, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user:       user,
		compareErr: errors.New("compare error"),
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := auth(ac)
	assert.Error(t, err)
}

func TestAuth_WrongPassword(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleUser, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user:      user,
		compareOk: false,
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=x&password=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := auth(ac)
	assert.ErrorIs(t, err, ErrWrongPassword)
}

func TestAuth_RevokeTokensError(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleUser, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user:      user,
		compareOk: true,
		revokeErr: errors.New("revoke error"),
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := auth(ac)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revoking old tokens")
}

func TestAuth_SetLastLoggedError(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleUser, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user:             user,
		compareOk:        true,
		setLastLoggedErr: errors.New("last logged error"),
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := auth(ac)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updating user failed")
}

func TestAuth_SetAuthCookiesError(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleUser, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user:           user,
		compareOk:      true,
		createTokenErr: errors.New("create token error"),
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := auth(ac)
	assert.Error(t, err)
}

func TestAuthByRefreshToken_SetAuthCookiesError(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleOwner, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user: user,
		token: &Token{
			UserID:    1,
			Token:     "refresh-val",
			DTExpired: time.Now().Add(time.Hour),
			User:      user,
		},
		createTokenErr: errors.New("create token error"),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "refresh-val"})
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := authByRefreshToken(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

func TestSetAuthCookies_CreateTokenError(t *testing.T) {
	app := newTestApp(t)
	app.config.UserCase = &mockUseCaseForAuth{
		createTokenErr: errors.New("create token error"),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	err := setAuthCookies(ac, &User{ID: 1, Role: RoleUser})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating refresh token")
}

func TestAuthByRefreshToken_ExpiredToken(t *testing.T) {
	app := newTestApp(t)

	app.config.UserCase = &mockUseCaseForAuth{
		tokenErr: NewTokenExpiredError("tok"),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "some-token"})
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := authByRefreshToken(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestAuthByRefreshToken_SearchTokenError(t *testing.T) {
	app := newTestApp(t)

	app.config.UserCase = &mockUseCaseForAuth{
		tokenErr: errors.New("db failure"),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "some-token"})
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := authByRefreshToken(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

func TestAuthByRefreshToken_RevokeError(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleOwner, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		token: &Token{
			UserID:    1,
			Token:     "refresh-val",
			DTExpired: time.Now().Add(time.Hour),
			User:      user,
		},
		revokeErr: errors.New("revoke error"),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "refresh-val"})
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := authByRefreshToken(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

func TestAuthByRefreshToken_SetLastLoggedError(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleOwner, Status: UserActive}
	app.config.UserCase = &mockUseCaseForAuth{
		user: user,
		token: &Token{
			UserID:    1,
			Token:     "refresh-val",
			DTExpired: time.Now().Add(time.Hour),
			User:      user,
		},
		setLastLoggedErr: errors.New("set last logged error"),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "refresh-val"})
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, err := authByRefreshToken(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

func TestAuthByRefreshToken_BlockedUser(t *testing.T) {
	app := newTestApp(t)

	user := &User{ID: 1, Role: RoleOwner, Status: UserBlocked}
	app.config.UserCase = &mockUseCaseForAuth{
		user: user,
		token: &Token{
			UserID:    1,
			Token:     "refresh-val",
			DTExpired: time.Now().Add(time.Hour),
			User:      user,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test_refresh", Value: "refresh-val"})

	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)
	ac := &AppContext{Context: ctx, app: app}

	_, authErr := authByRefreshToken(ac)

	var he *echo.HTTPError
	require.ErrorAs(t, authErr, &he)
	assert.Equal(t, http.StatusForbidden, he.Code)
}
