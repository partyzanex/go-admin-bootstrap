package goadmin

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

// hashRefreshToken returns a hex-encoded HMAC-SHA256 of the raw token value.
// Only the hash is stored in the database; the raw value lives only in the cookie.
func hashRefreshToken(key []byte, token string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(token))

	return hex.EncodeToString(mac.Sum(nil))
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

func auth(ctx *AppContext) (User, error) {
	login := ctx.FormValue("login")
	password := ctx.FormValue("password")

	user, err := ctx.UserCase().SearchByLogin(ctx.Ctx(), login)
	if err != nil {
		return User{}, err
	}

	// Check user status before allowing login
	if user.Status != UserActive {
		return User{}, ErrUserBlocked
	}

	ok, err := ctx.UserCase().ComparePassword(user, password)
	if err != nil {
		return User{}, err
	}

	if !ok {
		return User{}, ErrWrongPassword
	}

	// Revoke any existing refresh tokens before issuing new ones
	if revokeErr := ctx.UserCase().RevokeUserTokens(ctx.Ctx(), user.ID); revokeErr != nil {
		return User{}, fmt.Errorf("revoking old tokens: %w", revokeErr)
	}

	if err := setAuthCookies(ctx, user); err != nil {
		return User{}, err
	}

	if err := ctx.UserCase().SetLastLogged(ctx.Ctx(), user); err != nil {
		return User{}, fmt.Errorf("updating user failed: %w", err)
	}

	return *user, nil
}

func setAuthCookies(ctx *AppContext, user *User) error {
	cfg := ctx.app.config

	// Access token (JWT)
	accessToken, accessExpires, err := createAccessToken(user, cfg.JWTSecret, cfg.AccessTokenTTL)
	if err != nil {
		return fmt.Errorf("creating access token: %w", err)
	}

	http.SetCookie(ctx.Response(), &http.Cookie{ // #nosec G124 -- Secure is false only in DevMode; HttpOnly and SameSite=Strict are always set
		Name:     cfg.AccessCookieName + accessCookieSuffix,
		Value:    accessToken,
		Expires:  accessExpires,
		Path:     "/",
		Secure:   !cfg.DevMode,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Refresh token: raw value goes to the cookie, HMAC-SHA256 hash is stored in DB.
	refreshValue, err := generateSecureToken(SecureTokenLength)
	if err != nil {
		return fmt.Errorf("generating refresh token: %w", err)
	}

	refreshToken, err := ctx.UserCase().CreateAuthToken(
		ctx.Ctx(), user, hashRefreshToken(cfg.JWTSecret, refreshValue), cfg.RefreshTokenTTL,
	)
	if err != nil {
		return fmt.Errorf("creating refresh token: %w", err)
	}

	http.SetCookie(ctx.Response(), &http.Cookie{ // #nosec G124 -- Secure is false only in DevMode; HttpOnly and SameSite=Strict are always set
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

	// No access cookie — skip directly to refresh.
	cookie, cookieErr := ctx.Cookie(cfg.AccessCookieName + accessCookieSuffix)
	if cookieErr != nil {
		return authByRefreshToken(ctx)
	}

	user, jwtErr := authByJWT(ctx, cookie.Value)
	if jwtErr == nil {
		return user, nil
	}

	// Expired token — try refresh.
	if errors.Is(jwtErr, jwt.ErrTokenExpired) {
		return authByRefreshToken(ctx)
	}

	// Forbidden (blocked user) or infrastructure error (5xx) — propagate as-is.
	var he *echo.HTTPError
	if errors.As(jwtErr, &he) && (he.Code == http.StatusForbidden || he.Code >= http.StatusInternalServerError) {
		return nil, jwtErr
	}

	// Invalid JWT or user not found — treat as unauthenticated.
	return nil, echo.NewHTTPError(http.StatusUnauthorized)
}

func authByJWT(ctx *AppContext, tokenStr string) (*User, error) {
	claims, err := parseAccessToken(tokenStr, ctx.app.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	user, err := ctx.UserCase().SearchByID(ctx.Ctx(), claims.UserID)
	if err != nil {
		if IsNotFound(err) {
			return nil, echo.NewHTTPError(http.StatusUnauthorized)
		}

		return nil, echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("looking up user %d: %w", claims.UserID, err))
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

	token, err := ctx.UserCase().SearchToken(ctx.Ctx(), hashRefreshToken(cfg.JWTSecret, cookie.Value))
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
		// #nosec G124 -- Secure is false only in DevMode; HttpOnly and SameSite=Strict are always set
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
