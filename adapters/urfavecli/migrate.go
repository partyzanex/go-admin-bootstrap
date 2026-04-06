package urfavecli

import (
	"context"

	"github.com/urfave/cli/v3"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/pkg/commands"
)

func MigrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Run database migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dsn", Required: true, Usage: "PostgreSQL DSN"},
			&cli.StringFlag{Name: "direction", Value: "up", Usage: "migration direction (up|down)"},
			&cli.StringFlag{Name: "migrations-table", Value: goadmin.DefaultMigrationsTable},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			return commands.Migrate(&commands.MigrateParams{
				DSN:             cmd.String("dsn"),
				Direction:       cmd.String("direction"),
				MigrationsTable: cmd.String("migrations-table"),
			})
		},
	}
}
