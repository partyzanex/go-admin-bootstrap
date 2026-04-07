package goadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/partyzanex/go-admin-bootstrap/assets"
	"github.com/partyzanex/go-admin-bootstrap/widgets"
)

func Login(ctx *AppContext) error {
	// Check if user has a valid session — redirect to dashboard
	if user, err := authByCookie(ctx); err == nil && user != nil {
		return ctx.Redirect(http.StatusFound, ctx.URL(DashboardURL))
	}

	sortOrder := -1

	data := &Data{
		Title: "Login",
	}

	if csrfToken, ok := ctx.Get("csrf").(string); ok {
		data.Set("csrf_token", csrfToken)
	}

	data.Breadcrumbs.Add("Login", ctx.URL(LoginURL), &sortOrder)

	if ctx.Request().Method == http.MethodPost {
		user, err := auth(ctx)
		if IsNotFound(err) || errors.Is(err, ErrWrongPassword) || errors.Is(err, ErrUserBlocked) {
			data.Set("err", "Неверный логин или пароль")
			return ctx.Render(http.StatusUnauthorized, "auth/login", data)
		}

		if err != nil {
			return err
		}

		ctx.writeAuditLog(AuditLogin, user.ID, nil)

		return ctx.Redirect(http.StatusFound, ctx.URL(DashboardURL))
	}

	return ctx.Render(http.StatusOK, "auth/login", data)
}

func Logout(ctx *AppContext) error {
	cfg := ctx.app.config

	// Try to identify user from JWT access cookie
	var userID int64

	if cookie, err := ctx.Cookie(cfg.AccessCookieName + accessCookieSuffix); err == nil {
		if claims, parseErr := parseAccessToken(cookie.Value, cfg.JWTSecret); parseErr == nil {
			userID = claims.UserID
		}
	}

	// Fallback: look up user via refresh cookie if JWT is expired/missing.
	// The DB stores the HMAC-SHA256 hash, so hash the raw cookie value before searching.
	if userID == 0 {
		if cookie, err := ctx.Cookie(cfg.AccessCookieName + refreshCookieSuffix); err == nil {
			hashed := hashRefreshToken(cfg.JWTSecret, cookie.Value)
			if token, searchErr := ctx.UserCase().SearchToken(ctx.Ctx(), hashed); searchErr == nil {
				userID = token.UserID
			}
		}
	}

	// Revoke all server-side refresh tokens
	if userID > 0 {
		if err := ctx.UserCase().RevokeUserTokens(ctx.Ctx(), userID); err != nil {
			return fmt.Errorf("revoking tokens: %w", err)
		}

		// Fetch the actor for the audit log (best-effort; failures are ignored).
		if actor, fetchErr := ctx.UserCase().SearchByID(ctx.Ctx(), userID); fetchErr == nil {
			ctx.Set(UserContextKey, actor)
		}

		ctx.writeAuditLog(AuditLogout, userID, nil)
	}

	clearAuthCookies(ctx)

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
	searchQuery := strings.TrimSpace(ctx.QueryParam("q"))
	searchStatus := UserStatus(ctx.QueryParam("status"))

	qs := url.Values{}
	if searchQuery != "" {
		qs.Set("q", searchQuery)
	}

	if searchStatus != "" {
		qs.Set("status", string(searchStatus))
	}

	qs.Set("p", "{page}")

	nav := &widgets.Pagination{
		Ctx:         ctx,
		URLTemplate: ctx.URL(UserListURL) + "?" + qs.Encode(),
		PageParam:   "p",
		Limit:       DefaultLimit,
	}

	nav.ParsePage()

	filter := &UserFilter{
		Search: searchQuery,
		Status: searchStatus,
		Limit:  DefaultLimit,
		Offset: nav.Page*DefaultLimit - DefaultLimit,
	}

	users, count, err := ctx.UserCase().ListUsers(ctx.Ctx(), filter)
	if err != nil {
		return err
	}

	nav.Total = count

	data := ctx.Data()
	data.Set("users", users)
	data.Set("count", count)
	data.Set("pagination", nav)
	data.Set("searchQuery", searchQuery)
	data.Set("searchStatus", string(searchStatus))
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
			ctx.writeAuditLog(AuditCreateUser, user.ID, map[string]any{
				"login": user.Login,
				"role":  string(user.Role),
			})

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
		strings.ReplaceAll(UserUpdateURL, ":id", strconv.FormatInt(user.ID, 10)),
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

	_, err := ctx.UserCase().UpdateUser(ctx.Ctx(), user)
	if err != nil {
		data.Set("error", err.Error())
	} else {
		ctx.writeAuditLog(AuditUpdateUser, user.ID, map[string]any{
			"login": user.Login,
			"role":  string(user.Role),
		})

		return ctx.Redirect(http.StatusFound, ctx.URL(UserListURL))
	}

	return nil
}

func UserDelete(ctx *AppContext) error {
	userID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	actor := ctx.User()
	if actor.ID == userID {
		return echo.NewHTTPError(http.StatusLocked, "unable to delete your account")
	}

	// Fetch the target login before deletion for the audit record.
	meta := map[string]any{"user_id": userID}
	if target, fetchErr := ctx.UserCase().SearchByID(ctx.Ctx(), userID); fetchErr == nil {
		meta["login"] = target.Login
	}

	err = ctx.UserCase().DeleteUser(ctx.Ctx(), userID)
	if err != nil {
		return err
	}

	ctx.writeAuditLog(AuditDeleteUser, userID, meta)

	return ctx.Redirect(http.StatusFound, ctx.URL(UserListURL))
}

var lastModified = time.Now().UTC().Format(time.RFC1123)

func Favicon(ctx echo.Context) error {
	icon := ctx.Param("id")
	if icon == "" {
		return ctx.NoContent(http.StatusNotAcceptable)
	}

	b, err := assets.Favicon.ReadFile(filepath.Join("favicon", icon))
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	ctx.Response().Header().Add("Accept-Ranges", "bytes")
	ctx.Response().Header().Add(echo.HeaderLastModified, lastModified)

	return ctx.Blob(http.StatusOK, http.DetectContentType(b), b)
}

func AuditLogList(ctx *AppContext) error {
	repo := ctx.app.config.AuditLog
	if repo == nil {
		data := ctx.Data()
		data.Breadcrumbs.Add("Audit Log", ctx.URL(AuditLogURL), nil)
		data.Set("auditDisabled", true)

		return ctx.Render(http.StatusOK, "audit/index", data)
	}

	actionFilter := AuditAction(ctx.QueryParam("action"))

	qs := url.Values{}
	if actionFilter != "" {
		qs.Set("action", string(actionFilter))
	}

	qs.Set("p", "{page}")

	nav := &widgets.Pagination{
		Ctx:         ctx,
		URLTemplate: ctx.URL(AuditLogURL) + "?" + qs.Encode(),
		PageParam:   "p",
		Limit:       DefaultLimit,
	}

	nav.ParsePage()

	filter := &AuditLogFilter{
		Action: actionFilter,
		Limit:  DefaultLimit,
		Offset: nav.Page*DefaultLimit - DefaultLimit,
	}

	logs, err := repo.Search(ctx.Ctx(), filter)
	if err != nil {
		return err
	}

	count, err := repo.Count(ctx.Ctx(), filter)
	if err != nil {
		return err
	}

	nav.Total = count

	data := ctx.Data()
	data.Set("logs", logs)
	data.Set("count", count)
	data.Set("pagination", nav)
	data.Set("actionFilter", string(actionFilter))
	data.Breadcrumbs.Add("Audit Log", ctx.URL(AuditLogURL), nil)

	return ctx.Render(http.StatusOK, "audit/index", data)
}

func methodNotAllowed(ctx echo.Context) error {
	return echo.ErrMethodNotAllowed
}
