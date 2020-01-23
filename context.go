package goadmin

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

type AdminContext struct {
	echo.Context

	admin *Admin
}

func (c AdminContext) URL(path string, args ...interface{}) string {
	result := Path(c.admin.baseURL.Path, fmt.Sprintf(path, args...))
	return result
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

func Path(paths ...string) string {
	for i := range paths {
		if i == 0 {
			continue
		}

		paths[i] = strings.TrimLeft(paths[i], "/")
	}

	return strings.Join(paths, "/")
}
