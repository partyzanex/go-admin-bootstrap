package goadmin

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
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

	cookieToken, err := generateSecureToken(SecureTokenLength)
	if err != nil {
		return result, fmt.Errorf("generating secure token failed: %w", err)
	}

	token, err := ctx.UserCase().CreateAuthToken(ctx.Ctx(), user, cookieToken)
	if err != nil {
		return result, fmt.Errorf("creating auth token failed: %w", err)
	}

	err = ctx.UserCase().SetLastLogged(ctx.Ctx(), user)
	if err != nil {
		return result, fmt.Errorf("updating user failed: %w", err)
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
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	c := ctx.Request().Context()

	token, err := ctx.UserCase().SearchToken(c, cookie.Value)
	if errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrTokenNotFound) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	if token.Type != AuthToken {
		return nil, echo.NewHTTPError(http.StatusForbidden)
	}

	err = ctx.UserCase().SetLastLogged(c, token.User)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	token.User.Current = true

	return token.User, nil
}
