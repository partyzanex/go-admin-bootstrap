package goadmin

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidPort          = errors.New("invalid http port")
	ErrRequiredDB           = errors.New("required database connection instance")
	ErrContextNotConfigured = errors.New("admin context not configured")
	ErrRequiredUserName     = errors.New("required user name")
	ErrRequiredUserLogin    = errors.New("required user login")
	ErrInvalidUserLogin     = errors.New("invalid user login")
	ErrInvalidUserStatus    = errors.New("invalid user status")
	ErrRequiredUserID       = errors.New("required user id")
	ErrRequiredUserPassword = errors.New("required user password")
	ErrWrongPassword        = errors.New("wrong password")
	ErrInvalidUserRole      = errors.New("invalid user role")
	ErrRequiredConfig       = errors.New("required config")
	ErrRequiredJWTSecret    = errors.New("required jwt secret")
)

// NotFoundError indicates that a requested entity was not found.
type NotFoundError struct {
	Entity string
	ID     any
}

func (e *NotFoundError) Error() string {
	if e.ID != nil {
		return fmt.Sprintf("%s not found: %v", e.Entity, e.ID)
	}

	return fmt.Sprintf("%s not found", e.Entity)
}

func NewUserNotFoundError(id any) *NotFoundError {
	return &NotFoundError{Entity: "user", ID: id}
}

func NewTokenNotFoundError(token string) *NotFoundError {
	return &NotFoundError{Entity: "token", ID: token}
}

func IsNotFound(err error) bool {
	var nfe *NotFoundError

	return errors.As(err, &nfe)
}

// ExpiredError indicates that a resource has expired.
type ExpiredError struct {
	Entity string
	ID     any
}

func (e *ExpiredError) Error() string {
	if e.ID != nil {
		return fmt.Sprintf("%s expired: %v", e.Entity, e.ID)
	}

	return fmt.Sprintf("%s expired", e.Entity)
}

func NewTokenExpiredError(token string) *ExpiredError {
	return &ExpiredError{Entity: "token", ID: token}
}

func IsExpired(err error) bool {
	var ee *ExpiredError

	return errors.As(err, &ee)
}
