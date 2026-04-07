-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_auth_token_user_id ON goadmin.auth_token (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS goadmin.idx_auth_token_user_id;
-- +goose StatementEnd
