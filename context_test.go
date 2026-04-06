package goadmin

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func newTestAppContext(t *testing.T) (*AppContext, *echo.Echo) {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	baseURL, _ := url.Parse("http://localhost/admin")
	app := &App{
		config:  &Config{AccessCookieName: "test_cookie"},
		baseURL: baseURL,
		echo:    e,
		logger:  slog.Default(),
	}

	ac := &AppContext{Context: ctx, app: app}

	return ac, e
}

func TestPath(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		expect string
	}{
		{"single", []string{"/admin"}, "/admin"},
		{"two", []string{"/admin", "users"}, "/admin/users"},
		{"strip leading slash", []string{"/admin", "/users"}, "/admin/users"},
		{"empty", []string{""}, ""},
		{"multiple", []string{"/a", "/b", "/c"}, "/a/b/c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, Path(tt.paths...))
		})
	}
}

func TestAppContext_URL(t *testing.T) {
	ac, _ := newTestAppContext(t)

	assert.Equal(t, "/admin/users", ac.URL("/users"))
	assert.Equal(t, "/admin/users/42", ac.URL("/users/%d", 42))
	assert.Equal(t, "/admin/", ac.URL("/"))
}

func TestAppContext_Data(t *testing.T) {
	ac, _ := newTestAppContext(t)

	// No data set — returns empty Data
	data := ac.Data()
	assert.NotNil(t, data)

	// Set data
	expected := &Data{Title: "test"}
	ac.Set(DataContextKey, expected)

	assert.Equal(t, expected, ac.Data())
}

func TestAppContext_User(t *testing.T) {
	ac, _ := newTestAppContext(t)

	// No user — returns nil
	assert.Nil(t, ac.User())

	// Set user
	user := &User{ID: 1, Name: "test"}
	ac.Set(UserContextKey, user)

	assert.Equal(t, user, ac.User())
}

func TestAppContext_CookieName(t *testing.T) {
	ac, _ := newTestAppContext(t)
	assert.Equal(t, "test_cookie", ac.CookieName())
}

func TestAppContext_Log(t *testing.T) {
	ac, _ := newTestAppContext(t)

	// Default — app logger
	assert.Equal(t, slog.Default(), ac.Log())

	// Set request-scoped logger
	custom := slog.Default().With("custom", true)
	ac.Set(LoggerContextKey, custom)

	assert.Equal(t, custom, ac.Log())
}

func TestAppContext_Ctx(t *testing.T) {
	ac, _ := newTestAppContext(t)
	assert.NotNil(t, ac.Ctx())
}
