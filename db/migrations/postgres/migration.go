package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var Content embed.FS

func Up(db *sql.DB, migrationsTable string) (err error) {
	goose.SetBaseFS(Content)
	goose.SetTableName(migrationsTable)

	err = goose.Up(db, ".")
	if err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}

	return nil
}

func Down(db *sql.DB, migrationsTable string) error {
	goose.SetBaseFS(Content)
	goose.SetTableName(migrationsTable)

	err := goose.Down(db, ".")
	if err != nil {
		return fmt.Errorf("goose.Down: %w", err)
	}

	return nil
}
