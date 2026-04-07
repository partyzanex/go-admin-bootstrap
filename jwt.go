package goadmin

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultJWTIssuer = "go-admin-bootstrap"

var errUnexpectedClaims = errors.New("unexpected claims type")

type Claims struct {
	jwt.RegisteredClaims
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
}

func createAccessToken(user *User, secret []byte, ttl time.Duration, issuer string, audience jwt.ClaimStrings) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)

	if issuer == "" {
		issuer = defaultJWTIssuer
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Audience:  audience,
		},
		UserID: user.ID,
		Role:   string(user.Role),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing jwt: %w", err)
	}

	return token, expiresAt, nil
}

func parseAccessToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(_ *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing jwt: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errUnexpectedClaims
	}

	return claims, nil
}
