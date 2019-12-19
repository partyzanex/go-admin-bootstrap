package main

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/partyzanex/go-admin-bootstrap/repository/postgres"
	"github.com/partyzanex/go-admin-bootstrap/usecase"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("PG_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(time.Second)

	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewTokenRepository(db)
	userCase := usecase.NewUserCase(userRepo, tokenRepo)

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	goadmin.AccessCookieName = "access_token"

	admin, err := goadmin.New(goadmin.Config{
		Host:       "localhost",
		Port:       9900,
		DevMode:    true,
		BaseURL:    "http://localhost:9900/admin",
		ViewsPath:  "./views",
		AssetsPath: "./assets",
		DBConfig: goadmin.DBConfig{
			DB:             db,
			Driver:         "postgres",
			DBName:         "goadmin",
			MigrationsPath: "file:///" + dir + "/../db/migrations/postgres",
		},
		UserCase: userCase,
		Middleware: []echo.MiddlewareFunc{
			middleware.Recover(),
			middleware.Logger(),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	admin.Echo().Logger.Fatal(admin.Serve())
}
