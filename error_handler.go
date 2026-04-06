package goadmin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/labstack/echo/v4"
)

func errorHandler(e error, ctx echo.Context) {
	accept := ctx.Request().Header.Get(echo.HeaderAccept)

	switch {
	case strings.HasSuffix(ctx.Path(), "json"):
		JSONError(e, ctx)
	case strings.Contains(accept, "/json"):
		JSONError(e, ctx)
	case strings.Contains(accept, "text/html"):
		HTMLError(e, ctx)
	default:
		HTTPError(e, ctx)
	}
}

type viewData struct {
	Code    int
	Title   string
	Error   string
	Details string
}

func (data viewData) JetVars() jet.VarMap {
	vars := make(jet.VarMap)
	vars.Set("code", data.Code)
	vars.Set("error", data.Error)
	vars.Set("title", data.Title)
	vars.Set("details", data.Details)
	vars.Set("scripts", JS)
	vars.Set("styles", CSS)

	return vars
}

func (viewData) JetData() map[string]any {
	return nil
}

func HTMLError(e error, ctx echo.Context) {
	defer ctx.Logger().Errorf("html error: %s", e)

	code := http.StatusInternalServerError
	title := ""
	userMessage := "Внутренняя ошибка сервера"

	if he, ok := e.(*echo.HTTPError); ok {
		code = he.Code

		switch code {
		case http.StatusBadRequest:
			title = "Bad Request"
			userMessage = fmt.Sprintf("%v", he.Message)
		case http.StatusNotFound:
			title = "Not Found"
			userMessage = "Not Found"
		case http.StatusForbidden:
			title = "Forbidden"
			userMessage = "Forbidden"
		case http.StatusUnauthorized:
			title = "Unauthorized"
			userMessage = "Unauthorized"
		case http.StatusInternalServerError:
			title = "Internal Server Error"
		default:
			if code < http.StatusInternalServerError {
				title = fmt.Sprintf("%v", he.Message)
				userMessage = title
			} else {
				title = "Internal Server Error"
			}
		}
	}

	data := &viewData{
		Code:  code,
		Title: title,
		Error: userMessage,
	}

	err := ctx.Render(code, "errors/error.jet", data)
	if err != nil {
		ctx.Logger().Error(err)
	}
}

func JSONError(e error, ctx echo.Context) {
	defer ctx.Logger().Errorf("json error: %s", e)

	code := http.StatusInternalServerError
	userMessage := "Internal server error"

	if he, ok := e.(*echo.HTTPError); ok {
		code = he.Code
		if code < http.StatusInternalServerError {
			userMessage = fmt.Sprintf("%v", he.Message)
		}
	}

	resp := &Response{
		Success: false,
		Error:   userMessage,
	}

	if err := ctx.JSON(code, resp); err != nil {
		ctx.Logger().Error(err)
	}
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func HTTPError(e error, ctx echo.Context) {
	defer ctx.Logger().Errorf("http error: %s", e)

	code := http.StatusInternalServerError
	if he, ok := e.(*echo.HTTPError); ok {
		code = he.Code
	}

	if err := ctx.NoContent(code); err != nil {
		ctx.Logger().Error(err)
	}
}
