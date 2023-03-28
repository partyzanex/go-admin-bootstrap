package goadmin

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/partyzanex/go-admin-bootstrap/assets"
	"github.com/partyzanex/go-admin-bootstrap/widgets"
)

func Login(ctx *AppContext) error {
	_, err := ctx.Cookie(AccessCookieName)
	if err == nil {
		return ctx.Redirect(http.StatusFound, ctx.URL("/"))
	}

	u := ctx.User()
	if u != nil {
		return ctx.Redirect(http.StatusFound, ctx.URL("/"))
	}

	sortOrder := -1
	data := &Data{
		Title: "Login",
	}
	data.Breadcrumbs.Add("Login", ctx.URL(LoginURL), &sortOrder)

	if ctx.Request().Method == http.MethodPost {
		_, err = auth(ctx)
		if err != nil && !errors.Is(err, ErrUserNotFound) && !errors.Is(err, ErrWrongPassword) {
			return err
		}

		return ctx.Redirect(http.StatusFound, ctx.URL(DashboardURL))
	}

	return ctx.Render(http.StatusOK, "auth/login", data)
}

func Logout(ctx *AppContext) error {
	if user := ctx.User(); user != nil {
		ctx.SetCookie(&http.Cookie{
			Name:     AccessCookieName,
			Value:    "",
			Expires:  time.Now().Add(-48 * time.Hour),
			Path:     "/",
			HttpOnly: true,
		})
	}

	return ctx.Redirect(http.StatusFound, ctx.URL(LoginURL))
}

func Dashboard(ctx *AppContext) error {
	user := ctx.User()
	if user == nil {
		return ctx.Redirect(http.StatusFound, ctx.URL(LoginURL))
	}

	return ctx.Render(http.StatusOK, "index/dashboard", ctx.Data())
}

func UserList(ctx *AppContext) error {
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
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL), nil)

	return ctx.Render(http.StatusOK, "user/index", data)
}

func UserCreate(ctx *AppContext) error {
	data := ctx.Data()
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL), nil)
	data.Breadcrumbs.Add("Create User", ctx.URL(UserCreateURL), nil)

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

func UserUpdate(ctx *AppContext) error {
	userID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user, err := ctx.UserCase().SearchByID(ctx.Ctx(), userID)
	if err != nil {
		return err
	}

	if ctx.Request().Method == http.MethodPost {
		err = updateUser(ctx, user)
		if err != nil {
			return err
		}
	}

	data := ctx.Data()
	data.Set("user", user)
	data.Set(
		"formAction",
		strings.Replace(UserUpdateURL, ":id", strconv.FormatInt(user.ID, 10), -1),
	)
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL), nil)
	data.Breadcrumbs.Add(user.Name, ctx.URL(UserCreateURL), nil)

	return ctx.Render(http.StatusOK, "user/form", data)
}

func updateUser(ctx *AppContext, user *User) error {
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

	data := ctx.Data()

	err := ctx.UserCase().Validate(user, false)
	if err == nil {
		repo := ctx.UserCase().UserRepository()

		_, err = repo.Update(ctx.Ctx(), user)
		if err != nil {
			data.Set("error", err.Error())
		} else {
			return ctx.Redirect(http.StatusFound, ctx.URL(UserListURL))
		}
	} else {
		data.Set("error", err.Error())
	}

	return nil
}

func UserDelete(ctx *AppContext) error {
	userID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	user := ctx.User()
	if user.ID == userID {
		return echo.NewHTTPError(http.StatusLocked, "unable to delete your account")
	}

	repo := ctx.UserCase().UserRepository()

	err = repo.Delete(ctx.Ctx(), &User{ID: userID})
	if err != nil {
		return err
	}

	return ctx.Redirect(http.StatusFound, ctx.URL(UserListURL))
}

var (
	loc, _       = time.LoadLocation("GMT")
	lastModified = time.Now().In(loc).Format(time.RFC1123)
)

func Favicon(ctx echo.Context) error {
	icon := ctx.Param("id")
	if icon == "" {
		return ctx.NoContent(http.StatusNotAcceptable)
	}

	b, err := assets.Favicon.ReadFile(filepath.Join("favicon", icon))
	if err != nil {
		return errors.Wrap(err, "cannot read file")
	}

	ctx.Response().Header().Add("Accept-Ranges", "bytes")
	ctx.Response().Header().Add(echo.HeaderLastModified, lastModified)

	return ctx.Blob(http.StatusOK, http.DetectContentType(b), b)
}
