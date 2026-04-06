//go:build integration

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
	dsn := os.Getenv("TEST_PG")
	if dsn == "" {
		t.Skip("TEST_PG not set, skipping integration test")
	}

	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	require.NotNil(t, db)

	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()

	err := migrations.Up(db, goadmin.DefaultMigrationsTable)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin."user"`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin.auth_token`)
	require.NoError(t, err)
}
