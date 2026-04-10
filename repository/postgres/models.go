package postgres

import (
	"time"

	"github.com/uptrace/bun"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

type userModel struct {
	bun.BaseModel `bun:"table:goadmin.user,alias:u"`

	ID           int64              `bun:"id,pk,autoincrement"`
	Login        string             `bun:"login,notnull"`
	Password     string             `bun:"password,notnull"`
	Status       goadmin.UserStatus `bun:"status,notnull"`
	Name         string             `bun:"name,notnull"`
	Role         goadmin.UserRole   `bun:"role,notnull"`
	DTCreated    time.Time          `bun:"dt_created,notnull,default:now()"`
	DTUpdated    time.Time          `bun:"dt_updated,notnull"`
	DTLastLogged time.Time          `bun:"dt_last_logged,notnull"`
}

type tokenModel struct {
	bun.BaseModel `bun:"table:goadmin.auth_token,alias:t"`

	ID        int64             `bun:"id,pk,autoincrement"`
	UserID    int64             `bun:"user_id,notnull"`
	Token     string            `bun:"token,notnull"`
	Type      goadmin.TokenType `bun:"type,notnull"`
	DTExpired time.Time         `bun:"dt_expired,notnull"`
	DTCreated time.Time         `bun:"dt_created,notnull,default:now()"`
}

func userToModel(u *goadmin.User) *userModel {
	return &userModel{
		ID:           u.ID,
		Login:        u.Login,
		Password:     u.Password,
		Status:       u.Status,
		Name:         u.Name,
		Role:         u.Role,
		DTCreated:    u.DTCreated,
		DTUpdated:    u.DTUpdated,
		DTLastLogged: u.DTLastLogged,
	}
}

func modelToUser(m *userModel) *goadmin.User {
	return &goadmin.User{
		ID:                m.ID,
		Login:             m.Login,
		Password:          m.Password,
		Status:            m.Status,
		Name:              m.Name,
		Role:              m.Role,
		DTCreated:         m.DTCreated,
		DTUpdated:         m.DTUpdated,
		DTLastLogged:      m.DTLastLogged,
		PasswordIsEncoded: true,
	}
}

func tokenToModel(t *goadmin.Token) *tokenModel {
	return &tokenModel{
		ID:        t.ID,
		UserID:    t.UserID,
		Token:     t.Token,
		Type:      t.Type,
		DTExpired: t.DTExpired,
		DTCreated: t.DTCreated,
	}
}

func modelToToken(m *tokenModel) *goadmin.Token {
	return &goadmin.Token{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		Type:      m.Type,
		DTExpired: m.DTExpired,
		DTCreated: m.DTCreated,
		User:      &goadmin.User{ID: m.UserID},
	}
}

func mapSlice[T, U any](items []T, fn func(*T) U) []U {
	result := make([]U, len(items))
	for i := range items {
		result[i] = fn(&items[i])
	}

	return result
}
