package goadmin

import "github.com/labstack/echo/v4"

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

		sortOrder := -100

		data.Breadcrumbs.Add("Dashboard", ac.URL("/"), &sortOrder)
		ac.Set(DataContextKey, data)

		return handlerFunc(ctx)
	}
}
