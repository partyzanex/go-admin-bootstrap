package migrations_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun/driver/pgdriver"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
)

func TestUp(t *testing.T) {
	dsn := os.Getenv("CRYPCHS_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable"
	}

	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	require.NotNil(t, db)

	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()

	err := migrations.Up(db, goadmin.MigrationsTable)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin."user"`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin.auth_token`)
	require.NoError(t, err)
}
