-- +goose Up
ALTER TABLE projects ADD COLUMN commissioning_date TIMESTAMP WITH TIME ZONE;

-- +goose Down
ALTER TABLE projects DROP COLUMN commissioning_date;
