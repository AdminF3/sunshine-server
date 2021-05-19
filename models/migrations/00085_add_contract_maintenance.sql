-- +goose Up
ALTER TABLE contracts ADD COLUMN maintenance JSONB;

-- +goose Down
ALTER TABLE contracts DROP COLUMN maintenance;
