package goadmin

import (
	"encoding/hex"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/partyzanex/go-admin-bootstrap/encrypt"
	"github.com/pkg/errors"
)

func auth(ctx *AdminContext) (result User, err error) {
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

	token, err := ctx.UserCase().CreateAuthToken(ctx.Ctx(), user)
	if err != nil {
		return result, errors.Wrap(err, "creating auth token failed")
	}

	err = ctx.UserCase().SetLastLogged(ctx.Ctx(), user)
	if err != nil {
		return result, errors.Wrap(err, "updating user failed")
	}

	key, iv := encrypt.KeysFromString(ctx.RealIP() + ctx.Request().UserAgent())

	enc, err := encrypt.New("aes-256-cbc", key, iv)
	if err != nil {
		return result, errors.Wrap(err, "creating encryption failed")
	}

	tokenValue, err := enc.Encrypt([]byte(token.Token))
	if err != nil {
		return result, errors.Wrap(err, "encryption failed")
	}

	http.SetCookie(ctx.Response(), &http.Cookie{
		Name:    AccessCookieName,
		Value:   hex.EncodeToString(tokenValue),
		Expires: token.DTExpired,
		Path:    "/",
	})

	return result, nil
}

func authByCookie(ctx *AdminContext) (*User, error) {
	t, err := ctx.Cookie(AccessCookieName)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	value, err := hex.DecodeString(t.Value)
	if err != nil {
		return nil, errors.Wrap(err, "decoding cookie value failed")
	}

	key, iv := encrypt.KeysFromString(ctx.RealIP() + ctx.Request().UserAgent())

	enc, err := encrypt.New("aes-256-cbc", key, iv)
	if err != nil {
		return nil, errors.Wrap(err, "creating encryption failed")
	}

	tokenValue, err := enc.Decrypt(value)
	if err != nil {
		return nil, errors.Wrap(err, "decryption failed")
	}

	c := ctx.Request().Context()
	token, err := ctx.UserCase().SearchToken(c, string(tokenValue))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	if token.Type != AuthToken {
		return nil, echo.NewHTTPError(http.StatusForbidden)
	}

	if token.IsExpired() {
		return nil, echo.NewHTTPError(http.StatusNotFound)
	}

	token.User.Current = true

	return token.User, nil
}
