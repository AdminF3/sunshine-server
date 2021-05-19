-- +goose Up
ALTER TABLE meetings ADD COLUMN internal_project TEXT;
-- +goose Down
ALTER TABLE meetings DROP COLUMN internal_project;
