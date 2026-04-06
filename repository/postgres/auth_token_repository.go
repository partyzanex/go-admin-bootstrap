package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

type authTokenRepository struct {
	db *bun.DB
}

func NewTokenRepository(db *bun.DB) goadmin.TokenRepository {
	return &authTokenRepository{db: db}
}

func (repo *authTokenRepository) Search(ctx context.Context, token string) (*goadmin.Token, error) {
	model := new(tokenModel)

	err := repo.db.NewSelect().Model(model).Where("t.token = ?", token).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, goadmin.NewTokenNotFoundError(token)
	}

	if err != nil {
		return nil, fmt.Errorf("search token: %w", err)
	}

	return modelToToken(model), nil
}

func (repo *authTokenRepository) Create(ctx context.Context, token *goadmin.Token) (*goadmin.Token, error) {
	model := tokenToModel(token)

	_, err := repo.db.NewInsert().Model(model).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("insert token: %w", err)
	}

	return modelToToken(model), nil
}

func (repo *authTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := repo.db.NewDelete().
		Model((*tokenModel)(nil)).
		Where("dt_expired < NOW()").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete expired tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete expired rows affected: %w", err)
	}

	return rows, nil
}

func (repo *authTokenRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := repo.db.NewDelete().
		Model((*tokenModel)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete tokens by user id: %w", err)
	}

	return nil
}
