package postgres

import (
	"context"
	"database/sql"

	"github.com/partyzanex/go-admin-bootstrap/db/models/postgres"
	"github.com/partyzanex/layer"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	goadmin "github.com/partyzanex/go-admin-bootstrap"
)

type authTokenRepository struct {
	ex layer.BoilExecutor
}

func (repo *authTokenRepository) Search(ctx context.Context, token string) (*goadmin.Token, error) {
	c, ex := layer.GetExecutor(ctx, repo.ex)

	model, err := postgres.AuthTokens(qm.Where("token = ?", token)).One(c, ex)
	if err == sql.ErrNoRows {
		return nil, goadmin.ErrTokenNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "search token failed")
	}

	return modelToToken(model), nil
}

func (repo *authTokenRepository) Create(ctx context.Context, token goadmin.Token) (result *goadmin.Token, err error) {
	c, tr := layer.GetTransactor(ctx)
	if tr == nil {
		tr, err = repo.ex.BeginTx(ctx, nil)
		if err != nil {
			return nil, errors.Wrap(err, layer.ErrCreateTransaction.Error())
		}

		defer func() {
			errTr := layer.ExecuteTransaction(tr, err)
			if errTr != nil {
				err = errors.Wrap(errTr, "transaction error")
			}
		}()
	}

	model := tokenToModel(&token)

	err = model.Insert(c, tr, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "inserting token failed")
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

func NewTokenRepository(ex layer.BoilExecutor) *authTokenRepository {
	return &authTokenRepository{ex: ex}
}
