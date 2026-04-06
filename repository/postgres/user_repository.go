package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

type userRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) goadmin.UserRepository {
	return &userRepository{db: db}
}

func (repo *userRepository) Search(ctx context.Context, filter *goadmin.UserFilter) ([]*goadmin.User, error) {
	var models []userModel

	q := repo.db.NewSelect().Model(&models).OrderExpr("u.id ASC")
	q = applyUserFilter(q, filter)

	err := q.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}

	return mapSlice(models, modelToUser), nil
}

func (repo *userRepository) Count(ctx context.Context, filter *goadmin.UserFilter) (int64, error) {
	q := repo.db.NewSelect().Model((*userModel)(nil))
	q = applyUserFilter(q, filter)

	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return int64(count), nil
}

func applyUserFilter(q *bun.SelectQuery, f *goadmin.UserFilter) *bun.SelectQuery {
	if f == nil {
		return q
	}

	if len(f.IDs) > 0 {
		q = q.Where("u.id IN (?)", bun.List(f.IDs))
	}

	if f.Login != "" {
		q = q.Where("u.login = ?", f.Login)
	}

	if f.Name != "" {
		q = q.Where("u.name ILIKE ?", "%"+f.Name+"%")
	}

	if f.Status != "" {
		q = q.Where("u.status = ?", f.Status)
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset(f.Offset)
	}

	return q
}

func (repo *userRepository) Create(ctx context.Context, user *goadmin.User) (*goadmin.User, error) {
	model := userToModel(user)
	model.DTCreated = time.Now().UTC()
	model.DTUpdated = time.Now().UTC()

	_, err := repo.db.NewInsert().Model(model).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return modelToUser(model), nil
}

func (repo *userRepository) Update(ctx context.Context, user *goadmin.User) (*goadmin.User, error) {
	model := userToModel(user)
	model.DTUpdated = time.Now().UTC()

	_, err := repo.db.NewUpdate().Model(model).WherePK().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return modelToUser(model), nil
}

func (repo *userRepository) SetLastLogged(ctx context.Context, user *goadmin.User) error {
	model := userToModel(user)

	_, err := repo.db.NewUpdate().
		Model(model).
		Set("dt_last_logged = ?", time.Now().UTC()).
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("set last logged: %w", err)
	}

	return nil
}

func (repo *userRepository) Delete(ctx context.Context, user *goadmin.User) error {
	if user.ID == 0 {
		return goadmin.ErrRequiredUserID
	}

	result, err := repo.db.NewDelete().
		Model((*userModel)(nil)).
		Where("id = ?", user.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user rows affected: %w", err)
	}

	if rows == 0 {
		return goadmin.NewUserNotFoundError(user.ID)
	}

	return nil
}

func GetUserByID(ctx context.Context, db *bun.DB, id int64) (*goadmin.User, error) {
	model := new(userModel)

	err := db.NewSelect().Model(model).Where("u.id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, goadmin.NewUserNotFoundError(id)
	}

	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return modelToUser(model), nil
}
