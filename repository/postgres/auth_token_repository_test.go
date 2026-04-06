//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/partyzanex/testutils"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
)

func TestTokenRepository(t *testing.T) {
	dsn := os.Getenv("TEST_PG")
	if dsn == "" {
		t.Skip("TEST_PG not set")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	t.Cleanup(func() { db.Close() })

	repo := NewTokenRepository(db)

	suite.Run(t, &TokenSuite{
		db:   db,
		repo: repo.(*authTokenRepository),
	})
}

type TokenSuite struct {
	suite.Suite

	db   *bun.DB
	repo *authTokenRepository
}

func (s *TokenSuite) BeforeTest(_, _ string) {
	s.Require().NoError(migrations.Up(s.db.DB, goadmin.DefaultMigrationsTable))

	_, err := s.db.ExecContext(context.Background(), `DELETE FROM goadmin.auth_token`)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(context.Background(), `DELETE FROM goadmin."user"`)
	s.Require().NoError(err)
}

func (s *TokenSuite) TestCreate() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := s.createTestUser(ctx)
	token := s.createTestToken(ctx, user.ID, time.Now().Add(time.Hour))

	s.NotZero(token.ID)
	s.Equal(user.ID, token.UserID)
	s.NotEmpty(token.Token)
	s.Equal(goadmin.AuthToken, token.Type)
}

func (s *TokenSuite) TestSearch() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := s.createTestUser(ctx)
	created := s.createTestToken(ctx, user.ID, time.Now().Add(time.Hour))

	got, err := s.repo.Search(ctx, created.Token)
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(created.ID, got.ID)
	s.Equal(created.Token, got.Token)
	s.Equal(created.UserID, got.UserID)
	s.Equal(created.Type, got.Type)
}

func (s *TokenSuite) TestSearch_NotFound() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := s.repo.Search(ctx, testutils.RandomString(32))
	s.Require().Error(err)

	var notFound *goadmin.NotFoundError
	s.ErrorAs(err, &notFound)
}

func (s *TokenSuite) TestDeleteExpired() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := s.createTestUser(ctx)

	expired1 := s.createTestToken(ctx, user.ID, time.Now().Add(-time.Hour))
	expired2 := s.createTestToken(ctx, user.ID, time.Now().Add(-time.Minute))
	active := s.createTestToken(ctx, user.ID, time.Now().Add(time.Hour))

	rows, err := s.repo.DeleteExpired(ctx)
	s.Require().NoError(err)
	s.EqualValues(2, rows)

	_, err = s.repo.Search(ctx, expired1.Token)
	s.Error(err)

	_, err = s.repo.Search(ctx, expired2.Token)
	s.Error(err)

	got, err := s.repo.Search(ctx, active.Token)
	s.Require().NoError(err)
	s.Equal(active.ID, got.ID)
}

func (s *TokenSuite) TestDeleteExpired_NoneExpired() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := s.createTestUser(ctx)
	s.createTestToken(ctx, user.ID, time.Now().Add(time.Hour))
	s.createTestToken(ctx, user.ID, time.Now().Add(2*time.Hour))

	rows, err := s.repo.DeleteExpired(ctx)
	s.Require().NoError(err)
	s.EqualValues(0, rows)
}

func (s *TokenSuite) TestDeleteByUserID() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user1 := s.createTestUser(ctx)
	user2 := s.createTestUser(ctx)

	tok1 := s.createTestToken(ctx, user1.ID, time.Now().Add(time.Hour))
	tok2 := s.createTestToken(ctx, user1.ID, time.Now().Add(time.Hour))
	tok3 := s.createTestToken(ctx, user2.ID, time.Now().Add(time.Hour))

	err := s.repo.DeleteByUserID(ctx, user1.ID)
	s.Require().NoError(err)

	_, err = s.repo.Search(ctx, tok1.Token)
	s.Error(err)

	_, err = s.repo.Search(ctx, tok2.Token)
	s.Error(err)

	got, err := s.repo.Search(ctx, tok3.Token)
	s.Require().NoError(err)
	s.Equal(tok3.ID, got.ID)
}

func (s *TokenSuite) TestDeleteByUserID_NoTokens() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.repo.DeleteByUserID(ctx, testutils.RandInt64(999999, 9999999))
	s.NoError(err)
}

func (s *TokenSuite) createTestUser(ctx context.Context) *goadmin.User {
	userRepo := NewUserRepository(s.db)

	user, err := userRepo.Create(ctx, &goadmin.User{
		Login:    fmt.Sprintf("%s@example.com", testutils.RandomString(10)),
		Password: testutils.RandomString(64),
		Status:   goadmin.UserActive,
		Name:     testutils.RandomString(20),
		Role:     goadmin.RoleUser,
	})
	s.Require().NoError(err)

	return user
}

func (s *TokenSuite) createTestToken(ctx context.Context, userID int64, expiredAt time.Time) *goadmin.Token {
	token, err := s.repo.Create(ctx, &goadmin.Token{
		UserID:    userID,
		Token:     testutils.RandomString(32),
		Type:      goadmin.AuthToken,
		DTExpired: expiredAt,
	})
	s.Require().NoError(err)

	return token
}
