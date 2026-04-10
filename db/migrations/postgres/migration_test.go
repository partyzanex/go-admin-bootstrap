//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun/driver/pgdriver"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
)

func TestUp(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx, "postgres:14-alpine",
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	t.Cleanup(func() { _ = db.Close() })

	err = migrations.Up(db, goadmin.DefaultMigrationsTable)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin."user"`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `select * from goadmin.auth_token`)
	require.NoError(t, err)
}
