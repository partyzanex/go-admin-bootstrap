-- +goose Up
-- +goose StatementBegin
ALTER TABLE goadmin."user" ALTER COLUMN password TYPE VARCHAR(255);
ALTER TABLE goadmin.auth_token ALTER COLUMN token TYPE VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE goadmin.auth_token ALTER COLUMN token TYPE CHARACTER(64);
ALTER TABLE goadmin."user" ALTER COLUMN password TYPE CHARACTER(64);
-- +goose StatementEnd
