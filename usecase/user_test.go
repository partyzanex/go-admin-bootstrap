package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

// Mock implementations

type mockUserRepo struct {
	searchFn     func(ctx context.Context, filter *goadmin.UserFilter) ([]*goadmin.User, error)
	countFn      func(ctx context.Context, filter *goadmin.UserFilter) (int64, error)
	createFn     func(ctx context.Context, user *goadmin.User) (*goadmin.User, error)
	updateFn     func(ctx context.Context, user *goadmin.User) (*goadmin.User, error)
	setLastLogFn func(ctx context.Context, user *goadmin.User) error
	deleteFn     func(ctx context.Context, user *goadmin.User) error
}

func (m *mockUserRepo) Search(ctx context.Context, filter *goadmin.UserFilter) ([]*goadmin.User, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockUserRepo) Count(ctx context.Context, filter *goadmin.UserFilter) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx, filter)
	}
	return 0, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *goadmin.User) (*goadmin.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return user, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *goadmin.User) (*goadmin.User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return user, nil
}

func (m *mockUserRepo) SetLastLogged(ctx context.Context, user *goadmin.User) error {
	if m.setLastLogFn != nil {
		return m.setLastLogFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, user *goadmin.User) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, user)
	}
	return nil
}

type mockTokenRepo struct {
	searchFn         func(ctx context.Context, token string) (*goadmin.Token, error)
	createFn         func(ctx context.Context, token *goadmin.Token) (*goadmin.Token, error)
	deleteExpiredFn  func(ctx context.Context) (int64, error)
	deleteByUserIDFn func(ctx context.Context, userID int64) error
}

func (m *mockTokenRepo) Search(ctx context.Context, token string) (*goadmin.Token, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, token)
	}
	return nil, nil
}

func (m *mockTokenRepo) Create(ctx context.Context, token *goadmin.Token) (*goadmin.Token, error) {
	if m.createFn != nil {
		return m.createFn(ctx, token)
	}
	return token, nil
}

func (m *mockTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	if m.deleteExpiredFn != nil {
		return m.deleteExpiredFn(ctx)
	}
	return 0, nil
}

func (m *mockTokenRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	if m.deleteByUserIDFn != nil {
		return m.deleteByUserIDFn(ctx, userID)
	}
	return nil
}

// Helper to create a valid user for tests.
func validUser(id int64) *goadmin.User {
	return &goadmin.User{
		ID:       id,
		Login:    "user@example.com",
		Password: "StrongPass1",
		Status:   goadmin.UserActive,
		Name:     "Test User",
		Role:     goadmin.RoleUser,
	}
}

func newUseCase(users *mockUserRepo, tokens *mockTokenRepo) goadmin.UserUseCase {
	return NewUserCase(users, tokens)
}

// --- Validate tests ---

func TestValidate_CreateValid(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0) // ID=0 is fine for create
	err := uc.Validate(user, true)
	assert.NoError(t, err)
}

func TestValidate_UpdateValid(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.PasswordIsEncoded = true
	err := uc.Validate(user, false)
	assert.NoError(t, err)
}

func TestValidate_UpdateRequiresID(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	err := uc.Validate(user, false)
	assert.ErrorIs(t, err, goadmin.ErrRequiredUserID)
}

func TestValidate_MissingName(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Name = ""
	err := uc.Validate(user, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidate_MissingLogin(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Login = ""
	err := uc.Validate(user, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidate_InvalidEmail(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Login = "not-an-email"
	err := uc.Validate(user, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidate_MissingPassword(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Password = ""
	user.PasswordIsEncoded = false
	err := uc.Validate(user, true)
	assert.ErrorIs(t, err, goadmin.ErrRequiredUserPassword)
}

func TestValidate_EncodedPasswordAllowsEmpty(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.Password = ""
	user.PasswordIsEncoded = true
	err := uc.Validate(user, false)
	assert.NoError(t, err)
}

func TestValidate_InvalidStatus(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Status = "invalid"
	err := uc.Validate(user, true)
	assert.ErrorIs(t, err, goadmin.ErrInvalidUserStatus)
}

func TestValidate_InvalidRole(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Role = "invalid"
	err := uc.Validate(user, true)
	assert.ErrorIs(t, err, goadmin.ErrInvalidUserRole)
}

// --- SearchByLogin tests ---

func TestSearchByLogin_Found(t *testing.T) {
	expected := validUser(1)
	users := &mockUserRepo{
		searchFn: func(_ context.Context, f *goadmin.UserFilter) ([]*goadmin.User, error) {
			assert.Equal(t, "user@example.com", f.Login)
			assert.Equal(t, 1, f.Limit)
			return []*goadmin.User{expected}, nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	result, err := uc.SearchByLogin(context.Background(), "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSearchByLogin_NotFound(t *testing.T) {
	users := &mockUserRepo{
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return nil, nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	_, err := uc.SearchByLogin(context.Background(), "missing@example.com")
	require.Error(t, err)
	assert.True(t, goadmin.IsNotFound(err))
}

func TestSearchByLogin_RepoError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	users := &mockUserRepo{
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return nil, repoErr
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	_, err := uc.SearchByLogin(context.Background(), "user@example.com")
	assert.ErrorIs(t, err, repoErr)
}

// --- SearchByID tests ---

func TestSearchByID_Found(t *testing.T) {
	expected := validUser(42)
	users := &mockUserRepo{
		searchFn: func(_ context.Context, f *goadmin.UserFilter) ([]*goadmin.User, error) {
			assert.Equal(t, []int64{42}, f.IDs)
			return []*goadmin.User{expected}, nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	result, err := uc.SearchByID(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSearchByID_NotFound(t *testing.T) {
	users := &mockUserRepo{
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return nil, nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	_, err := uc.SearchByID(context.Background(), 999)
	require.Error(t, err)
	assert.True(t, goadmin.IsNotFound(err))
}

func TestSearchByID_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	users := &mockUserRepo{
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return nil, repoErr
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	_, err := uc.SearchByID(context.Background(), 1)
	assert.ErrorIs(t, err, repoErr)
}

// --- SetLastLogged tests ---

func TestSetLastLogged_Success(t *testing.T) {
	var capturedUser *goadmin.User
	users := &mockUserRepo{
		setLastLogFn: func(_ context.Context, u *goadmin.User) error {
			capturedUser = u
			return nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	user := validUser(1)
	before := time.Now()
	err := uc.SetLastLogged(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, user, capturedUser)
	assert.False(t, user.DTLastLogged.Before(before))
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	created := validUser(10)
	created.PasswordIsEncoded = true
	users := &mockUserRepo{
		createFn: func(_ context.Context, u *goadmin.User) (*goadmin.User, error) {
			assert.True(t, u.PasswordIsEncoded)
			assert.NotEqual(t, "StrongPass1", u.Password) // should be hashed
			result := *u
			result.ID = 10
			return &result, nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	user := validUser(0)
	err := uc.Register(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, int64(10), user.ID)
	assert.True(t, user.PasswordIsEncoded)
}

func TestRegister_ValidationError(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Name = "" // invalid
	err := uc.Register(context.Background(), user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestRegister_WeakPassword(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0)
	user.Password = "weak"
	err := uc.Register(context.Background(), user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password policy")
}

func TestRegister_RepoError(t *testing.T) {
	repoErr := errors.New("create failed")
	users := &mockUserRepo{
		createFn: func(_ context.Context, _ *goadmin.User) (*goadmin.User, error) {
			return nil, repoErr
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	user := validUser(0)
	err := uc.Register(context.Background(), user)
	assert.ErrorIs(t, err, repoErr)
}

// --- UpdateUser tests ---

func TestUpdateUser_Success(t *testing.T) {
	users := &mockUserRepo{
		updateFn: func(_ context.Context, u *goadmin.User) (*goadmin.User, error) {
			result := *u
			result.DTUpdated = time.Now()
			return &result, nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	user := validUser(1)
	user.PasswordIsEncoded = true
	result, err := uc.UpdateUser(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
}

func TestUpdateUser_ValidationError_NoID(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(0) // ID=0 invalid for update
	_, err := uc.UpdateUser(context.Background(), user)
	assert.ErrorIs(t, err, goadmin.ErrRequiredUserID)
}

func TestUpdateUser_RepoError(t *testing.T) {
	repoErr := errors.New("update failed")
	users := &mockUserRepo{
		updateFn: func(_ context.Context, _ *goadmin.User) (*goadmin.User, error) {
			return nil, repoErr
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	user := validUser(1)
	user.PasswordIsEncoded = true
	_, err := uc.UpdateUser(context.Background(), user)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

// --- DeleteUser tests ---

func TestDeleteUser_Success(t *testing.T) {
	var deletedID int64
	users := &mockUserRepo{
		deleteFn: func(_ context.Context, u *goadmin.User) error {
			deletedID = u.ID
			return nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	err := uc.DeleteUser(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), deletedID)
}

func TestDeleteUser_RepoError(t *testing.T) {
	repoErr := errors.New("delete failed")
	users := &mockUserRepo{
		deleteFn: func(_ context.Context, _ *goadmin.User) error {
			return repoErr
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	err := uc.DeleteUser(context.Background(), 1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

// --- ListUsers tests ---

func TestListUsers_Success(t *testing.T) {
	expectedUsers := []*goadmin.User{validUser(1), validUser(2)}
	users := &mockUserRepo{
		countFn: func(_ context.Context, _ *goadmin.UserFilter) (int64, error) {
			return 2, nil
		},
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return expectedUsers, nil
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	result, count, err := uc.ListUsers(context.Background(), &goadmin.UserFilter{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Len(t, result, 2)
}

func TestListUsers_CountError(t *testing.T) {
	repoErr := errors.New("count failed")
	users := &mockUserRepo{
		countFn: func(_ context.Context, _ *goadmin.UserFilter) (int64, error) {
			return 0, repoErr
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	_, _, err := uc.ListUsers(context.Background(), &goadmin.UserFilter{})
	assert.ErrorIs(t, err, repoErr)
}

func TestListUsers_SearchError(t *testing.T) {
	repoErr := errors.New("search failed")
	users := &mockUserRepo{
		countFn: func(_ context.Context, _ *goadmin.UserFilter) (int64, error) {
			return 5, nil
		},
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return nil, repoErr
		},
	}
	uc := newUseCase(users, &mockTokenRepo{})
	_, _, err := uc.ListUsers(context.Background(), &goadmin.UserFilter{})
	assert.ErrorIs(t, err, repoErr)
}

// --- EncodePassword tests ---

func TestEncodePassword_NewPassword(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.Password = "StrongPass1"
	user.PasswordIsEncoded = false
	err := uc.EncodePassword(user)
	require.NoError(t, err)
	assert.True(t, user.PasswordIsEncoded)
	assert.Contains(t, user.Password, "$argon2id$")
}

func TestEncodePassword_AlreadyEncoded(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.Password = "$argon2id$existing-hash"
	user.PasswordIsEncoded = true
	err := uc.EncodePassword(user)
	require.NoError(t, err)
	assert.Equal(t, "$argon2id$existing-hash", user.Password) // unchanged
}

func TestEncodePassword_WeakPassword(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.Password = "short"
	user.PasswordIsEncoded = false
	err := uc.EncodePassword(user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password policy")
	assert.False(t, user.PasswordIsEncoded)
}

// --- ComparePassword tests ---

func TestComparePassword_Correct(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.Password = "StrongPass1"
	user.PasswordIsEncoded = false
	err := uc.EncodePassword(user)
	require.NoError(t, err)

	ok, err := uc.ComparePassword(user, "StrongPass1")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestComparePassword_Wrong(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.Password = "StrongPass1"
	user.PasswordIsEncoded = false
	err := uc.EncodePassword(user)
	require.NoError(t, err)

	ok, err := uc.ComparePassword(user, "WrongPass1")
	assert.False(t, ok)
	assert.ErrorIs(t, err, goadmin.ErrWrongPassword)
}

func TestComparePassword_NotEncoded(t *testing.T) {
	uc := newUseCase(&mockUserRepo{}, &mockTokenRepo{})
	user := validUser(1)
	user.PasswordIsEncoded = false

	ok, err := uc.ComparePassword(user, "anything")
	assert.NoError(t, err)
	assert.False(t, ok)
}

// --- CreateAuthToken tests ---

func TestCreateAuthToken_Success(t *testing.T) {
	tokens := &mockTokenRepo{
		createFn: func(_ context.Context, tok *goadmin.Token) (*goadmin.Token, error) {
			assert.Equal(t, goadmin.AuthToken, tok.Type)
			assert.Equal(t, "cookie-value", tok.Token)
			assert.Equal(t, int64(1), tok.UserID)
			result := *tok
			result.ID = 100
			return &result, nil
		},
	}
	uc := newUseCase(&mockUserRepo{}, tokens)
	user := validUser(1)
	tok, err := uc.CreateAuthToken(context.Background(), user, "cookie-value", time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(100), tok.ID)
	assert.Equal(t, user, tok.User)
}

func TestCreateAuthToken_RepoError(t *testing.T) {
	repoErr := errors.New("token create failed")
	tokens := &mockTokenRepo{
		createFn: func(_ context.Context, _ *goadmin.Token) (*goadmin.Token, error) {
			return nil, repoErr
		},
	}
	uc := newUseCase(&mockUserRepo{}, tokens)
	user := validUser(1)
	_, err := uc.CreateAuthToken(context.Background(), user, "val", time.Hour)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

// --- SearchToken tests ---

func TestSearchToken_Found(t *testing.T) {
	user := validUser(1)
	tokens := &mockTokenRepo{
		searchFn: func(_ context.Context, tok string) (*goadmin.Token, error) {
			return &goadmin.Token{
				ID:        1,
				UserID:    1,
				Token:     tok,
				DTExpired: time.Now().Add(time.Hour),
			}, nil
		},
	}
	users := &mockUserRepo{
		searchFn: func(_ context.Context, f *goadmin.UserFilter) ([]*goadmin.User, error) {
			return []*goadmin.User{user}, nil
		},
	}
	uc := newUseCase(users, tokens)
	result, err := uc.SearchToken(context.Background(), "valid-token")
	require.NoError(t, err)
	assert.Equal(t, user, result.User)
	assert.Equal(t, "valid-token", result.Token)
}

func TestSearchToken_Expired(t *testing.T) {
	user := validUser(1)
	tokens := &mockTokenRepo{
		searchFn: func(_ context.Context, tok string) (*goadmin.Token, error) {
			return &goadmin.Token{
				ID:        1,
				UserID:    1,
				Token:     tok,
				DTExpired: time.Now().Add(-time.Hour), // expired
			}, nil
		},
	}
	users := &mockUserRepo{
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return []*goadmin.User{user}, nil
		},
	}
	uc := newUseCase(users, tokens)
	result, err := uc.SearchToken(context.Background(), "expired-token")
	assert.Error(t, err)
	assert.True(t, goadmin.IsExpired(err))
	assert.NotNil(t, result) // still returns token even when expired
}

func TestSearchToken_NotFound(t *testing.T) {
	repoErr := errors.New("not found")
	tokens := &mockTokenRepo{
		searchFn: func(_ context.Context, _ string) (*goadmin.Token, error) {
			return nil, repoErr
		},
	}
	uc := newUseCase(&mockUserRepo{}, tokens)
	_, err := uc.SearchToken(context.Background(), "missing-token")
	assert.Error(t, err)
}

func TestSearchToken_UserNotFound(t *testing.T) {
	tokens := &mockTokenRepo{
		searchFn: func(_ context.Context, tok string) (*goadmin.Token, error) {
			return &goadmin.Token{
				ID:        1,
				UserID:    999,
				Token:     tok,
				DTExpired: time.Now().Add(time.Hour),
			}, nil
		},
	}
	users := &mockUserRepo{
		searchFn: func(_ context.Context, _ *goadmin.UserFilter) ([]*goadmin.User, error) {
			return nil, nil // empty result = not found
		},
	}
	uc := newUseCase(users, tokens)
	_, err := uc.SearchToken(context.Background(), "token-orphan")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search user failed")
}

// --- RevokeUserTokens tests ---

func TestRevokeUserTokens_Success(t *testing.T) {
	var revokedUserID int64
	tokens := &mockTokenRepo{
		deleteByUserIDFn: func(_ context.Context, userID int64) error {
			revokedUserID = userID
			return nil
		},
	}
	uc := newUseCase(&mockUserRepo{}, tokens)
	err := uc.RevokeUserTokens(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, int64(42), revokedUserID)
}

func TestRevokeUserTokens_Error(t *testing.T) {
	repoErr := errors.New("revoke failed")
	tokens := &mockTokenRepo{
		deleteByUserIDFn: func(_ context.Context, _ int64) error {
			return repoErr
		},
	}
	uc := newUseCase(&mockUserRepo{}, tokens)
	err := uc.RevokeUserTokens(context.Background(), 1)
	assert.ErrorIs(t, err, repoErr)
}

// --- CleanupExpiredTokens tests ---

func TestCleanupExpiredTokens_Success(t *testing.T) {
	tokens := &mockTokenRepo{
		deleteExpiredFn: func(_ context.Context) (int64, error) {
			return 5, nil
		},
	}
	uc := newUseCase(&mockUserRepo{}, tokens)
	count, err := uc.CleanupExpiredTokens(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestCleanupExpiredTokens_Error(t *testing.T) {
	repoErr := errors.New("cleanup failed")
	tokens := &mockTokenRepo{
		deleteExpiredFn: func(_ context.Context) (int64, error) {
			return 0, repoErr
		},
	}
	uc := newUseCase(&mockUserRepo{}, tokens)
	_, err := uc.CleanupExpiredTokens(context.Background())
	assert.ErrorIs(t, err, repoErr)
}
