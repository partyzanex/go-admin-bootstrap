package goadmin

import "net/http"

func Dashboard(ctx *AdminContext) error {
	user := ctx.User()
	if user == nil {
		return ctx.Redirect(http.StatusFound, ctx.URL(LoginURL))
	}

	return ctx.Render(http.StatusOK, "index/dashboard", ctx.Data())
}
