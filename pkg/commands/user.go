package commands

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

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
	errFieldsRequired   = errors.New("login, name and role are required")
	errPasswordRequired = errors.New("password is required: use --password flag, GOADMIN_PASSWORD env, or pipe via stdin")
	errPasswordEmpty    = errors.New("password cannot be empty")
)

// CreateUserParams holds parameters for the create-user command.
// Password is optional at construction time: if empty it is resolved from
// the GOADMIN_PASSWORD environment variable or read interactively from Stdin.
type CreateUserParams struct {
	DSN             string
	Login           string
	Password        string // optional — resolved from env/stdin when empty
	Name            string
	Role            string
	Migrate         bool
	MigrationsTable string
	Stdin           io.Reader // used for password prompt/pipe; defaults to os.Stdin
}

func (p *CreateUserParams) Validate() error {
	if p.DSN == "" {
		return errDSNRequired
	}

	if p.Login == "" || p.Name == "" || p.Role == "" {
		return errFieldsRequired
	}

	if p.MigrationsTable == "" {
		p.MigrationsTable = goadmin.DefaultMigrationsTable
	}

	return nil
}

// resolvePassword determines the password to use, in priority order:
//  1. flagValue (non-empty --password flag)
//  2. GOADMIN_PASSWORD environment variable
//  3. Interactive prompt with masked input if r is a TTY
//  4. First line read from r (piped stdin)
func resolvePassword(flagValue string, r io.Reader) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	if env := os.Getenv("GOADMIN_PASSWORD"); env != "" {
		return env, nil
	}

	if r == nil {
		r = os.Stdin
	}

	if f, ok := r.(*os.File); ok && term.IsTerminal(int(f.Fd())) { // #nosec G115 -- fd always fits int on supported 64-bit platforms
		fmt.Fprint(os.Stderr, "Password: ")

		b, err := term.ReadPassword(int(f.Fd())) // #nosec G115 -- same

		fmt.Fprintln(os.Stderr)

		if err != nil {
			return "", fmt.Errorf("reading password: %w", err)
		}

		pw := strings.TrimSpace(string(b))
		if pw == "" {
			return "", errPasswordEmpty
		}

		return pw, nil
	}

	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		pw := strings.TrimSpace(scanner.Text())
		if pw == "" {
			return "", errPasswordEmpty
		}

		return pw, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading password from stdin: %w", err)
	}

	return "", errPasswordRequired
}

func CreateUser(ctx context.Context, p *CreateUserParams) error {
	if err := p.Validate(); err != nil {
		return err
	}

	pw, err := resolvePassword(p.Password, p.Stdin)
	if err != nil {
		return err
	}

	p.Password = pw

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(p.DSN)))
	db := bun.NewDB(sqldb, pgdialect.New())

	defer func() { _ = db.Close() }()

	if p.Migrate {
		if migrErr := migrations.Up(db.DB, p.MigrationsTable); migrErr != nil {
			return fmt.Errorf("migration failed: %w", migrErr)
		}
	}

	userRepo := postgres.NewUserRepository(db)
	userCase := usecase.NewUserCase(userRepo, nil)

	_, err = userCase.SearchByLogin(ctx, p.Login)
	if err == nil {
		fmt.Printf("user with login %s already exists, skipping\n", p.Login)

		return nil
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
