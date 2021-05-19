-- +goose Up
ALTER TABLE organizations DROP COLUMN lear_name;

-- +goose Down
ALTER TABLE organizations ADD COLUMN lear_name CHARACTER VARYING(255);
