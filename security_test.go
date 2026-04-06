package goadmin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRoutedTestApp creates an App with fully configured routes and middleware.
func newRoutedTestApp(t *testing.T) *App {
	t.Helper()

	app := newTestApp(t)
	app.setDefaultMiddleware()
	app.setStaticGroup()
	app.setDefaultRoutes()

	return app
}

// csrfTokenFromGET performs a GET request to the given path and returns the _csrf cookie value.
// The CSRF middleware sets the cookie before calling the handler, so the cookie is present
// even if the handler returned an error (e.g., missing templates).
func csrfTokenFromGET(t *testing.T, app *App, path string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	for _, c := range rec.Result().Cookies() {
		if c.Name == "_csrf" {
			return c.Value
		}
	}

	t.Fatal("_csrf cookie not found in GET response")

	return ""
}

// --- Logout ---

func TestLogout_GET_Returns405(t *testing.T) {
	app := newRoutedTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/logout", nil)
	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
		"GET /logout должен вернуть 405, а не перенаправлять на login или выполнять logout")
}

func TestLogout_POST_NoCSRF_Returns4xx(t *testing.T) {
	app := newRoutedTestApp(t)

	// POST without CSRF cookie and without CSRF header.
	// Echo v4.15+ returns 400 (missing token), not 403 (invalid token).
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"POST /logout без CSRF-токена должен вернуть 400")
}

func TestLogout_POST_WithCSRF_Redirects(t *testing.T) {
	app := newRoutedTestApp(t)

	// Step 1: GET /admin/login → obtain the _csrf cookie
	csrfToken := csrfTokenFromGET(t, app, "/admin/login")
	require.NotEmpty(t, csrfToken)

	// Step 2: POST /admin/logout with a valid CSRF token
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: csrfToken})

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	// Logout does not require authentication — should execute and redirect to /login
	assert.Equal(t, http.StatusFound, rec.Code,
		"POST /logout с валидным CSRF должен редиректить на login")
	assert.Contains(t, rec.Header().Get("Location"), LoginURL)
}

// --- UserDelete ---

func TestUserDelete_GET_Returns405(t *testing.T) {
	app := newRoutedTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/1/delete", nil)
	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
		"GET /users/:id/delete должен вернуть 405")
}

func TestUserDelete_POST_NoCSRF_Returns4xx(t *testing.T) {
	app := newRoutedTestApp(t)

	// POST without a CSRF token. Echo v4.15+ returns 400 (missing token).
	req := httptest.NewRequest(http.MethodPost, "/admin/users/1/delete", nil)
	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"POST /users/:id/delete без CSRF-токена должен вернуть 400")
}

func TestUserDelete_POST_WithCSRF_NoAuth_RedirectsToLogin(t *testing.T) {
	app := newRoutedTestApp(t)

	csrfToken := csrfTokenFromGET(t, app, "/admin/login")
	require.NotEmpty(t, csrfToken)

	// POST with CSRF, but without an authentication cookie
	req := httptest.NewRequest(http.MethodPost, "/admin/users/2/delete", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: csrfToken})

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	// AuthByCookie should redirect to login
	assert.Equal(t, http.StatusFound, rec.Code,
		"без аутентификации должен быть редирект на login")
	assert.Contains(t, rec.Header().Get("Location"), LoginURL)
}

func TestUserDelete_POST_WithCSRF_Authenticated_Deletes(t *testing.T) {
	owner := &User{ID: 1, Role: RoleOwner, Status: UserActive}

	app := newRoutedTestApp(t)
	app.config.UserCase = &mockUseCaseForAuth{user: owner}

	// Step 1: JWT for the authenticated owner
	tokenStr, _, err := createAccessToken(owner, app.config.JWTSecret, app.config.AccessTokenTTL)
	require.NoError(t, err)

	// Step 2: CSRF token
	csrfToken := csrfTokenFromGET(t, app, "/admin/login")
	require.NotEmpty(t, csrfToken)

	// Step 3: POST /admin/users/2/delete — deleting another user (not self)
	req := httptest.NewRequest(http.MethodPost, "/admin/users/2/delete", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: csrfToken})
	req.AddCookie(&http.Cookie{Name: "test", Value: tokenStr}) // JWT cookie

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	// After deletion — redirect to the user list
	assert.Equal(t, http.StatusFound, rec.Code,
		"успешное удаление должно редиректить на список пользователей")
	assert.Contains(t, rec.Header().Get("Location"), UserListURL)
}

func TestUserDelete_POST_SelfDelete_Returns423(t *testing.T) {
	owner := &User{ID: 1, Role: RoleOwner, Status: UserActive}

	app := newRoutedTestApp(t)
	app.config.UserCase = &mockUseCaseForAuth{user: owner}

	tokenStr, _, err := createAccessToken(owner, app.config.JWTSecret, app.config.AccessTokenTTL)
	require.NoError(t, err)

	csrfToken := csrfTokenFromGET(t, app, "/admin/login")
	require.NotEmpty(t, csrfToken)

	// Attempt to delete yourself (ID=1 → /users/1/delete)
	req := httptest.NewRequest(http.MethodPost, "/admin/users/1/delete", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: csrfToken})
	req.AddCookie(&http.Cookie{Name: "test", Value: tokenStr})

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusLocked, rec.Code,
		"самоудаление должно вернуть 423 Locked")
}

// TestCSRFEnforcement_ReplayAttack verifies that a token from one session is not accepted
// without the corresponding _csrf cookie (simulating a CSRF attack from another domain).
func TestCSRFEnforcement_ReplayAttack(t *testing.T) {
	app := newRoutedTestApp(t)

	// The attacker knows the token (e.g., stolen from JS) but cannot send the cookie
	// due to SameSite=Strict. Simulating: header is present, cookie is not.
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	req.Header.Set("X-CSRF-Token", "stolen-token-without-cookie")

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code,
		"токен без соответствующей cookie должен быть отклонён")
}

// TestCSRFEnforcement_MismatchedTokens verifies that mismatched cookie and header tokens are rejected.
func TestCSRFEnforcement_MismatchedTokens(t *testing.T) {
	app := newRoutedTestApp(t)

	csrfToken := csrfTokenFromGET(t, app, "/admin/login")
	require.NotEmpty(t, csrfToken)

	req := httptest.NewRequest(http.MethodPost, "/admin/logout", strings.NewReader("_csrf=wrong-token"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: csrfToken}) // cookie is correct
	// but form _csrf = wrong-token (does not match)

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code,
		"несовпадающий CSRF-токен должен вернуть 403")
}

// TestLogout_DoesNotExecute_OnGETEvenWithCSRFCookie verifies that a GET request does not execute
// logout even when it carries a CSRF cookie (protection against CSRF via <img src>).
func TestLogout_DoesNotExecute_OnGETEvenWithCSRFCookie(t *testing.T) {
	owner := &User{ID: 1, Role: RoleOwner, Status: UserActive}

	app := newRoutedTestApp(t)
	app.config.UserCase = &mockUseCaseForAuth{user: owner}

	tokenStr, _, err := createAccessToken(owner, app.config.JWTSecret, 15*time.Minute)
	require.NoError(t, err)

	// GET with JWT and CSRF cookie — simulating <img src="/admin/logout">
	req := httptest.NewRequest(http.MethodGet, "/admin/logout", nil)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: "any-csrf-value"})
	req.AddCookie(&http.Cookie{Name: "test", Value: tokenStr})

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	// Should return 405, not execute logout
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
		"GET /logout не должен выполняться даже с авторизацией")
}
