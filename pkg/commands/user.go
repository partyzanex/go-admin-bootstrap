package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
	"github.com/partyzanex/go-admin-bootstrap/repository/postgres"
	"github.com/partyzanex/go-admin-bootstrap/usecase"
)

var (
	errDSNRequired      = errors.New("dsn is required")
	errFieldsRequired   = errors.New("login, password, name and role are required")
	errUserAlreadyExist = errors.New("user already exists")
)

type CreateUserParams struct {
	DSN             string
	Login           string
	Password        string
	Name            string
	Role            string
	Migrate         bool
	MigrationsTable string
}

func (p *CreateUserParams) Validate() error {
	if p.DSN == "" {
		return errDSNRequired
	}

	if p.Login == "" || p.Password == "" || p.Name == "" || p.Role == "" {
		return errFieldsRequired
	}

	if p.MigrationsTable == "" {
		p.MigrationsTable = goadmin.DefaultMigrationsTable
	}

	return nil
}

func CreateUser(ctx context.Context, p *CreateUserParams) error {
	if err := p.Validate(); err != nil {
		return err
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(p.DSN)))
	db := bun.NewDB(sqldb, pgdialect.New())

	defer func() { _ = db.Close() }()

	if p.Migrate {
		if err := migrations.Up(db.DB, p.MigrationsTable); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	userRepo := postgres.NewUserRepository(db)
	userCase := usecase.NewUserCase(userRepo, nil)

	_, err := userCase.SearchByLogin(ctx, p.Login)
	if err == nil {
		return fmt.Errorf("%w: %s", errUserAlreadyExist, p.Login)
	}

	if !goadmin.IsNotFound(err) {
		return fmt.Errorf("searching user: %w", err)
	}

	user := &goadmin.User{
		Login:    p.Login,
		Password: p.Password,
		Status:   goadmin.UserActive,
		Name:     p.Name,
		Role:     goadmin.UserRole(p.Role),
	}

	if err := userCase.Register(ctx, user); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	fmt.Printf("user created with id %d\n", user.ID)

	return nil
}
