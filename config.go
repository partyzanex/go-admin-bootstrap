package goadmin

import (
	"context"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// DBPinger is satisfied by *sql.DB, *bun.DB, and any other driver that supports
// context-aware connectivity checks. It is used by the /health endpoint.
type DBPinger interface {
	PingContext(ctx context.Context) error
}

type (
	// DBConfig holds optional database-related settings.
	// DB is used only for the /health endpoint connectivity check; leave nil to skip the check.
	// MigrateFunc, when set, is called once during New() before the server starts.
	DBConfig struct {
		DB          DBPinger
		MigrateFunc func() error
	}

	Config struct {
		Host string
		Port uint16

		BaseURL          string
		ViewsPath        string
		AssetsPath       string
		AccessCookieName string

		JWTSecret       []byte
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration

		DevMode bool

		DBConfig DBConfig
		UserCase UserUseCase
		Logger   *slog.Logger

		middleware []echo.MiddlewareFunc
		assets     []*Asset
	}

	Option func(*App) error
)

func (config *Config) Validate() error {
	if config == nil {
		return ErrRequiredConfig
	}

	if config.Port == 0 {
		return ErrInvalidPort
	}

	if len(config.JWTSecret) == 0 {
		return ErrRequiredJWTSecret
	}

	if !config.DevMode && len(config.JWTSecret) < MinJWTSecretLen {
		return ErrJWTSecretTooShort
	}

	if config.UserCase == nil {
		return ErrRequiredUserCase
	}

	return nil
}

func (config *Config) Clone() *Config {
	clone := *config

	return &clone
}
