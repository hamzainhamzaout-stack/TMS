-- +goose Up
-- +goose StatementBegin
ALTER TABLE trip_expenses ADD COLUMN transaction_date DATE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trip_expenses DROP COLUMN transaction_date;
-- +goose StatementEnd
