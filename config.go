package goadmin

import (
	"database/sql"
	"github.com/golang-migrate/migrate/database"
	"github.com/labstack/echo/v4"
)

type (
	DBConfig struct {
		DB       *sql.DB
		Instance database.Driver

		Driver         string
		DBName         string
		MigrationsPath string
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

		Middleware []echo.MiddlewareFunc
	}
)

func (config *Config) Validate() error {
	if config.DBConfig.DB == nil {
		return ErrRequiredDB
	}

	if config.Port == 0 {
		return ErrInvalidPort
	}

	return nil
}
