package goadmin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/partyzanex/go-admin-bootstrap/widgets"
)

func Login(ctx *AdminContext) error {
	_, err := ctx.Cookie(AccessCookieName)
	if err == nil {
		return ctx.Redirect(http.StatusFound, ctx.URL("/"))
	}

	u := ctx.User()
	if u != nil {
		return ctx.Redirect(http.StatusFound, ctx.URL("/"))
	}

	data := &Data{
		Title: "Login",
	}
	data.Breadcrumbs.Add("Login", ctx.URL(LoginURL), -1)

	if ctx.Request().Method == http.MethodPost {
		_, err := auth(ctx)
		if err != nil && err != ErrUserNotFound && err != ErrWrongPassword {
			return err
		}

		return ctx.Redirect(http.StatusFound, ctx.URL(DashboardURL))
	}

	return ctx.Render(http.StatusOK, "auth/login", data)
}

func Logout(ctx *AdminContext) error {
	if user := ctx.User(); user != nil {
		ctx.SetCookie(&http.Cookie{
			Name:    AccessCookieName,
			Expires: time.Now().Add(-24 * time.Hour),
			Domain:  ctx.Request().Host,
			Path:    "/",
		})
	}

	return ctx.Redirect(http.StatusFound, ctx.URL(LoginURL))
}

func Dashboard(ctx *AdminContext) error {
	user := ctx.User()
	if user == nil {
		return ctx.Redirect(http.StatusFound, ctx.URL(LoginURL))
	}

	return ctx.Render(http.StatusOK, "index/dashboard", ctx.Data())
}

func UserList(ctx *AdminContext) error {
	repo := ctx.UserCase().UserRepository()

	nav := &widgets.Pagination{
		Ctx:         ctx,
		URLTemplate: ctx.URL("/users?p={page}"),
		PageParam:   "p",
		Limit:       DefaultLimit,
	}

	nav.ParsePage()

	filter := &UserFilter{
		Limit:  DefaultLimit,
		Offset: nav.Page*DefaultLimit - DefaultLimit,
	}

	count, err := repo.Count(ctx.Ctx(), filter)
	if err != nil {
		return err
	}

	users, err := repo.Search(ctx.Ctx(), filter)
	if err != nil {
		return err
	}

	nav.Total = count

	data := ctx.Data()
	data.Set("users", users)
	data.Set("count", count)
	data.Set("pagination", nav)
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL), 0)

	return ctx.Render(http.StatusOK, "user/index", data)
}

func UserCreate(ctx *AdminContext) error {
	data := ctx.Data()
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL), 1)
	data.Breadcrumbs.Add("Create User", ctx.URL(UserCreateURL), 2)

	user := &User{}

	if ctx.Request().Method == http.MethodPost {
		user.Login = ctx.FormValue("login")
		user.Name = ctx.FormValue("name")
		user.Password = ctx.FormValue("password")
		user.Role = UserRole(ctx.FormValue("role"))
		user.Status = UserStatus(ctx.FormValue("status"))

		err := ctx.UserCase().Register(ctx.Ctx(), user)
		if err != nil {
			data.Set("error", err.Error())
		} else {
			return ctx.Redirect(http.StatusFound, ctx.URL(UserListURL))
		}
	}

	data.Set("user", user)
	data.Set("formAction", UserCreateURL)

	return ctx.Render(http.StatusOK, "user/form", data)
}

func UserUpdate(ctx *AdminContext) error {
	userID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user, err := ctx.UserCase().SearchByID(ctx.Ctx(), userID)
	if err != nil {
		return err
	}

	data := ctx.Data()

	if ctx.Request().Method == http.MethodPost {
		user.Login = ctx.FormValue("login")
		user.Name = ctx.FormValue("name")
		user.Role = UserRole(ctx.FormValue("role"))
		user.Status = UserStatus(ctx.FormValue("status"))

		if password := ctx.FormValue("password"); password != "" {
			user.Password = password
			user.PasswordIsEncoded = false

			err := ctx.UserCase().EncodePassword(user)
			if err != nil {
				return err
			}
		}

		err := ctx.UserCase().Validate(user, false)
		if err == nil {
			repo := ctx.UserCase().UserRepository()

			_, err := repo.Update(ctx.Ctx(), *user)
			if err != nil {
				data.Set("error", err.Error())
			} else {
				return ctx.Redirect(http.StatusFound, ctx.URL(UserListURL))
			}
		} else {
			data.Set("error", err.Error())
		}
	}

	data.Set("user", user)
	data.Set("formAction", strings.Replace(UserUpdateURL, ":id", strconv.FormatInt(user.ID, 10), -1))
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL), 1)
	data.Breadcrumbs.Add(user.Name, ctx.URL(UserCreateURL), 2)

	return ctx.Render(http.StatusOK, "user/form", data)
}

func UserDelete(ctx *AdminContext) error {
	userID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user := ctx.User()
	if user.ID == userID {
		return echo.NewHTTPError(http.StatusLocked, "unable to delete your account")
	}

	repo := ctx.UserCase().UserRepository()

	err = repo.Delete(ctx.Ctx(), User{ID: userID})
	if err != nil {
		return err
	}

	return ctx.Redirect(http.StatusFound, ctx.URL(UserListURL))
}
