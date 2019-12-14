package goadmin

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
)

type AdminContext struct {
	echo.Context

	admin *Admin
}

func (c AdminContext) URL(path string) string {
	return Path(c.admin.baseURL.Path, path)
}

func (c *AdminContext) Data() *Data {
	data, ok := c.Get(DataContextKey).(*Data)
	if ok {
		return data
	}

	return &Data{}
}

func (c *AdminContext) User() *User {
	user, ok := c.Get(UserContextKey).(*User)
	if ok {
		return user
	}

	return nil
}

func (c *AdminContext) Ctx() context.Context {
	return c.Request().Context()
}

func (c *AdminContext) UserCase() UserUseCase {
	return c.admin.UserCase
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

func Path(paths ...string) string {
	for i := range paths {
		if i == 0 {
			continue
		}

		paths[i] = strings.TrimLeft(paths[i], "/")
	}

	return strings.Join(paths, "/")
}
