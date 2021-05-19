-- +goose Up
ALTER TABLE notifications ADD COLUMN country country;

-- +goose Down
ALTER TABLE notifications DROP COLUMN country;
