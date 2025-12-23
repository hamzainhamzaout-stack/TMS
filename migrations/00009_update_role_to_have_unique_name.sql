-- +goose Up
-- +goose StatementBegin
ALTER TABLE role ADD CONSTRAINT unique_role_name UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE role DROP CONSTRAINT unique_role_name;
-- +goose StatementEnd
