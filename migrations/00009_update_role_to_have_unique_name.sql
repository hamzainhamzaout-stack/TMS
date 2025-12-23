-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles ADD CONSTRAINT unique_role_name UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles DROP CONSTRAINT unique_role_name;
-- +goose StatementEnd
