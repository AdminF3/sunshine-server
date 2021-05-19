-- +goose Up
ALTER TABLE assets ADD COLUMN esco_id UUID REFERENCES organizations ON DELETE CASCADE;

-- +goose Down
ALTER TABLE assets DROP COLUMN esco_id;
