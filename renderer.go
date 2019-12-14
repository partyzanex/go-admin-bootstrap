package goadmin

import (
	"github.com/CloudyKit/jet"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type Renderer struct {
	Views *jet.Set
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	if !strings.HasSuffix(name, ".jet") {
		name += ".jet"
	}

	if data == nil {
		return errors.New("data cannot be empty")
	}

	viewData, ok := data.(ViewData)
	if !ok {
		return errors.New("data is not implements TemplateData interface")
	}

	view, err := r.Views.GetTemplate(name)
	if err != nil {
		return errors.Wrap(err, "getting template failed")
	}

	err = view.Execute(w, viewData.JetVars(), viewData.JetData())
	if err != nil {
		return errors.Wrap(err, "executing template failed")
	}

	return nil
}
