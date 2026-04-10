package goadmin

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAccessToken_Valid(t *testing.T) {
	user := &User{
		ID:   42,
		Role: RoleRoot,
	}
	secret := []byte("test-secret-key")
	ttl := 15 * time.Minute

	tokenStr, expiresAt, err := createAccessToken(user, secret, ttl)

	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
	assert.WithinDuration(t, time.Now().Add(ttl), expiresAt, 2*time.Second)
}

func TestCreateAccessToken_ClaimsCorrect(t *testing.T) {
	user := &User{
		ID:   99,
		Role: RoleOwner,
	}
	secret := []byte("my-secret")
	ttl := 30 * time.Minute

	tokenStr, _, err := createAccessToken(user, secret, ttl)
	require.NoError(t, err)

	// Parse back and verify claims
	claims, err := parseAccessToken(tokenStr, secret)
	require.NoError(t, err)

	assert.Equal(t, int64(99), claims.UserID)
	assert.Equal(t, string(RoleOwner), claims.Role)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, 2*time.Second)
}

func TestParseAccessToken_Valid(t *testing.T) {
	user := &User{
		ID:   10,
		Role: RoleUser,
	}
	secret := []byte("secret123")

	tokenStr, _, err := createAccessToken(user, secret, time.Hour)
	require.NoError(t, err)

	claims, err := parseAccessToken(tokenStr, secret)
	require.NoError(t, err)

	assert.Equal(t, int64(10), claims.UserID)
	assert.Equal(t, string(RoleUser), claims.Role)
}

func TestParseAccessToken_Expired(t *testing.T) {
	user := &User{
		ID:   1,
		Role: RoleUser,
	}
	secret := []byte("secret")

	// Create a token that's already expired by using a negative TTL
	// We need to manually construct an expired token
	expiresAt := time.Now().Add(-1 * time.Hour)
	claimsIn := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		UserID: user.ID,
		Role:   string(user.Role),
	}

	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsIn).SignedString(secret)
	require.NoError(t, err)

	_, err = parseAccessToken(tokenStr, secret)
	assert.Error(t, err, "expired token should fail parsing")
	assert.Contains(t, err.Error(), "parsing jwt")
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	user := &User{
		ID:   1,
		Role: RoleUser,
	}
	correctSecret := []byte("correct-secret")
	wrongSecret := []byte("wrong-secret")

	tokenStr, _, err := createAccessToken(user, correctSecret, time.Hour)
	require.NoError(t, err)

	_, err = parseAccessToken(tokenStr, wrongSecret)
	assert.Error(t, err, "wrong secret should fail parsing")
}

func TestParseAccessToken_Malformed(t *testing.T) {
	secret := []byte("secret")

	_, err := parseAccessToken("not.a.valid.token", secret)
	assert.Error(t, err, "malformed token should fail parsing")

	_, err = parseAccessToken("", secret)
	assert.Error(t, err, "empty token should fail parsing")

	_, err = parseAccessToken("garbage", secret)
	assert.Error(t, err, "garbage token should fail parsing")
}

func TestCreateAndParseAccessToken_Roundtrip(t *testing.T) {
	users := []*User{
		{ID: 1, Role: RoleOwner},
		{ID: 100, Role: RoleRoot},
		{ID: 999, Role: RoleUser},
	}
	secret := []byte("roundtrip-secret")

	for _, user := range users {
		tokenStr, _, err := createAccessToken(user, secret, time.Hour)
		require.NoError(t, err)

		claims, err := parseAccessToken(tokenStr, secret)
		require.NoError(t, err)

		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, string(user.Role), claims.Role)
	}
}
