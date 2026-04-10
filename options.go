package goadmin

import "github.com/labstack/echo/v4"

func WithMiddleware(mw ...echo.MiddlewareFunc) Option {
	return func(app *App) error {
		app.config.middleware = append(app.config.middleware, mw...)

		return nil
	}
}

func WithAssets(a ...*Asset) Option {
	return func(app *App) error {
		app.config.assets = append(app.config.assets, a...)

		return nil
	}
}
