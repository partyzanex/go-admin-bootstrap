package goadmin

import (
	"context"
	"time"
)

type (
	UserRole string

	UserStatus string

	User struct {
		ID       int64      `json:"id"`
		Login    string     `json:"login" validate:"required,email"`
		Password string     `json:"password"`
		Status   UserStatus `json:"status" validate:"required"`
		Name     string     `json:"name" validate:"required"`
		Role     UserRole   `json:"role" validate:"required"`

		DTCreated    time.Time `json:"dt_created"`
		DTUpdated    time.Time `json:"dt_updated"`
		DTLastLogged time.Time `json:"dt_last_logged"`

		PasswordIsEncoded bool `json:"-"`
		Current           bool `json:"-"`
	}

	TokenType string

	Token struct {
		ID        int64     `json:"-"`
		UserID    int64     `json:"user_id"`
		Token     string    `json:"token"`
		Type      TokenType `json:"type"`
		DTExpired time.Time `json:"dt_expired"`
		DTCreated time.Time `json:"dt_created"`

		User *User `json:"-"`
	}

	UserFilter struct {
		IDs    []int64
		Name   string
		Login  string
		Search string // ILIKE match against both login and name
		Status UserStatus
		Limit  int
		Offset int
	}

	UserRepository interface {
		Search(ctx context.Context, filter *UserFilter) ([]*User, error)
		Count(ctx context.Context, filter *UserFilter) (int64, error)
		Create(ctx context.Context, user *User) (*User, error)
		Update(ctx context.Context, user *User) (*User, error)
		SetLastLogged(ctx context.Context, user *User) error
		Delete(ctx context.Context, user *User) error
		GetUserByID(ctx context.Context, id int64) (*User, error)
	}

	TokenRepository interface {
		Search(ctx context.Context, token string) (*Token, error)
		Create(ctx context.Context, token *Token) (*Token, error)
		DeleteExpired(ctx context.Context) (int64, error)
		DeleteByUserID(ctx context.Context, userID int64) error
	}

	UserUseCase interface {
		Validate(user *User, create bool) error

		SearchByLogin(ctx context.Context, login string) (*User, error)
		SearchByID(ctx context.Context, id int64) (*User, error)
		SetLastLogged(ctx context.Context, user *User) error
		Register(ctx context.Context, user *User) error
		UpdateUser(ctx context.Context, user *User) (*User, error)
		DeleteUser(ctx context.Context, id int64) error
		ListUsers(ctx context.Context, filter *UserFilter) ([]*User, int64, error)

		ComparePassword(user *User, password string) (bool, error)
		EncodePassword(user *User) error

		CreateAuthToken(ctx context.Context, user *User, cookieToken string, ttl time.Duration) (*Token, error)
		SearchToken(ctx context.Context, token string) (*Token, error)
		RevokeUserTokens(ctx context.Context, userID int64) error
		CleanupExpiredTokens(ctx context.Context) (int64, error)
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
