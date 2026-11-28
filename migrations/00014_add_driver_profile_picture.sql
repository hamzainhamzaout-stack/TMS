-- +goose Up
-- +goose StatementBegin
ALTER TABLE driver ADD COLUMN profile_picture_url TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE driver DROP COLUMN profile_picture_url;
-- +goose StatementEnd
