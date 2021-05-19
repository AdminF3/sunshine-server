-- +goose Up
ALTER TABLE assets RENAME COLUMN owner TO owner_id;
ALTER TABLE attachments RENAME COLUMN owner TO owner_id;
ALTER TABLE contracts RENAME COLUMN project TO project_id;
ALTER TABLE indoor_climas RENAME COLUMN project TO project_id;
ALTER TABLE pipes RENAME COLUMN indoorclima TO indoorclima_id;


-- +goose Down
ALTER TABLE pipes RENAME COLUMN indoorclima_id TO indoorclima;
ALTER TABLE indoor_climas RENAME COLUMN project_id TO project;
ALTER TABLE contracts RENAME COLUMN project_id TO project;
ALTER TABLE attachments RENAME COLUMN owner_id TO owner;
ALTER TABLE assets RENAME COLUMN owner_id TO owner;
