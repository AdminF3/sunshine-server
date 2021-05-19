-- +goose Up
ALTER TABLE projects ADD COLUMN milestone milestone NOT NULL DEFAULT 'acquisition_meeting';

-- +goose Down
ALTER TABLE projects DROP COLUMN milestone;
