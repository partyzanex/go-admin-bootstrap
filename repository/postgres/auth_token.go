package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/partyzanex/layer"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
	"github.com/partyzanex/go-admin-bootstrap/db/models/postgres"
)

type authTokenRepository struct {
	ex layer.BoilExecutor
}

func (repo *authTokenRepository) Search(ctx context.Context, token string) (*goadmin.Token, error) {
	c, ex := layer.GetExecutor(ctx, repo.ex)

	model, err := postgres.AuthTokens(qm.Where("token = ?", token)).One(c, ex)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, goadmin.ErrTokenNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("search token failed: %w", err)
	}

	return modelToToken(model), nil
}

func (repo *authTokenRepository) Create(ctx context.Context, token *goadmin.Token) (result *goadmin.Token, err error) {
	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", layer.ErrCreateTransaction, err)
		}

		defer layer.ExecuteTransaction(tr, &err)
	}

	model := tokenToModel(token)

	err = model.Insert(c, tr, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("inserting token failed: %w", err)
	}

	return modelToToken(model), nil
}

func tokenToModel(token *goadmin.Token) *postgres.AuthToken {
	model := &postgres.AuthToken{
		UserID:    token.UserID,
		Token:     token.Token,
		Type:      string(token.Type),
		DTExpired: token.DTExpired,
		DTCreated: token.DTCreated,
	}

	return model
}

func modelToToken(model *postgres.AuthToken) *goadmin.Token {
	token := &goadmin.Token{
		UserID:    model.UserID,
		Token:     model.Token,
		Type:      goadmin.TokenType(model.Type),
		DTExpired: model.DTExpired,
		DTCreated: model.DTCreated,
		User: &goadmin.User{
			ID: model.UserID,
		},
	}

	return token
}

func (repo *authTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	c, ex := layer.GetExecutor(ctx, repo.ex)

	result, err := postgres.AuthTokens(qm.Where("dt_expired < NOW()")).DeleteAll(c, ex)
	if err != nil {
		return 0, fmt.Errorf("deleting expired tokens failed: %w", err)
	}

	return result, nil
}

func NewTokenRepository(ex layer.BoilExecutor) goadmin.TokenRepository {
	return &authTokenRepository{ex: ex}
}
