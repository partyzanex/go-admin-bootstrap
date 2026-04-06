package migrations_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

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

	ctx := context.Background()

	err = migrations.Up(db, goadmin.MigrationsTable)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin."user"`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin.auth_token`)
	require.NoError(t, err)
}
