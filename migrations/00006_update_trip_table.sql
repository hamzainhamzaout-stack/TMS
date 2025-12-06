-- +goose Up
-- +goose StatementBegin
ALTER TABLE trip ADD COLUMN client_phone_number VARCHAR(20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trip DROP COLUMN client_phone_number;
-- +goose StatementEnd
