package goadmin

import (
	"github.com/pkg/errors"
	"time"
)

const (
	UserNew     UserStatus = "new"
	UserActive  UserStatus = "active"
	UserBlocked UserStatus = "blocked"

	UserContextKey = "user_context"
)

type UserStatus string

func (status UserStatus) IsValid() bool {
	switch status {
	case UserNew, UserActive, UserBlocked:
		return true
	}

	return false
}

type User struct {
	ID int64 `db:"id" json:"id"`

	Login    string     `db:"login" json:"login"`
	Password string     `db:"password" json:"password"`
	Status   UserStatus `db:"status" json:"status"`

	Name string `db:"name" json:"name"`

	DTCreated    time.Time `db:"dt_created" json:"dt_created"`
	DTUpdated    time.Time `db:"dt_updated" json:"dt_updated"`
	DTLastLogged time.Time `db:"dt_last_logged" json:"dt_last_logged"`

	PasswordIsEncoded bool `db:"password_is_encoded" json:"-"`
	Current           bool `db:"-" json:"-"`
}

var (
	ErrRequiredUserName     = errors.New("required user name")
	ErrRequiredUserLogin    = errors.New("required user login")
	ErrInvalidUserLogin     = errors.New("invalid user login")
	ErrInvalidUserStatus    = errors.New("invalid user status")
	ErrRequiredUserID       = errors.New("required user id")
	ErrRequiredUserPassword = errors.New("required user password")
	ErrWrongPassword        = errors.New("wrong password")
	ErrUserNotFound         = errors.New("user not found")
	ErrTokenNotFound        = errors.New("token not found")
	ErrTokenExpired         = errors.New("token expired")
)

type Token struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"token"`
	Type      TokenType `db:"type" json:"type"`
	DTExpired time.Time `db:"dt_expired" json:"dt_expired"`
	DTCreated time.Time `db:"dt_created" json:"dt_created"`

	User *User `db:"-"`
}

func (t Token) IsExpired() bool {
	return time.Now().After(t.DTExpired)
}

type TokenType string

func (t TokenType) IsValid() bool {
	return t == AuthToken
}

const (
	AuthToken TokenType = "auth"
)
