package cobracli

import (
	"github.com/spf13/cobra"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/pkg/commands"
)

func MigrateCmd() *cobra.Command {
	p := commands.MigrateParams{}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(_ *cobra.Command, _ []string) error {
			return commands.Migrate(&p)
		},
	}

	cmd.Flags().StringVar(&p.DSN, "dsn", "", "PostgreSQL DSN")
	cmd.Flags().StringVar(&p.Direction, "direction", "up", "migration direction (up|down)")
	cmd.Flags().StringVar(&p.MigrationsTable, "migrations-table", goadmin.DefaultMigrationsTable, "migrations table name")

	_ = cmd.MarkFlagRequired("dsn")

	return cmd
}
