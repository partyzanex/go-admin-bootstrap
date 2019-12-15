package goadmin

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/database"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/url"
)

var (
	DashboardURL     = "/"
	LoginURL         = "/login"
	LogoutURL        = "/logout"
	AccessCookieName = "auth_token"

	ErrContextNotConfigured = errors.New("admin context not configured")
)

type DBConfig struct {
	DB         *sql.DB
	DBInstance database.Driver

	DriverName     string
	DatabaseName   string
	MigrationsPath string
}

type Config struct {
	Host       string
	Port       uint16
	Middleware []echo.MiddlewareFunc
	BaseURL    string
	ViewsPath  string
	DevMode    bool
	AssetsPath string
	DBConfig   DBConfig
	UserCase   UserUseCase

	baseURL *url.URL
}

func (config *Config) Validate() error {
	if config.DBConfig.DB == nil {
		return errors.New("required DB")
	}

	uri, err := url.Parse(config.BaseURL)
	if err != nil {
		return errors.Wrap(err, "parse base url failed")
	}

	config.baseURL = uri

	if config.Port == 0 {
		return errors.New("invalid port")
	}

	return nil
}

type UserUseCase interface {
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

type UserFilter struct {
	IDs           []int64
	Name          string
	Login         string
	Status        UserStatus
	Limit, Offset int
}

type UserRepository interface {
	Search(ctx context.Context, filter *UserFilter) ([]*User, error)
	Count(ctx context.Context, filter *UserFilter) (int64, error)
	Create(ctx context.Context, user User) (*User, error)
	Update(ctx context.Context, user User) (*User, error)
	Delete(ctx context.Context, user User) error
}

type TokenRepository interface {
	Search(ctx context.Context, token string) (*Token, error)
	Create(ctx context.Context, token Token) (*Token, error)
}
