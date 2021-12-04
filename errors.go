package goadmin

import "github.com/pkg/errors"

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
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidUserRole      = errors.New("invalid user role")
	ErrTokenNotFound        = errors.New("token not found")
	ErrTokenExpired         = errors.New("token expired")
	ErrRequiredConfig       = errors.New("required config")
)
