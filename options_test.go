package goadmin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestWithMiddleware(t *testing.T) {
	app := newTestApp(t)
	called := false

	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			called = true

			return next(ctx)
		}
	}

	opt := WithMiddleware(mw)
	err := opt(app)
	assert.NoError(t, err)
	assert.Len(t, app.config.middleware, 1)

	// Verify middleware works
	handler := app.config.middleware[0](func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := app.echo.NewContext(req, rec)

	err = handler(ctx)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWithAssets(t *testing.T) {
	app := newTestApp(t)

	asset := &Asset{Path: "custom.js", SortOrder: 0, Kind: JavaScript}
	opt := WithAssets(asset)

	err := opt(app)
	assert.NoError(t, err)
	assert.Len(t, app.config.assets, 1)
	assert.Equal(t, "custom.js", app.config.assets[0].Path)
}
