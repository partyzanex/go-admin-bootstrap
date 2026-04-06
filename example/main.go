package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/repository/postgres"
	"github.com/partyzanex/go-admin-bootstrap/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	db, err := sql.Open("postgres", os.Getenv("PG_DSN"))
	if err != nil {
		slog.Error("open sql connection failed", "err", err)
		os.Exit(1)
	}

	db.SetConnMaxLifetime(time.Second)

	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewTokenRepository(db)
	userCase := usecase.NewUserCase(userRepo, tokenRepo)

	goadmin.AccessCookieName = "access_token"

	admin, err := goadmin.New(&goadmin.Config{
		Host:       "localhost",
		Port:       9900,
		DevMode:    true,
		BaseURL:    "http://localhost:9900/admin",
		ViewsPath:  "./views",
		AssetsPath: "./assets",
		DBConfig: goadmin.DBConfig{
			DB:              db,
			MigrationsTable: goadmin.MigrationsTable,
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
		slog.Error("creating admin failed", "err", err)
		os.Exit(1)
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
		slog.Error("shutdown failed", "err", err)
	}
}
