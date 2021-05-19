-- +goose Up
ALTER TABLE projects
      ADD CONSTRAINT organizations_owner_fkey
      FOREIGN KEY (owner)
      REFERENCES organizations(id);

-- +goose Down
ALTER TABLE projects DROP CONSTRAINT organizations_owner_fkey;
