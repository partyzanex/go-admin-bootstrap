package goadmin

import (
	"io"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type Renderer struct {
	Views *jet.Set
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	if !strings.HasSuffix(name, ".jet") {
		name += ".jet"
	}

	if data == nil {
		return errors.New("data cannot be empty")
	}

	v, ok := data.(ViewData)
	if !ok {
		return errors.New("data is not implements TemplateData interface")
	}

	view, err := r.Views.GetTemplate(name)
	if err != nil {
		return errors.Wrap(err, "getting template failed")
	}

	err = view.Execute(w, v.JetVars(), v.JetData())
	if err != nil {
		return errors.Wrap(err, "executing template failed")
	}

	return nil
}
