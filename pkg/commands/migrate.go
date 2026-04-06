package commands

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
)

var errDirectionRequired = errors.New("direction must be 'up' or 'down'")

type MigrateParams struct {
	DSN             string
	MigrationsTable string
	Direction       string // "up" or "down"
}

func (p *MigrateParams) Validate() error {
	if p.DSN == "" {
		return errDSNRequired
	}

	if p.Direction != "up" && p.Direction != "down" {
		return errDirectionRequired
	}

	if p.MigrationsTable == "" {
		p.MigrationsTable = goadmin.DefaultMigrationsTable
	}

	return nil
}

func Migrate(p *MigrateParams) error {
	if err := p.Validate(); err != nil {
		return err
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(p.DSN)))
	db := bun.NewDB(sqldb, pgdialect.New())

	defer func() { _ = db.Close() }()

	switch p.Direction {
	case "up":
		if err := migrations.Up(db.DB, p.MigrationsTable); err != nil {
			return fmt.Errorf("migration up: %w", err)
		}
	case "down":
		if err := migrations.Down(db.DB, p.MigrationsTable); err != nil {
			return fmt.Errorf("migration down: %w", err)
		}
	}

	return nil
}
