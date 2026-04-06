package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/repository/postgres"
	"github.com/partyzanex/go-admin-bootstrap/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(os.Getenv("PG_DSN"))))
	db := bun.NewDB(sqldb, pgdialect.New())

	defer func() { _ = db.Close() }()

	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewTokenRepository(db)
	userCase := usecase.NewUserCase(userRepo, tokenRepo)

	admin, err := goadmin.New(&goadmin.Config{
		Host:             "localhost",
		Port:             9900,
		DevMode:          true,
		BaseURL:          "http://localhost:9900/admin",
		ViewsPath:        "./views",
		AssetsPath:       "./assets",
		AccessCookieName: "access_token",
		Logger:           logger,
		DBConfig: goadmin.DBConfig{
			DB: db,
		},
		UserCase: userCase,
		Middleware: []echo.MiddlewareFunc{
			middleware.Recover(),
			middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
				LogStatus: true,
				LogURI:    true,
				LogMethod: true,
				LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
					slog.Info("request",
						"method", v.Method,
						"uri", v.URI,
						"status", v.Status,
					)

					return nil
				},
			}),
		},
	})
	if err != nil {
		return fmt.Errorf("creating admin: %w", err)
	}

	go func() {
		if errServe := admin.Serve(); errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
			slog.Error("shutting down the server", "err", errServe)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	const timeout = 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err = admin.Echo().Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
