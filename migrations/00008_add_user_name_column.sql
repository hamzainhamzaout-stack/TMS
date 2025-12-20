-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN user_name VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN user_name;
-- +goose StatementEnd
