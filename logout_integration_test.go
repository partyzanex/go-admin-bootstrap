//go:build integration

package goadmin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
	pgrepository "github.com/partyzanex/go-admin-bootstrap/repository/postgres"
	"github.com/partyzanex/go-admin-bootstrap/usecase"
)

// TestLogout_RevokesTokenFromDB verifies that posting to /logout with a valid
// refresh cookie causes the corresponding token to be deleted from the database.
func TestLogout_RevokesTokenFromDB(t *testing.T) {
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

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, migrations.Up(db.DB, DefaultMigrationsTable))

	userRepo := pgrepository.NewUserRepository(db)
	tokenRepo := pgrepository.NewTokenRepository(db)
	uc := usecase.NewUserCase(userRepo, tokenRepo)

	// Create a test user.
	user, err := userRepo.Create(ctx, &User{
		Login:             fmt.Sprintf("logout-test-%d@example.com", time.Now().UnixNano()),
		Password:          "alreadyhashed",
		Status:            UserActive,
		Name:              "Logout Test User",
		Role:              RoleUser,
		PasswordIsEncoded: true,
	})
	require.NoError(t, err)

	// The app config uses JWTSecret "test-secret-32-bytes-long!!!!!!!!".
	// Logout hashes the raw cookie value with that key before querying the DB.
	jwtSecret := []byte("test-secret-32-bytes-long!!!!!!!!")
	rawRefreshValue := "raw-refresh-token-for-logout-test"
	hashedToken := hashRefreshToken(jwtSecret, rawRefreshValue)

	// Insert the token as the DB would store it (hashed).
	_, err = tokenRepo.Create(ctx, &Token{
		UserID:    user.ID,
		Token:     hashedToken,
		Type:      AuthToken,
		DTExpired: time.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	// Confirm the token is present before logout.
	_, err = tokenRepo.Search(ctx, hashedToken)
	require.NoError(t, err, "token must exist before logout")

	// Set up the routed app with the real use case.
	app := newTestApp(t)
	app.config.UserCase = uc
	// newTestApp already sets JWTSecret to the same value; be explicit.
	app.config.JWTSecret = jwtSecret
	app.setDefaultMiddleware()
	app.setStaticGroup()
	app.setDefaultRoutes()

	// Obtain a CSRF token via GET /admin/login.
	csrfToken := csrfTokenFromGET(t, app, "/admin/login")
	require.NotEmpty(t, csrfToken)

	// POST /admin/logout carrying the refresh cookie.
	// newTestApp sets AccessCookieName = "test", so the refresh cookie is "test_refresh".
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: csrfToken})
	req.AddCookie(&http.Cookie{Name: "test" + refreshCookieSuffix, Value: rawRefreshValue})

	rec := httptest.NewRecorder()
	app.echo.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code, "logout should redirect to login")

	// Verify the token has been deleted from the database.
	_, err = tokenRepo.Search(ctx, hashedToken)
	require.Error(t, err, "token must be deleted from DB after logout")

	var notFound *NotFoundError
	require.ErrorAs(t, err, &notFound, "error should be a NotFoundError")
}
