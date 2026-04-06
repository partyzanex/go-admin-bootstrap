package goadmin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_GetDTCreated(t *testing.T) {
	u := &User{DTCreated: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	assert.Equal(t, "2026-01-01T00:00:00Z", u.GetDTCreated())
}

func TestUser_GetDTUpdated(t *testing.T) {
	u := &User{DTUpdated: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)}
	assert.Equal(t, "2026-06-15T12:00:00Z", u.GetDTUpdated())
}

func TestUser_GetDTLastLogged(t *testing.T) {
	u := &User{DTLastLogged: time.Date(2026, 3, 1, 8, 30, 0, 0, time.UTC)}
	assert.Equal(t, "2026-03-01T08:30:00Z", u.GetDTLastLogged())
}

func TestUserRole_IsValid(t *testing.T) {
	assert.True(t, RoleOwner.IsValid())
	assert.True(t, RoleRoot.IsValid())
	assert.True(t, RoleUser.IsValid())
	assert.False(t, UserRole("admin").IsValid())
	assert.False(t, UserRole("").IsValid())
}

func TestUserStatus_IsValid(t *testing.T) {
	assert.True(t, UserNew.IsValid())
	assert.True(t, UserActive.IsValid())
	assert.True(t, UserBlocked.IsValid())
	assert.False(t, UserStatus("deleted").IsValid())
	assert.False(t, UserStatus("").IsValid())
}

func TestToken_IsExpired(t *testing.T) {
	expired := &Token{DTExpired: time.Now().Add(-time.Hour)}
	assert.True(t, expired.IsExpired())

	valid := &Token{DTExpired: time.Now().Add(time.Hour)}
	assert.False(t, valid.IsExpired())
}

func TestTokenType_IsValid(t *testing.T) {
	assert.True(t, AuthToken.IsValid())
	assert.False(t, TokenType("refresh").IsValid())
	assert.False(t, TokenType("").IsValid())
}
