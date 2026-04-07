//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/partyzanex/testutils"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	migrations "github.com/partyzanex/go-admin-bootstrap/db/migrations/postgres"
)

func TestUserRepository(t *testing.T) {
	repo := NewUserRepository(testDB)

	suite.Run(t, &UserSuite{
		db:   testDB,
		repo: repo.(*userRepository),
	})
}

type UserSuite struct {
	suite.Suite

	db   *bun.DB
	repo *userRepository
}

func (s *UserSuite) BeforeTest(_, _ string) {
	s.Require().NoError(migrations.Up(s.db.DB, goadmin.DefaultMigrationsTable))

	_, err := s.db.ExecContext(context.Background(), `DELETE FROM goadmin.auth_token`)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(context.Background(), `DELETE FROM goadmin."user"`)
	s.Require().NoError(err)
}

//nolint:funlen
func (s *UserSuite) TestSearch() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user1 := s.createTestUser(ctx)
	user2 := s.createTestUser(ctx)
	user3 := s.createTestUser(ctx)
	user4 := s.createTestUser(ctx)
	user5 := s.createTestUser(ctx)
	user6 := s.createTestUser(ctx)
	user7 := s.createTestUser(ctx)

	testCases := []*struct {
		Name       string
		Filter     *goadmin.UserFilter
		Count      int64
		WantResult []*goadmin.User
		WantErr    string
	}{
		{
			Name: "search by id",
			Filter: &goadmin.UserFilter{
				IDs: []int64{user1.ID},
			},
			Count:      1,
			WantResult: []*goadmin.User{user1},
		},
		{
			Name: "not found by id",
			Filter: &goadmin.UserFilter{
				IDs: []int64{testutils.RandInt64(999999, 9999999)},
			},
			Count:      0,
			WantResult: nil,
		},
		{
			Name: "search by ids",
			Filter: &goadmin.UserFilter{
				IDs: []int64{user1.ID, user2.ID},
			},
			Count:      2,
			WantResult: []*goadmin.User{user1, user2},
		},
		{
			Name:       "search by name",
			Filter:     &goadmin.UserFilter{Name: user1.Name},
			Count:      1,
			WantResult: []*goadmin.User{user1},
		},
		{
			Name:       "not found by name",
			Filter:     &goadmin.UserFilter{Name: user1.Login},
			Count:      0,
			WantResult: nil,
		},
		{
			Name:       "search by login",
			Filter:     &goadmin.UserFilter{Login: user3.Login},
			Count:      1,
			WantResult: []*goadmin.User{user3},
		},
		{
			Name:       "search by status",
			Filter:     &goadmin.UserFilter{Login: user2.Login, Status: user2.Status},
			Count:      1,
			WantResult: []*goadmin.User{user2},
		},
		{
			Name: "not found by status",
			Filter: &goadmin.UserFilter{
				Login: user2.Login,
				Status: func() goadmin.UserStatus {
					switch user2.Status { //nolint:exhaustive
					case goadmin.UserBlocked, goadmin.UserActive:
						return goadmin.UserNew
					default:
						return goadmin.UserBlocked
					}
				}(),
			},
			Count:      0,
			WantResult: nil,
		},
		{
			Name: "limit",
			Filter: &goadmin.UserFilter{
				IDs:   []int64{user1.ID, user2.ID, user3.ID},
				Limit: 1,
			},
			Count:      3,
			WantResult: []*goadmin.User{user1},
		},
		{
			Name: "limit and offset",
			Filter: &goadmin.UserFilter{
				IDs:    []int64{user1.ID, user2.ID, user3.ID},
				Limit:  2,
				Offset: 1,
			},
			Count:      3,
			WantResult: []*goadmin.User{user2, user3},
		},
		{
			Name:   "search all",
			Filter: nil,
			Count:  7,
			WantResult: []*goadmin.User{
				user1, user2, user3, user4, user5, user6, user7,
			},
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.Name, func() {
			result, err := s.repo.Search(ctx, testCase.Filter)

			if testCase.WantErr != "" {
				s.EqualError(err, testCase.WantErr)
			} else {
				s.NoError(err)
			}

			if testCase.WantResult != nil {
				s.NotNil(result)
				s.Len(result, len(testCase.WantResult))
			} else {
				s.Empty(result)
			}

			for i, got := range result {
				if s.Equal(testCase.WantResult[i].DTCreated.Unix(), got.DTCreated.Unix()) {
					testCase.WantResult[i].DTCreated = got.DTCreated
				}

				if s.Equal(testCase.WantResult[i].DTUpdated.Unix(), got.DTUpdated.Unix()) {
					testCase.WantResult[i].DTUpdated = got.DTUpdated
				}

				if s.Equal(testCase.WantResult[i].DTLastLogged.Unix(), got.DTLastLogged.Unix()) {
					testCase.WantResult[i].DTLastLogged = got.DTLastLogged
				}

				s.EqualValues(testCase.WantResult[i], got)
			}

			count, err := s.repo.Count(ctx, testCase.Filter)
			s.NoError(err)
			s.Equal(testCase.Count, count)
		})
	}
}

func (s *UserSuite) TestCreate() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, err := s.repo.Create(ctx, &goadmin.User{
		Login:    fmt.Sprintf("%s@example.com", testutils.RandomString(10)),
		Password: testutils.RandomString(64),
		Status:   goadmin.UserActive,
		Name:     testutils.RandomString(20),
		Role:     goadmin.RoleUser,
	})
	s.Require().NoError(err)
	s.Require().NotNil(user)
	s.NotZero(user.ID)
	s.NotZero(user.DTCreated)
	s.NotZero(user.DTUpdated)
	s.True(user.PasswordIsEncoded)
}

func (s *UserSuite) TestUpdate() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	created := s.createTestUser(ctx)

	created.Name = testutils.RandomString(20)
	created.Status = goadmin.UserBlocked
	created.Role = goadmin.RoleOwner

	updated, err := s.repo.Update(ctx, created)
	s.Require().NoError(err)
	s.Require().NotNil(updated)
	s.Equal(created.ID, updated.ID)
	s.Equal(created.Name, updated.Name)
	s.Equal(goadmin.UserBlocked, updated.Status)
	s.Equal(goadmin.RoleOwner, updated.Role)
}

func (s *UserSuite) TestSetLastLogged() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := s.createTestUser(ctx)
	s.Zero(user.DTLastLogged)

	err := s.repo.SetLastLogged(ctx, user)
	s.Require().NoError(err)

	result, err := s.repo.GetUserByID(ctx, user.ID)
	s.Require().NoError(err)
	s.NotZero(result.DTLastLogged)
}

func (s *UserSuite) TestDelete() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := s.createTestUser(ctx)

	err := s.repo.Delete(ctx, user)
	s.Require().NoError(err)

	_, err = s.repo.GetUserByID(ctx, user.ID)
	s.Require().Error(err)

	var notFound *goadmin.NotFoundError
	s.ErrorAs(err, &notFound)
}

func (s *UserSuite) TestDelete_ZeroID() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.repo.Delete(ctx, &goadmin.User{ID: 0})
	s.ErrorIs(err, goadmin.ErrRequiredUserID)
}

func (s *UserSuite) TestDelete_NotFound() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.repo.Delete(ctx, &goadmin.User{ID: testutils.RandInt64(999999, 9999999)})
	s.Require().Error(err)

	var notFound *goadmin.NotFoundError
	s.ErrorAs(err, &notFound)
}

func (s *UserSuite) TestGetUserByID() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	created := s.createTestUser(ctx)

	got, err := s.repo.GetUserByID(ctx, created.ID)
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(created.ID, got.ID)
	s.Equal(created.Login, got.Login)
	s.Equal(created.Name, got.Name)
	s.Equal(created.Status, got.Status)
	s.Equal(created.Role, got.Role)
}

func (s *UserSuite) TestGetUserByID_NotFound() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := s.repo.GetUserByID(ctx, testutils.RandInt64(999999, 9999999))
	s.Require().Error(err)

	var notFound *goadmin.NotFoundError
	s.ErrorAs(err, &notFound)
}

func (s *UserSuite) createTestUser(ctx context.Context) *goadmin.User {
	statuses := []any{
		goadmin.UserActive,
		goadmin.UserBlocked,
		goadmin.UserNew,
	}
	roles := []any{
		goadmin.RoleUser,
		goadmin.RoleRoot,
		goadmin.RoleOwner,
	}

	user, err := s.repo.Create(ctx, &goadmin.User{
		Login:             fmt.Sprintf("%s@example.com", testutils.RandomString(10)),
		Password:          testutils.RandomString(64),
		Status:            testutils.RandomCase(statuses...).(goadmin.UserStatus),
		Name:              testutils.RandomString(20),
		Role:              testutils.RandomCase(roles...).(goadmin.UserRole),
		DTCreated:         time.Time{},
		DTUpdated:         time.Time{},
		DTLastLogged:      time.Time{},
		PasswordIsEncoded: false,
		Current:           false,
	})
	s.Require().NoError(err)

	return user
}
