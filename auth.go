package goadmin

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func auth(ctx *AppContext) (result User, err error) {
	login := ctx.FormValue("login")
	password := ctx.FormValue("password")

	result.Login = login
	result.Password = password

	user, err := ctx.UserCase().SearchByLogin(ctx.Ctx(), login)
	if err != nil {
		return result, err
	}

	ok, err := ctx.UserCase().ComparePassword(user, password)
	if err != nil {
		return result, err
	}

	if !ok {
		return result, ErrWrongPassword
	}

	cookieToken, err := generateSecureToken(32)
	if err != nil {
		return result, errors.Wrap(err, "generating secure token failed")
	}

	token, err := ctx.UserCase().CreateAuthToken(ctx.Ctx(), user, cookieToken)
	if err != nil {
		return result, errors.Wrap(err, "creating auth token failed")
	}

	err = ctx.UserCase().SetLastLogged(ctx.Ctx(), user)
	if err != nil {
		return result, errors.Wrap(err, "updating user failed")
	}

	http.SetCookie(ctx.Response(), &http.Cookie{
		Name:     AccessCookieName,
		Value:    cookieToken,
		Expires:  token.DTExpired,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return result, nil
}

func authByCookie(ctx *AppContext) (*User, error) {
	cookie, err := ctx.Cookie(AccessCookieName)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	c := ctx.Request().Context()

	token, err := ctx.UserCase().SearchToken(c, cookie.Value)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	if token.Type != AuthToken {
		return nil, echo.NewHTTPError(http.StatusForbidden)
	}

	if token.IsExpired() {
		return nil, echo.NewHTTPError(http.StatusNotFound)
	}

	err = ctx.UserCase().SetLastLogged(c, token.User)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	token.User.Current = true

	return token.User, nil
}
