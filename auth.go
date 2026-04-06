package goadmin

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const (
	accessCookieSuffix  = ""
	refreshCookieSuffix = "_refresh"
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

	// Check user status before allowing login
	if user.Status != UserActive {
		return result, ErrUserBlocked
	}

	ok, err := ctx.UserCase().ComparePassword(user, password)
	if err != nil {
		return result, err
	}

	if !ok {
		return result, ErrWrongPassword
	}

	err = setAuthCookies(ctx, user)
	if err != nil {
		return result, err
	}

	err = ctx.UserCase().SetLastLogged(ctx.Ctx(), user)
	if err != nil {
		return result, fmt.Errorf("updating user failed: %w", err)
	}

	return result, nil
}

func setAuthCookies(ctx *AppContext, user *User) error {
	cfg := ctx.app.config

	// Access token (JWT)
	accessToken, accessExpires, err := createAccessToken(user, cfg.JWTSecret, cfg.AccessTokenTTL)
	if err != nil {
		return fmt.Errorf("creating access token: %w", err)
	}

	http.SetCookie(ctx.Response(), &http.Cookie{
		Name:     cfg.AccessCookieName + accessCookieSuffix,
		Value:    accessToken,
		Expires:  accessExpires,
		Path:     "/",
		Secure:   !cfg.DevMode,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Refresh token (random, stored in DB)
	refreshValue, err := generateSecureToken(SecureTokenLength)
	if err != nil {
		return fmt.Errorf("generating refresh token: %w", err)
	}

	refreshToken, err := ctx.UserCase().CreateAuthToken(ctx.Ctx(), user, refreshValue, cfg.RefreshTokenTTL)
	if err != nil {
		return fmt.Errorf("creating refresh token: %w", err)
	}

	http.SetCookie(ctx.Response(), &http.Cookie{
		Name:     cfg.AccessCookieName + refreshCookieSuffix,
		Value:    refreshValue,
		Expires:  refreshToken.DTExpired,
		Path:     "/",
		Secure:   !cfg.DevMode,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

func authByCookie(ctx *AppContext) (*User, error) {
	cfg := ctx.app.config

	// Try JWT access token first (no DB hit)
	cookie, err := ctx.Cookie(cfg.AccessCookieName + accessCookieSuffix)
	if err == nil {
		user, jwtErr := authByJWT(ctx, cookie.Value)
		if jwtErr == nil {
			return user, nil
		}

		// JWT expired or invalid — fall through to refresh
		if !errors.Is(jwtErr, jwt.ErrTokenExpired) {
			return nil, echo.NewHTTPError(http.StatusUnauthorized)
		}
	}

	// Try refresh token (DB hit)
	return authByRefreshToken(ctx)
}

func authByJWT(ctx *AppContext, tokenStr string) (*User, error) {
	claims, err := parseAccessToken(tokenStr, ctx.app.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	user, err := ctx.UserCase().SearchByID(ctx.Ctx(), claims.UserID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	if user.Status != UserActive {
		return nil, echo.NewHTTPError(http.StatusForbidden).SetInternal(ErrUserBlocked)
	}

	user.Current = true

	return user, nil
}

func authByRefreshToken(ctx *AppContext) (*User, error) {
	cfg := ctx.app.config

	cookie, err := ctx.Cookie(cfg.AccessCookieName + refreshCookieSuffix)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	token, err := ctx.UserCase().SearchToken(ctx.Ctx(), cookie.Value)
	if IsExpired(err) || IsNotFound(err) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	// Check user status before allowing refresh
	if token.User.Status != UserActive {
		return nil, echo.NewHTTPError(http.StatusForbidden).SetInternal(ErrUserBlocked)
	}

	// Revoke old refresh tokens before issuing new ones
	if delErr := ctx.UserCase().RevokeUserTokens(ctx.Ctx(), token.User.ID); delErr != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(delErr)
	}

	// Refresh succeeded — issue new access + refresh tokens
	err = setAuthCookies(ctx, token.User)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	// Update last logged only on refresh (not every request)
	err = ctx.UserCase().SetLastLogged(ctx.Ctx(), token.User)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	token.User.Current = true

	return token.User, nil
}

func clearAuthCookies(ctx *AppContext) {
	cfg := ctx.app.config

	for _, suffix := range []string{accessCookieSuffix, refreshCookieSuffix} {
		http.SetCookie(ctx.Response(), &http.Cookie{
			Name:     cfg.AccessCookieName + suffix,
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			Secure:   !cfg.DevMode,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}
}
