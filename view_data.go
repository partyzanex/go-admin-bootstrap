package goadmin

import (
	"github.com/CloudyKit/jet"
	"github.com/labstack/echo/v4"
	"github.com/partyzanex/go-admin-bootstrap/widgets"
)

type ViewData interface {
	JetVars() jet.VarMap
	JetData() map[string]interface{}
}

type Data struct {
	jet.VarMap

	Title       string
	User        *User
	Breadcrumbs widgets.Breadcrumbs
}

func (data *Data) JetVars() jet.VarMap {
	if data.VarMap == nil {
		data.VarMap = make(jet.VarMap)
	}

	data.Set("scripts", js)
	data.Set("styles", css)
	data.Set("title", data.Title)
	return data.VarMap
}

func (data *Data) JetData() map[string]interface{} {
	result := map[string]interface{}{}

	if data.User != nil {
		result["User"] = data.User
	}
	if data.Breadcrumbs != nil {
		result["Breadcrumbs"] = data.Breadcrumbs
	}

	return result
}

func (data *Data) Set(name string, value interface{}) {
	if data.VarMap == nil {
		data.VarMap = make(jet.VarMap)
	}

	data.VarMap.Set(name, value)
}

var (
	DataContextKey = "data_context"
)

func withViewData(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ac := ctx.(*AdminContext)

		data := &Data{}
		user, ok := ac.Get(UserContextKey).(*User)
		if ok {
			data.User = user
		}
		data.Breadcrumbs.Add("Dashboard", ac.URL("/"))

		ac.Set(DataContextKey, data)

		return handlerFunc(ctx)
	}
}
