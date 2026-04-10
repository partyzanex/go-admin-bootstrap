package goadmin

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/labstack/echo/v4"
)

var (
	errDataEmpty      = errors.New("data cannot be empty")
	errDataNotViewDat = errors.New("data does not implement ViewData interface")
)

type Renderer struct {
	Views *jet.Set
}

func (r *Renderer) Render(w io.Writer, name string, data any, _ echo.Context) error {
	if !strings.HasSuffix(name, ".jet") {
		name += ".jet"
	}

	if data == nil {
		return errDataEmpty
	}

	v, ok := data.(ViewData)
	if !ok {
		return errDataNotViewDat
	}

	view, err := r.Views.GetTemplate(name)
	if err != nil {
		return fmt.Errorf("getting template failed: %w", err)
	}

	err = view.Execute(w, v.JetVars(), v.JetData())
	if err != nil {
		return fmt.Errorf("executing template failed: %w", err)
	}

	return nil
}
