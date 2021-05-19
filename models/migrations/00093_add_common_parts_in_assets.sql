-- +goose Up
ALTER TABLE assets ADD COLUMN common_parts_area INTEGER;

-- +goose Down
ALTER TABLE assets DROP COLUMN common_parts_area;
