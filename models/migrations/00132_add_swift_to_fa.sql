-- +goose Up
ALTER TABLE forfaiting_applications ADD COLUMN swift TEXT;

-- +goose Down
ALTER TABLE forfaiting_applications DROP COLUMN swift;
