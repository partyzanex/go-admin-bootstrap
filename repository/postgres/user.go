package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/partyzanex/layer"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/db/models/postgres"
)

type userRepository struct {
	ex layer.BoilExecutor
}

func NewUserRepository(ex layer.BoilExecutor) goadmin.UserRepository {
	return &userRepository{ex: ex}
}

func (repo *userRepository) Search(ctx context.Context, filter *goadmin.UserFilter) ([]*goadmin.User, error) {
	mods := userSearchQuery(filter)

	c, ex := layer.GetExecutor(ctx, repo.ex)

	models, err := postgres.Users(mods...).All(c, ex)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("search users failed: %w", err)
	}

	return mapSlice(models, modelToUser), nil
}

func (repo *userRepository) Count(ctx context.Context, filter *goadmin.UserFilter) (int64, error) {
	mods := userFilterMods(filter)

	c, ex := layer.GetExecutor(ctx, repo.ex)

	count, err := postgres.Users(mods...).Count(c, ex)
	if err != nil {
		return 0, fmt.Errorf("getting count of users failed: %w", err)
	}

	return count, nil
}

func userFilterMods(filter *goadmin.UserFilter) []qm.QueryMod {
	if filter == nil {
		return nil
	}

	ids := make([]any, len(filter.IDs))
	for i, id := range filter.IDs {
		ids[i] = id
	}

	return buildQuery(
		nil,
		withWhereIn("id", ids),
		withWhereLike("name", filter.Name),
		withWhereEq("status", string(filter.Status)),
		withWhereEq("login", filter.Login),
	)
}

func userSearchQuery(filter *goadmin.UserFilter) []qm.QueryMod {
	mods := append(userFilterMods(filter), qm.OrderBy("id"))

	if filter != nil && filter.Limit > 0 {
		mods = buildQuery(mods, withLimit(filter.Limit), withOffset(filter.Offset))
	}

	return mods
}

func (repo *userRepository) Create(ctx context.Context, user *goadmin.User) (result *goadmin.User, err error) {
	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", layer.ErrCreateTransaction, err)
		}

		defer layer.ExecuteTransaction(tr, &err)
	}

	model := userToModel(user)
	model.DTCreated = time.Now().UTC()

	err = model.Insert(c, tr, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("inserting user failed: %w", err)
	}

	return modelToUser(model), nil
}

func (repo *userRepository) Update(ctx context.Context, user *goadmin.User) (result *goadmin.User, err error) {
	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", layer.ErrCreateTransaction, err)
		}

		defer layer.ExecuteTransaction(tr, &err)
	}

	model := userToModel(user)
	model.DTUpdated = time.Now().UTC()

	_, err = model.Update(c, tr, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("updating user failed: %w", err)
	}

	return modelToUser(model), err
}

func (repo *userRepository) SetLastLogged(ctx context.Context, user *goadmin.User) (err error) {
	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("%s: %w", layer.ErrCreateTransaction, err)
		}

		defer layer.ExecuteTransaction(tr, &err)
	}

	model := userToModel(user)
	model.DTLastLogged = time.Now().UTC()

	_, err = model.Update(c, tr, boil.Whitelist(postgres.UserColumns.DTLastLogged))
	if err != nil {
		return fmt.Errorf("updating user failed: %w", err)
	}

	return err
}

func (repo *userRepository) Delete(ctx context.Context, user *goadmin.User) (err error) {
	if user.ID == 0 {
		return goadmin.ErrRequiredUserID
	}

	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("%s: %w", layer.ErrCreateTransaction, err)
		}

		defer layer.ExecuteTransaction(tr, &err)
	}

	model, err := postgres.Users(qm.Where("id = ?", user.ID)).One(c, tr)
	if errors.Is(err, sql.ErrNoRows) {
		return goadmin.ErrUserNotFound
	}

	if err != nil {
		return fmt.Errorf("search user failed: %w", err)
	}

	_, err = model.Delete(c, tr)
	if err != nil {
		return fmt.Errorf("deleting user failed: %w", err)
	}

	return
}

func modelToUser(model *postgres.User) *goadmin.User {
	user := &goadmin.User{
		ID:                model.ID,
		Login:             model.Login,
		Password:          model.Password,
		Status:            goadmin.UserStatus(model.Status),
		Name:              model.Name,
		Role:              goadmin.UserRole(model.Role),
		DTCreated:         model.DTCreated,
		DTUpdated:         model.DTUpdated,
		DTLastLogged:      model.DTLastLogged,
		PasswordIsEncoded: true,
		Current:           false,
	}

	return user
}

func userToModel(user *goadmin.User) *postgres.User {
	model := &postgres.User{
		ID:           user.ID,
		Login:        user.Login,
		Password:     user.Password,
		Status:       string(user.Status),
		Name:         user.Name,
		Role:         string(user.Role),
		DTCreated:    user.DTCreated,
		DTUpdated:    user.DTUpdated,
		DTLastLogged: user.DTLastLogged,
	}

	return model
}
