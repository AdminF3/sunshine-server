-- +goose Up
ALTER TABLE meetings ADD COLUMN project UUID REFERENCES projects;

-- +goose Down
ALTER TABLE meetings DROP COLUMN project;

