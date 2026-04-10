package goadmin

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NotFoundError ---

func TestNotFoundError_ErrorWithID(t *testing.T) {
	err := &NotFoundError{Entity: "user", ID: 42}
	assert.Equal(t, "user not found: 42", err.Error())
}

func TestNotFoundError_ErrorWithStringID(t *testing.T) {
	err := &NotFoundError{Entity: "token", ID: "abc123"}
	assert.Equal(t, "token not found: abc123", err.Error())
}

func TestNotFoundError_ErrorWithoutID(t *testing.T) {
	err := &NotFoundError{Entity: "user", ID: nil}
	assert.Equal(t, "user not found", err.Error())
}

func TestNewUserNotFoundError(t *testing.T) {
	err := NewUserNotFoundError(int64(5))

	assert.Equal(t, "user", err.Entity)
	assert.Equal(t, int64(5), err.ID)
	assert.Equal(t, "user not found: 5", err.Error())
}

func TestNewUserNotFoundError_NilID(t *testing.T) {
	err := NewUserNotFoundError(nil)

	assert.Equal(t, "user", err.Entity)
	assert.Nil(t, err.ID)
	assert.Equal(t, "user not found", err.Error())
}

func TestNewTokenNotFoundError(t *testing.T) {
	err := NewTokenNotFoundError("tok-xyz")

	assert.Equal(t, "token", err.Entity)
	assert.Equal(t, "tok-xyz", err.ID)
	assert.Equal(t, "token not found: tok-xyz", err.Error())
}

func TestIsNotFound_Match(t *testing.T) {
	err := NewUserNotFoundError(1)
	assert.True(t, IsNotFound(err))
}

func TestIsNotFound_NoMatch(t *testing.T) {
	err := errors.New("some other error")
	assert.False(t, IsNotFound(err))
}

func TestIsNotFound_Nil(t *testing.T) {
	assert.False(t, IsNotFound(nil))
}

func TestIsNotFound_Wrapped(t *testing.T) {
	inner := NewUserNotFoundError(1)
	wrapped := fmt.Errorf("operation failed: %w", inner)

	assert.True(t, IsNotFound(wrapped), "should detect NotFoundError through wrapping")
}

func TestIsNotFound_DoubleWrapped(t *testing.T) {
	inner := NewTokenNotFoundError("abc")
	wrapped := fmt.Errorf("layer1: %w", fmt.Errorf("layer2: %w", inner))

	assert.True(t, IsNotFound(wrapped))
}

// --- ExpiredError ---

func TestExpiredError_ErrorWithID(t *testing.T) {
	err := &ExpiredError{Entity: "token", ID: "tok-123"}
	assert.Equal(t, "token expired: tok-123", err.Error())
}

func TestExpiredError_ErrorWithIntID(t *testing.T) {
	err := &ExpiredError{Entity: "session", ID: 99}
	assert.Equal(t, "session expired: 99", err.Error())
}

func TestExpiredError_ErrorWithoutID(t *testing.T) {
	err := &ExpiredError{Entity: "token", ID: nil}
	assert.Equal(t, "token expired", err.Error())
}

func TestNewTokenExpiredError(t *testing.T) {
	err := NewTokenExpiredError("tok-abc")

	assert.Equal(t, "token", err.Entity)
	assert.Equal(t, "tok-abc", err.ID)
	assert.Equal(t, "token expired: tok-abc", err.Error())
}

func TestIsExpired_Match(t *testing.T) {
	err := NewTokenExpiredError("tok-1")
	assert.True(t, IsExpired(err))
}

func TestIsExpired_NoMatch(t *testing.T) {
	err := errors.New("random error")
	assert.False(t, IsExpired(err))
}

func TestIsExpired_Nil(t *testing.T) {
	assert.False(t, IsExpired(nil))
}

func TestIsExpired_Wrapped(t *testing.T) {
	inner := NewTokenExpiredError("tok-2")
	wrapped := fmt.Errorf("check failed: %w", inner)

	assert.True(t, IsExpired(wrapped), "should detect ExpiredError through wrapping")
}

func TestIsExpired_DoubleWrapped(t *testing.T) {
	inner := NewTokenExpiredError("tok-3")
	wrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", inner))

	assert.True(t, IsExpired(wrapped))
}

// --- errors.As / errors.Is cross-type checks ---

func TestNotFoundError_IsNotExpired(t *testing.T) {
	err := NewUserNotFoundError(1)
	assert.False(t, IsExpired(err), "NotFoundError should not match IsExpired")
}

func TestExpiredError_IsNotNotFound(t *testing.T) {
	err := NewTokenExpiredError("tok-1")
	assert.False(t, IsNotFound(err), "ExpiredError should not match IsNotFound")
}

func TestErrorsAs_NotFoundError(t *testing.T) {
	original := NewUserNotFoundError(int64(7))
	wrapped := fmt.Errorf("wrapped: %w", original)

	var nfe *NotFoundError
	require.True(t, errors.As(wrapped, &nfe))
	assert.Equal(t, "user", nfe.Entity)
	assert.Equal(t, int64(7), nfe.ID)
}

func TestErrorsAs_ExpiredError(t *testing.T) {
	original := NewTokenExpiredError("session-abc")
	wrapped := fmt.Errorf("wrapped: %w", original)

	var ee *ExpiredError
	require.True(t, errors.As(wrapped, &ee))
	assert.Equal(t, "token", ee.Entity)
	assert.Equal(t, "session-abc", ee.ID)
}
