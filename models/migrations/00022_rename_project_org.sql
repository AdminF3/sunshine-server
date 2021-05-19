-- +goose Up
ALTER TABLE projects RENAME COLUMN organization TO owner;

-- +goose Down
ALTER TABLE projects RENAME COLUMN owner TO organization;

