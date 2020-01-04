package goadmin

import "github.com/labstack/echo/v4"

type AdminHandler func(ctx *AdminContext) error

func WrapHandler(handleFunc AdminHandler) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac, ok := ctx.(*AdminContext)
		if !ok {
			return ErrContextNotConfigured
		}

		return handleFunc(ac)
	}
}

func AuthByCookie(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac, ok := ctx.(*AdminContext)
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

func withAdminContext(admin *Admin) echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ac := &AdminContext{
				Context: ctx,
				admin:   admin,
			}

			return handlerFunc(ac)
		}
	}
}

func withViewData(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac := ctx.(*AdminContext)

		data := &Data{}
		user, ok := ac.Get(UserContextKey).(*User)
		if ok {
			data.User = user
		}
		data.Breadcrumbs.Add("Dashboard", ac.URL("/"), -100)

		ac.Set(DataContextKey, data)

		return handlerFunc(ctx)
	}
}
