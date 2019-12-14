package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/asaskevich/govalidator"
	astistring "github.com/asticode/go-astitools/string"
	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
)

type userCase struct {
	Users  goadmin.UserRepository
	Tokens goadmin.TokenRepository
}

func (uc *userCase) Validate(user *goadmin.User, create bool) error {
	if !create && user.ID == 0 {
		return goadmin.ErrRequiredUserID
	}

	if user.Name == "" {
		return goadmin.ErrRequiredUserName
	}

	if user.Login == "" {
		return goadmin.ErrRequiredUserLogin
	}

	if govalidator.IsEmail(user.Login) {
		return goadmin.ErrInvalidUserLogin
	}

	if !user.PasswordIsEncoded && user.Password == "" {
		return goadmin.ErrRequiredUserPassword
	}

	if !user.Status.IsValid() {
		return goadmin.ErrInvalidUserStatus
	}

	return nil
}

func (uc *userCase) SearchByLogin(ctx context.Context, login string) (*goadmin.User, error) {
	users, err := uc.Users.Search(ctx, &goadmin.UserFilter{
		Limit:  1,
		Offset: 0,
		Login:  login,
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, goadmin.ErrUserNotFound
	}

	return users[0], nil
}

func (uc *userCase) SearchByID(ctx context.Context, id int64) (*goadmin.User, error) {
	users, err := uc.Users.Search(ctx, &goadmin.UserFilter{
		IDs:    []int64{id},
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, goadmin.ErrUserNotFound
	}

	return users[0], nil
}

func (uc *userCase) SetLastLogged(ctx context.Context, user *goadmin.User) error {
	user.DTLastLogged = time.Now()

	_, err := uc.Users.Update(ctx, *user)
	if err != nil {
		return err
	}

	return nil
}

func (uc *userCase) EncodePassword(user *goadmin.User) error {
	if user.PasswordIsEncoded {
		return nil
	}

	p, err := bcrypt.GenerateFromPassword([]byte(user.Password), 15)
	if err != nil {
		return errors.Wrap(err, "encoding password failed")
	}

	user.PasswordIsEncoded = true
	user.Password = string(p)

	return nil
}

func (uc *userCase) ComparePassword(user *goadmin.User, password string) (bool, error) {
	err := uc.EncodePassword(user)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err = goadmin.ErrWrongPassword
	}

	return err == nil, err

}

func (uc *userCase) CreateAuthToken(ctx context.Context, user *goadmin.User) (*goadmin.Token, error) {
	uniq := []string{
		strconv.FormatInt(user.ID, 10),
		strconv.FormatInt(time.Now().Unix(), 10),
		user.Login, astistring.RandomString(32),
	}

	t := sha256.Sum256([]byte(strings.Join(uniq, "_")))
	token := &goadmin.Token{
		User:      user,
		UserID:    user.ID,
		Type:      goadmin.AuthToken,
		Token:     hex.EncodeToString(t[:]),
		DTExpired: time.Now().Add(24 * time.Hour),
	}

	token, err := uc.Tokens.Create(ctx, *token)
	if err != nil {
		return nil, errors.Wrap(err, "creating token failed")
	}

	return token, nil
}

func (uc *userCase) SearchToken(ctx context.Context, token string) (*goadmin.Token, error) {
	authToken, err := uc.Tokens.Search(ctx, token)
	if err != nil {
		return nil, errors.Wrap(err, "search token failed")
	}

	authToken.User, err = uc.SearchByID(ctx, authToken.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "search user failed")
	}

	if authToken.DTExpired.After(time.Now()) {
		return nil, goadmin.ErrTokenExpired
	}

	return authToken, nil
}

func NewUserCase(users goadmin.UserRepository, tokens goadmin.TokenRepository) *userCase {
	return &userCase{
		Users:  users,
		Tokens: tokens,
	}
}
