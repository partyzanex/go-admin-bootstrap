package goadmin

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type (
	DBConfig struct {
		DB              *bun.DB
		MigrationsTable string
	}

	Config struct {
		Host string
		Port uint16

		BaseURL          string
		ViewsPath        string
		AssetsPath       string
		AccessCookieName string

		DevMode bool

		DBConfig DBConfig
		UserCase UserUseCase
		Logger   *slog.Logger

		Middleware []echo.MiddlewareFunc
		Assets     []*Asset
	}
)

func (config *Config) Validate() error {
	if config == nil {
		return ErrRequiredConfig
	}

	if config.DBConfig.DB == nil {
		return ErrRequiredDB
	}

	if config.Port == 0 {
		return ErrInvalidPort
	}

	return nil
}

func (config *Config) Clone() *Config {
	clone := *config

	return &clone
}
