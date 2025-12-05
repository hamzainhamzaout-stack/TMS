-- +goose Up
-- +goose StatementBegin
ALTER TABLE trip
ADD COLUMN returned_amount NUMERIC DEFAULT 0,
ADD COLUMN item_weight NUMERIC DEFAULT 0,
ADD COLUMN co_driver_name VARCHAR(255) DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trip
DROP COLUMN returned_amount,
DROP COLUMN item_weight,
DROP COLUMN co_driver_name;
-- +goose StatementEnd
