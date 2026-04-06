package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/spf13/pflag"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
	"github.com/partyzanex/go-admin-bootstrap/repository/postgres"
	"github.com/partyzanex/go-admin-bootstrap/usecase"
)

func main() {
	var (
		dsn             = pflag.String("dsn", "", "postgres dsn")
		login           = pflag.String("login", "", "user login")
		password        = pflag.String("password", "", "user password")
		name            = pflag.String("name", "", "user name")
		role            = pflag.String("role", "", "user role name, available: owner, root, user")
		migrate         = pflag.Bool("mig", false, "if need up migrations")
		migrationsTable = pflag.String("migrations-table", "goadmin-migrations", "migration table name")
	)

	pflag.Parse()

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		fmt.Printf("open sql connection failed: %s\n", err)
		return
	}

	if *login == "" || *password == "" || *name == "" || *role == "" {
		fmt.Println("user name, login and password are required")
		return
	}

	if *migrate {
		err = migrations.Up(db, *migrationsTable)
		if err != nil {
			fmt.Println("migration failed")
			return
		}
	}

	userRepo := postgres.NewUserRepository(db)
	userCase := usecase.NewUserCase(userRepo, nil)

	ctx := context.Background()

	_, err = userCase.SearchByLogin(ctx, *login)
	if err == nil {
		fmt.Printf("user with login %s already exists\n", *login)

		return
	}

	if !errors.Is(err, goadmin.ErrUserNotFound) {
		fmt.Printf("searching user failed: %s\n", err)

		return
	}

	user := &goadmin.User{
		Login:    *login,
		Password: *password,
		Status:   goadmin.UserActive,
		Name:     *name,
		Role:     goadmin.UserRole(*role),
	}

	err = userCase.Register(ctx, user)
	if err != nil {
		fmt.Printf("register failed: %s\n", err)
		return
	}

	fmt.Printf("user successful created with id %d\n", user.ID)
}
