package main

import (
	"database/sql"
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

	admin, err := goadmin.New(goadmin.Config{
		Host:       "localhost",
		Port:       9900,
		BaseURL:    "http://localhost:9000/admin",
		ViewsPath:  "../views",
		DevMode:    true,
		AssetsPath: "../assets",
		DBConfig: goadmin.DBConfig{
			DB:             db,
			DriverName:     "postgres",
			DatabaseName:   "goadmin",
			MigrationsPath: "../db/migrations/postgres",
		},
		UserCase: userCase,
	})
	if err != nil {
		log.Fatal(err)
	}

	admin.Echo().Logger.Fatal(admin.Serve())
}
