-- +goose Up
-- +goose StatementBegin
ALTER TABLE trip_expenses
ADD COLUMN receipt_url TEXT;

ALTER TABLE trip_expenses
ADD COLUMN remark TEXT;

ALTER TABLE trip_expenses
ADD COLUMN cost NUMERIC NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE trip_expenses
DROP COLUMN IF EXISTS receipt_url;

ALTER TABLE trip_expenses
DROP COLUMN IF EXISTS remark;

ALTER TABLE trip_expenses
DROP COLUMN IF EXISTS cost;
-- +goose StatementEnd
