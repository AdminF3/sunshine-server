-- +goose Up
ALTER TABLE assets ADD COLUMN country TEXT;
ALTER TABLE organizations ADD COLUMN country TEXT;
ALTER TABLE projects ADD COLUMN country TEXT;

CREATE INDEX assets_country_idx ON assets (country);
CREATE INDEX organizations_country_idx ON organizations (country);
CREATE INDEX projects_country_idx ON projects (country);

-- +goose Down
DROP INDEX assets_country_idx;
DROP INDEX organizations_country_idx;
DROP INDEX projects_country_idx;

ALTER TABLE assets DROP COLUMN country;
ALTER TABLE organizations DROP COLUMN country;
ALTER TABLE projects DROP COLUMN country;
