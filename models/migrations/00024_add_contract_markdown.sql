-- +goose Up
ALTER TABLE contracts ADD COLUMN markdown TEXT;

-- +goose Down
ALTER TABLE contracts DROP COLUMN markdown;
