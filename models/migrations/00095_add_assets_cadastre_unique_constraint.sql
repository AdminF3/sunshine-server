-- +goose Up
UPDATE assets SET cadastre=clock_timestamp() WHERE cadastre IS NULL;
ALTER TABLE assets
ALTER COLUMN cadastre SET NOT NULL,
ADD CONSTRAINT assets_cadastre_key UNIQUE (cadastre, country);

-- +goose Down
ALTER TABLE assets
DROP CONSTRAINT assets_cadastre_key,
ALTER COLUMN cadastre DROP NOT NULL;
