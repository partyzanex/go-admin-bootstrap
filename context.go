package goadmin

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

func Path(paths ...string) string {
	for i := range paths {
		if i == 0 {
			continue
		}

		paths[i] = strings.TrimLeft(paths[i], "/")
	}

	return strings.Join(paths, "/")
}

type AppContext struct {
	echo.Context

	app *App
}

func (c *AppContext) URL(path string, args ...interface{}) string {
	result := Path(c.app.baseURL.Path, fmt.Sprintf(path, args...))
	return result
}

func (c *AppContext) Data() *Data {
	data, ok := c.Get(DataContextKey).(*Data)
	if ok {
		return data
	}

	return &Data{}
}

func (c *AppContext) User() *User {
	user, ok := c.Get(UserContextKey).(*User)
	if ok {
		return user
	}

	return nil
}

func (c *AppContext) Ctx() context.Context {
	return c.Request().Context()
}

func (c *AppContext) UserCase() UserUseCase {
	return c.app.config.UserCase
}
