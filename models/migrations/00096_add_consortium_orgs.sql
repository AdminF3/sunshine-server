-- +goose Up
ALTER TABLE projects ADD COLUMN consortium_orgs TEXT[];

-- +goose Down
ALTER TABLE projects DROP COLUMN consortium_orgs;
