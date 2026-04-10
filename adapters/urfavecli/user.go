package urfavecli

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/pkg/commands"
)

func CreateUserCommand() *cli.Command {
	return &cli.Command{
		Name:  "create-user",
		Usage: "Create an admin panel user",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dsn", Required: true, Usage: "PostgreSQL DSN"},
			&cli.StringFlag{Name: "login", Required: true, Usage: "user email"},
			&cli.StringFlag{Name: "password", Usage: "password (dev/CI shortcut; prefer GOADMIN_PASSWORD env or interactive prompt)"},
			&cli.StringFlag{Name: "name", Required: true, Usage: "user display name"},
			&cli.StringFlag{Name: "role", Value: "user", Usage: "user role (owner|root|user)"},
			&cli.BoolFlag{Name: "migrate", Usage: "run migrations before creating user"},
			&cli.StringFlag{Name: "migrations-table", Value: goadmin.DefaultMigrationsTable},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return commands.CreateUser(ctx, &commands.CreateUserParams{
				DSN:             cmd.String("dsn"),
				Login:           cmd.String("login"),
				Password:        cmd.String("password"),
				Name:            cmd.String("name"),
				Role:            cmd.String("role"),
				Migrate:         cmd.Bool("migrate"),
				MigrationsTable: cmd.String("migrations-table"),
				Stdin:           os.Stdin,
			})
		},
	}
}
