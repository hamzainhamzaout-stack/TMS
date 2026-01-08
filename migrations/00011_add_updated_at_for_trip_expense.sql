-- +goose Up
-- +goose StatementBegin
ALTER TABLE trip_expenses
ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trip_expenses
DROP COLUMN updated_at;
-- +goose StatementEnd
