package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	pgmigrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
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

	sqldb.SetMaxOpenConns(25)                 //nolint:mnd
	sqldb.SetMaxIdleConns(5)                  //nolint:mnd
	sqldb.SetConnMaxLifetime(5 * time.Minute) //nolint:mnd
	sqldb.SetConnMaxIdleTime(1 * time.Minute)

	db := bun.NewDB(sqldb, pgdialect.New())

	defer func() { _ = db.Close() }()

	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewTokenRepository(db)
	userCase := usecase.NewUserCase(userRepo, tokenRepo)

	admin, err := goadmin.New(
		&goadmin.Config{
			Host:             "localhost",
			Port:             9900,
			DevMode:          true,
			BaseURL:          "http://localhost:9900/admin",
			ViewsPath:        "./views",
			AssetsPath:       "./assets",
			AccessCookieName: "access_token",
			JWTSecret:        []byte(os.Getenv("JWT_SECRET")),
			Logger:           logger,
			DBConfig: goadmin.DBConfig{
				DB: db,
				MigrateFunc: func() error {
					return pgmigrations.Up(sqldb, goadmin.DefaultMigrationsTable)
				},
			},
			UserCase: userCase,
		},
		goadmin.WithMiddleware(
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
		),
	)
	if err != nil {
		return fmt.Errorf("creating admin: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	return admin.Serve(ctx)
}
