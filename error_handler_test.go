package goadmin

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRenderer — minimal renderer for HTMLError tests.
type testRenderer struct{}

func (r *testRenderer) Render(w io.Writer, _ string, _ any, _ echo.Context) error {
	_, _ = io.WriteString(w, "rendered")
	return nil
}

func newEchoContextWithRenderer(method, path, accept string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Renderer = &testRenderer{}
	req := httptest.NewRequest(method, path, nil)

	if accept != "" {
		req.Header.Set(echo.HeaderAccept, accept)
	}

	rec := httptest.NewRecorder()

	return e.NewContext(req, rec), rec
}

func newEchoContext(method, path string, accept string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)

	if accept != "" {
		req.Header.Set(echo.HeaderAccept, accept)
	}

	rec := httptest.NewRecorder()

	return e.NewContext(req, rec), rec
}

func TestJSONError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   int
		wantErrMsg string
	}{
		{
			name:       "internal error hides details",
			err:        errors.New("db connection failed"),
			wantCode:   http.StatusInternalServerError,
			wantErrMsg: "Internal server error",
		},
		{
			name:       "http 500 hides details",
			err:        echo.NewHTTPError(http.StatusInternalServerError, "secret"),
			wantCode:   http.StatusInternalServerError,
			wantErrMsg: "Internal server error",
		},
		{
			name:       "http 400 shows message",
			err:        echo.NewHTTPError(http.StatusBadRequest, "bad input"),
			wantCode:   http.StatusBadRequest,
			wantErrMsg: "bad input",
		},
		{
			name:       "http 404",
			err:        echo.NewHTTPError(http.StatusNotFound),
			wantCode:   http.StatusNotFound,
			wantErrMsg: "Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := newEchoContext(http.MethodGet, "/api/test", "application/json")

			JSONError(tt.err, ctx)

			assert.Equal(t, tt.wantCode, rec.Code)

			var resp Response
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.False(t, resp.Success)
			assert.Equal(t, tt.wantErrMsg, resp.Error)
		})
	}
}

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"generic error", errors.New("oops"), http.StatusInternalServerError},
		{"http 403", echo.NewHTTPError(http.StatusForbidden), http.StatusForbidden},
		{"http 401", echo.NewHTTPError(http.StatusUnauthorized), http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := newEchoContext(http.MethodGet, "/test", "")

			HTTPError(tt.err, ctx)

			assert.Equal(t, tt.wantCode, rec.Code)
		})
	}
}

func TestErrorHandler_ContentNegotiation(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		accept   string
		wantCode int
	}{
		{"json accept", "/api/data", "application/json", http.StatusBadRequest},
		{"no accept defaults to http", "/page", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := newEchoContext(http.MethodGet, tt.path, tt.accept)

			errorHandler(echo.NewHTTPError(http.StatusBadRequest, "test"), ctx)

			assert.Equal(t, tt.wantCode, rec.Code)
		})
	}
}

func TestErrorHandler_JSONSuffix(t *testing.T) {
	// ctx.Path() returns the route, not the URL — SetPath is required
	ctx, rec := newEchoContext(http.MethodGet, "/api/data.json", "")
	ctx.SetPath("/api/data.json")

	errorHandler(echo.NewHTTPError(http.StatusBadRequest, "test"), ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogFromCtx(t *testing.T) {
	ctx, _ := newEchoContext(http.MethodGet, "/", "")

	// No logger set — returns default
	logger := logFromCtx(ctx)
	assert.NotNil(t, logger)
}

func TestLogFromCtx_WithLogger(t *testing.T) {
	ctx, _ := newEchoContext(http.MethodGet, "/", "")

	custom := slog.Default().With("test", "value")
	ctx.Set(LoggerContextKey, custom)

	got := logFromCtx(ctx)
	assert.Equal(t, custom, got)
}

func TestErrorHandler_HTMLAccept(t *testing.T) {
	ctx, rec := newEchoContextWithRenderer(http.MethodGet, "/page", "text/html")

	errorHandler(echo.NewHTTPError(http.StatusBadRequest, "bad"), ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestViewData_JetVars(t *testing.T) {
	data := viewData{Code: 404, Title: "Not Found", Error: "not found", Details: "detail"}

	vars := data.JetVars()
	assert.Equal(t, 404, vars["code"].Interface())
	assert.Equal(t, "not found", vars["error"].Interface())
	assert.Equal(t, "Not Found", vars["title"].Interface())
	assert.Equal(t, "detail", vars["details"].Interface())
}

func TestViewData_JetData(t *testing.T) {
	data := viewData{}
	assert.Nil(t, data.JetData())
}

func TestHTMLError_Variants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"generic error", errors.New("internal"), http.StatusInternalServerError},
		{"400 bad request", echo.NewHTTPError(http.StatusBadRequest, "bad input"), http.StatusBadRequest},
		{"404 not found", echo.NewHTTPError(http.StatusNotFound), http.StatusNotFound},
		{"403 forbidden", echo.NewHTTPError(http.StatusForbidden), http.StatusForbidden},
		{"401 unauthorized", echo.NewHTTPError(http.StatusUnauthorized), http.StatusUnauthorized},
		{"500 internal", echo.NewHTTPError(http.StatusInternalServerError), http.StatusInternalServerError},
		{"422 unprocessable", echo.NewHTTPError(http.StatusUnprocessableEntity, "unprocessable"), http.StatusUnprocessableEntity},
		{"503 service unavailable", echo.NewHTTPError(http.StatusServiceUnavailable), http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := newEchoContextWithRenderer(http.MethodGet, "/page", "text/html")
			HTMLError(tt.err, ctx)
			assert.Equal(t, tt.wantCode, rec.Code)
		})
	}
}
