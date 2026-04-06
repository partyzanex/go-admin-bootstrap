package goadmin

import (
	"context"
	"fmt"
	"log/slog"
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

func (c *AppContext) URL(path string, args ...any) string {
	if len(args) > 0 {
		path = fmt.Sprintf(path, args...)
	}

	return Path(c.app.baseURL.Path, path)
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

func (c *AppContext) Log() *slog.Logger {
	if logger, ok := c.Get(LoggerContextKey).(*slog.Logger); ok {
		return logger
	}

	return c.app.logger
}
