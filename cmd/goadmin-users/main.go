package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/partyzanex/go-admin-bootstrap/adapters/cobracli"
)

func main() {
	root := &cobra.Command{
		Use:   "goadmin-users",
		Short: "Admin panel user management",
	}

	root.AddCommand(cobracli.CreateUserCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
