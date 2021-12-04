package migrations

import (
	"database/sql"
	"embed"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var Content embed.FS

func Up(db *sql.DB, migrationsTable string) (err error) {
	goose.SetBaseFS(Content)
	goose.SetTableName(migrationsTable)

	err = goose.Up(db, ".")
	if err != nil {
		return errors.Wrap(err, "goose.Up")
	}

	return nil
}

func Down(db *sql.DB, migrationsTable string) error {
	goose.SetBaseFS(Content)
	goose.SetTableName(migrationsTable)

	err := goose.Down(db, ".")
	if err != nil {
		return errors.Wrap(err, "goose.Up")
	}

	return nil
}
