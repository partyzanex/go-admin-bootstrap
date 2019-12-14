package repository

import (
	"context"
	"database/sql"
	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/db/models/postgres"
	"github.com/partyzanex/layer"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type userRepository struct {
	ex layer.BoilExecutor
}

func NewUserRepository(ex layer.BoilExecutor) *userRepository {
	return &userRepository{ex: ex}
}

func (repo *userRepository) Search(ctx context.Context, filter *goadmin.UserFilter) ([]*goadmin.User, error) {
	mods := repo.applyFilter(filter, nil)

	c, ex := layer.GetExecutor(ctx, repo.ex)

	models, err := postgres.Users(mods...).All(c, ex)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "search users failed")
	}

	users := make([]*goadmin.User, len(models))

	for i, model := range models {
		users[i] = modelToUser(model)
	}

	return users, nil
}

func (repo *userRepository) Count(ctx context.Context, filter *goadmin.UserFilter) (int64, error) {
	var mods []qm.QueryMod

	if filter != nil {
		f := *filter
		f.Limit = 0
		mods = repo.applyFilter(&f, mods)
	}

	c, ex := layer.GetExecutor(ctx, repo.ex)

	count, err := postgres.Users(mods...).Count(c, ex)
	if err != nil {
		return 0, errors.Wrap(err, "getting count of users failed")
	}

	return count, nil
}

func (*userRepository) applyFilter(filter *goadmin.UserFilter, mods []qm.QueryMod) []qm.QueryMod {
	if mods == nil {
		mods = []qm.QueryMod{}
	}

	if filter != nil {
		if filter.Name != "" {
			clause := "%" + filter.Name + "%"
			mods = append(mods, qm.Where("name like ?", clause))
		}

		if filter.Status != "" {
			mods = append(mods, qm.Where("status = ?", filter.Status))
		}

		if filter.Login != "" {
			mods = append(mods, qm.Where("login = ?", filter.Login))
		}

		if filter.Limit > 0 {
			mods = append(mods, qm.Limit(filter.Limit))

			if filter.Offset >= 0 {
				mods = append(mods, qm.Offset(filter.Offset))
			}
		}
	}

	return mods
}

func (repo *userRepository) Create(ctx context.Context, user goadmin.User) (result *goadmin.User, err error) {
	if err := user.Validate(true); err != nil {
		return nil, errors.Wrap(err, "validation failed")
	}

	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return nil, errors.Wrap(err, layer.ErrCreateTransaction.Error())
		}

		defer func() {
			err = layer.ExecuteTransaction(tr, err)
		}()
	}

	model := userToModel(&user)

	err = model.Insert(c, tr, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "inserting user failed")
	}

	return modelToUser(model), nil
}

func (repo *userRepository) Update(ctx context.Context, user goadmin.User) (result *goadmin.User, err error) {
	if err := user.Validate(false); err != nil {
		return nil, errors.Wrap(err, "validation failed")
	}

	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return nil, errors.Wrap(err, layer.ErrCreateTransaction.Error())
		}

		defer func() {
			err = layer.ExecuteTransaction(tr, err)
		}()
	}

	model := userToModel(&user)

	_, err = model.Update(c, tr, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "updating user failed")
	}

	return modelToUser(model), err
}

func (repo *userRepository) Delete(ctx context.Context, user goadmin.User) (err error) {
	if user.ID == 0 {
		return goadmin.ErrRequiredUserID
	}

	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return errors.Wrap(err, layer.ErrCreateTransaction.Error())
		}

		defer func() {
			err = layer.ExecuteTransaction(tr, err)
		}()
	}

	c, ex := layer.GetExecutor(ctx, repo.ex)

	model, err := postgres.Users(qm.Where("id = ?", user.ID)).One(c, tr)
	if err == sql.ErrNoRows {
		return goadmin.ErrUserNotFound
	}
	if err != nil {
		return errors.Wrap(err, "search user failed")
	}

	_, err = model.Delete(c, ex)
	if err != nil {
		return errors.Wrap(err, "deleting user failed")
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
		DTCreated:    user.DTCreated,
		DTUpdated:    user.DTUpdated,
		DTLastLogged: user.DTLastLogged,
	}

	return model
}
