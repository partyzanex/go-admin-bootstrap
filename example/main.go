package main

import (
	"database/sql"
	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "dbname=goadmin user=postgres password=535353 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(time.Second)

	admin, err := goadmin.New(goadmin.Config{
		Host:       "localhost",
		Port:       9900,
		BaseURL:    "http://localhost:9000",
		ViewsPath:  "../views",
		DevMode:    true,
		AssetsPath: "../assets",
		DBConfig: goadmin.DBConfig{
			DB:             db,
			DriverName:     "postgres",
			DatabaseName:   "goadmin",
			MigrationsPath: "../db/migrations/postgres",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	admin.Echo().Logger.Fatal(admin.Serve())
}
