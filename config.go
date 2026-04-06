package goadmin

import (
	"database/sql"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type (
	DBConfig struct {
		DB              *sql.DB
		MigrationsTable string
	}

	Config struct {
		Host string
		Port uint16

		BaseURL    string
		ViewsPath  string
		AssetsPath string

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
