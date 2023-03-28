package goadmin

import (
	"context"
	"time"
)

type (
	UserRole string

	UserStatus string

	User struct {
		ID int64 `db:"id" json:"id"`

		Login    string     `db:"login" json:"login"`
		Password string     `db:"password" json:"password"`
		Status   UserStatus `db:"status" json:"status"`
		Name     string     `db:"name" json:"name"`
		Role     UserRole   `db:"role" json:"role"`

		DTCreated    time.Time `db:"dt_created" json:"dt_created"`
		DTUpdated    time.Time `db:"dt_updated" json:"dt_updated"`
		DTLastLogged time.Time `db:"dt_last_logged" json:"dt_last_logged"`

		PasswordIsEncoded bool `db:"password_is_encoded" json:"-"`
		Current           bool `db:"-" json:"-"`
	}

	TokenType string

	Token struct {
		UserID    int64     `db:"user_id" json:"user_id"`
		Token     string    `db:"token" json:"token"`
		Type      TokenType `db:"type" json:"type"`
		DTExpired time.Time `db:"dt_expired" json:"dt_expired"`
		DTCreated time.Time `db:"dt_created" json:"dt_created"`

		User *User `db:"-"`
	}

	UserFilter struct {
		IDs           []int64
		Name          string
		Login         string
		Status        UserStatus
		Limit, Offset int
	}

	UserRepository interface {
		Search(ctx context.Context, filter *UserFilter) ([]*User, error)
		Count(ctx context.Context, filter *UserFilter) (int64, error)
		Create(ctx context.Context, user *User) (*User, error)
		Update(ctx context.Context, user *User) (*User, error)
		SetLastLogged(ctx context.Context, user *User) error
		Delete(ctx context.Context, user *User) error
	}

	TokenRepository interface {
		Search(ctx context.Context, token string) (*Token, error)
		Create(ctx context.Context, token *Token) (*Token, error)
	}

	UserUseCase interface {
		Validate(user *User, create bool) error

		SearchByLogin(ctx context.Context, login string) (*User, error)
		SearchByID(ctx context.Context, id int64) (*User, error)
		SetLastLogged(ctx context.Context, user *User) error
		Register(ctx context.Context, user *User) error

		ComparePassword(user *User, password string) (bool, error)
		EncodePassword(user *User) error

		CreateAuthToken(ctx context.Context, user *User) (*Token, error)
		SearchToken(ctx context.Context, token string) (*Token, error)

		UserRepository() UserRepository
	}
)

func (user *User) GetDTCreated() string {
	return user.DTCreated.Format(time.RFC3339)
}

func (user *User) GetDTUpdated() string {
	return user.DTUpdated.Format(time.RFC3339)
}

func (user *User) GetDTLastLogged() string {
	return user.DTLastLogged.Format(time.RFC3339)
}

func (role UserRole) IsValid() bool {
	switch role {
	case RoleOwner, RoleRoot, RoleUser:
		return true
	}

	return false
}

func (status UserStatus) IsValid() bool {
	switch status {
	case UserNew, UserActive, UserBlocked:
		return true
	}

	return false
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.DTExpired)
}

func (t TokenType) IsValid() bool {
	return t == AuthToken
}
