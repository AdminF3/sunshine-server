-- +goose Up
ALTER TABLE fa_reviews ALTER COLUMN author DROP NOT NULL;

-- +goose Down
ALTER TABLE fa_reviews ALTER COLUMN author SET NOT NULL;
