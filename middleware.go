package goadmin

import (
	"errors"
	"log/slog"
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
			var he *echo.HTTPError
			if errors.As(err, &he) && he.Code == http.StatusUnauthorized {
				return ctx.Redirect(http.StatusFound, ac.URL(LoginURL))
			}

			return err
		}

		u.Current = true
		ctx.Set(UserContextKey, u)

		return withViewData(handlerFunc)(ctx)
	}
}

func RequireRole(roles ...UserRole) echo.MiddlewareFunc {
	allowed := make(map[UserRole]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			user, ok := ctx.Get(UserContextKey).(*User)
			if !ok || user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			if !allowed[user.Role] {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}

			return next(ctx)
		}
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

func withRequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			req := ctx.Request()
			reqLogger := logger.With(
				"method", req.Method,
				"path", req.URL.Path,
				"remote_ip", ctx.RealIP(),
			)

			if reqID := ctx.Response().Header().Get(echo.HeaderXRequestID); reqID != "" {
				reqLogger = reqLogger.With("request_id", reqID)
			}

			ctx.Set(LoggerContextKey, reqLogger)

			return next(ctx)
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
