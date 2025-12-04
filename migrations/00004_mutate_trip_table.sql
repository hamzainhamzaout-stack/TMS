-- +goose Up
-- +goose StatementBegin
ALTER TABLE trip
ADD COLUMN origin VARCHAR(255),
ADD COLUMN destination VARCHAR(255),
ADD COLUMN departure_time TIMESTAMP,
ADD COLUMN arrival_time TIMESTAMP,
ADD COLUMN items TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trip
DROP COLUMN origin,
DROP COLUMN destination,
DROP COLUMN departure_time,
DROP COLUMN arrival_time,
DROP COLUMN item_type;

-- +goose StatementEnd
