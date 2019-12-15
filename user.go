package goadmin

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	UserNew     UserStatus = "new"
	UserActive  UserStatus = "active"
	UserBlocked UserStatus = "blocked"

	RoleOwner UserRole = "owner"
	RoleRoot  UserRole = "root"
	RoleUser  UserRole = "user"

	UserContextKey = "user_context"
)

type UserRole string

func (role UserRole) IsValid() bool {
	switch role {
	case RoleOwner, RoleRoot, RoleUser:
		return true
	}

	return false
}

type UserStatus string

func (status UserStatus) IsValid() bool {
	switch status {
	case UserNew, UserActive, UserBlocked:
		return true
	}

	return false
}

type User struct {
	ID int64 `db:"id" json:"id"`

	Login    string     `db:"login" json:"login"`
	Password string     `db:"password" json:"password"`
	Status   UserStatus `db:"status" json:"status"`
	Name     string     `db:"name" json:"name"`
	Role     UserRole   `db:"role" json:"role"`

	DTCreated    time.Time `db:"dt_created" json:"dt_created"`
	DTUpdated    time.Time `db:"dt_updated" json:"dt_updated"`
	DTLastLogged time.Time `db:"dt_last_logged" json:"dt_last_logged"`

	PasswordIsEncoded bool `db:"password_is_encoded" json:"-"`
	Current           bool `db:"-" json:"-"`
}

var (
	ErrRequiredUserName     = errors.New("required user name")
	ErrRequiredUserLogin    = errors.New("required user login")
	ErrInvalidUserLogin     = errors.New("invalid user login")
	ErrInvalidUserStatus    = errors.New("invalid user status")
	ErrRequiredUserID       = errors.New("required user id")
	ErrRequiredUserPassword = errors.New("required user password")
	ErrWrongPassword        = errors.New("wrong password")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidUserRole      = errors.New("invalid user role")
	ErrTokenNotFound        = errors.New("token not found")
	ErrTokenExpired         = errors.New("token expired")
)

type Token struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"token"`
	Type      TokenType `db:"type" json:"type"`
	DTExpired time.Time `db:"dt_expired" json:"dt_expired"`
	DTCreated time.Time `db:"dt_created" json:"dt_created"`

	User *User `db:"-"`
}

func (t Token) IsExpired() bool {
	return time.Now().After(t.DTExpired)
}

type TokenType string

func (t TokenType) IsValid() bool {
	return t == AuthToken
}

const (
	AuthToken TokenType = "auth"
)

func UserList(ctx *AdminContext) error {
	repo := ctx.UserCase().UserRepository()

	nav := &Pagination{
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
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL))

	return ctx.Render(http.StatusOK, "user/index", data)
}

func UserCreate(ctx *AdminContext) error {
	data := ctx.Data()
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL))
	data.Breadcrumbs.Add("Create User", ctx.URL(UserCreateURL))

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
	data.Breadcrumbs.Add("Users", ctx.URL(UserListURL))
	data.Breadcrumbs.Add(user.Name, ctx.URL(UserCreateURL))

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
