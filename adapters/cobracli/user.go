package cobracli

import (
	"github.com/spf13/cobra"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/pkg/commands"
)

func CreateUserCmd() *cobra.Command {
	p := commands.CreateUserParams{}

	cmd := &cobra.Command{
		Use:   "create-user",
		Short: "Create an admin panel user",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return commands.CreateUser(cmd.Context(), &p)
		},
	}

	cmd.Flags().StringVar(&p.DSN, "dsn", "", "PostgreSQL DSN")
	cmd.Flags().StringVar(&p.Login, "login", "", "user email")
	cmd.Flags().StringVar(&p.Password, "password", "", "user password")
	cmd.Flags().StringVar(&p.Name, "name", "", "user display name")
	cmd.Flags().StringVar(&p.Role, "role", "user", "user role (owner|root|user)")
	cmd.Flags().BoolVar(&p.Migrate, "migrate", false, "run migrations before creating user")
	cmd.Flags().StringVar(&p.MigrationsTable, "migrations-table", goadmin.DefaultMigrationsTable, "migrations table name")

	_ = cmd.MarkFlagRequired("dsn")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
