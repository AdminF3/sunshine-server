-- +goose Up
ALTER TABLE users ADD COLUMN country TEXT;
CREATE INDEX users_country_idx ON users (country);

-- +goose Down
ALTER TABLE users DROP COLUMN country;
