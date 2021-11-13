package migrations_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
)

func TestUp(t *testing.T) {
	dsn := os.Getenv("CRYPCHS_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NotNil(t, db)

	err = migrations.Up(db, goadmin.MigrationsTable)
	require.NoError(t, err)

	_, err = db.Exec(`select * from goadmin."user"`)
	require.NoError(t, err)

	err = migrations.Down(db, goadmin.MigrationsTable)
	require.NoError(t, err)

	_, err = db.Exec(`select * from goadmin."user"`)
	require.Error(t, err)
}
