package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/partyzanex/go-admin-bootstrap/adapters/urfavecli"
)

func main() {
	app := &cli.Command{
		Name:  "admin",
		Usage: "Admin panel management tool (urfave/cli)",
		Commands: []*cli.Command{
			urfavecli.CreateUserCommand(),
			urfavecli.MigrateCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
