-- +goose Up
UPDATE users SET country = 'Latvia' WHERE country IS NULL OR country = '';
ALTER TABLE users ALTER COLUMN country SET NOT NULL;

UPDATE assets SET country = 'Latvia' WHERE country IS NULL OR country = '';
ALTER TABLE assets ALTER COLUMN country SET NOT NULL;

UPDATE organizations SET country = 'Latvia' WHERE country IS NULL OR country = '';
ALTER TABLE organizations ALTER COLUMN country SET NOT NULL;

UPDATE projects SET country = 'Latvia' WHERE country IS NULL OR country = '';
ALTER TABLE projects ALTER COLUMN country SET NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN country DROP NOT NULL;
ALTER TABLE assets ALTER COLUMN country DROP NOT NULL;
ALTER TABLE organizations ALTER COLUMN country DROP NOT NULL;
ALTER TABLE projects ALTER COLUMN country DROP NOT NULL;
