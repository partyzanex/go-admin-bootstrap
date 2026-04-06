package goadmin

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type AdminHandler func(ctx *AppContext) error

func WrapHandler(handleFunc AdminHandler) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac, ok := ctx.(*AppContext)
		if !ok {
			return ErrContextNotConfigured
		}

		return handleFunc(ac)
	}
}

func AuthByCookie(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac, ok := ctx.(*AppContext)
		if !ok {
			return ErrContextNotConfigured
		}

		u, err := authByCookie(ac)
		if err != nil {
			if he, ok := err.(*echo.HTTPError); ok && he.Code == http.StatusUnauthorized {
				return ctx.Redirect(http.StatusFound, ac.URL(LoginURL))
			}

			return err
		}

		u.Current = true
		ctx.Set(UserContextKey, u)

		return withViewData(handlerFunc)(ctx)
	}
}

func withAppContext(app *App) echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ac := &AppContext{
				Context: ctx,
				app:     app,
			}

			return handlerFunc(ac)
		}
	}
}

func withViewData(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac := ctx.(*AppContext)

		data := &Data{}

		user, ok := ac.Get(UserContextKey).(*User)
		if ok {
			data.User = user
		}

		// Pass CSRF token to templates if available
		if csrfToken, ok := ac.Get("csrf").(string); ok {
			data.Set("csrf_token", csrfToken)
		}

		sortOrder := -100

		data.Breadcrumbs.Add("Dashboard", ac.URL("/"), &sortOrder)
		ac.Set(DataContextKey, data)

		return handlerFunc(ctx)
	}
}
