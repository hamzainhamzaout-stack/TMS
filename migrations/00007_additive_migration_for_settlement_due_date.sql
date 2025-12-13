-- +goose Up
-- +goose StatementBegin
ALTER TABLE trip ADD COLUMN settlement_due_date DATE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trip DROP COLUMN settlement_due_date;
-- +goose StatementEnd
