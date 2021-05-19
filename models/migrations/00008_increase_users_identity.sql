-- +goose Up
ALTER TABLE users ALTER COLUMN identity TYPE CHARACTER VARYING(255);
ALTER TABLE assets
	DROP CONSTRAINT assets_cadastre_key,
	ALTER COLUMN cadastre DROP NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN identity TYPE CHARACTER VARYING(64);
ALTER TABLE assets
	ALTER COLUMN cadastre SET NOT NULL,
	ADD CONSTRAINT assets_cadastre_key UNIQUE (cadastre);
