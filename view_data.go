package goadmin

import (
	"github.com/CloudyKit/jet/v6"
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

	data.Set("scripts", JS)
	data.Set("styles", CSS)
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

func (data Data) Has(key string) bool {
	if data.VarMap == nil {
		return false
	}

	v, ok := data.VarMap[key]
	if !ok || v.IsNil() {
		return false
	}

	return true
}
