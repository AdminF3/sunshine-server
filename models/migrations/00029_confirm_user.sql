-- +goose Up
ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE users SET is_active = TRUE;

-- +goose Down
ALTER TABLE users DROP COLUMN is_active;
