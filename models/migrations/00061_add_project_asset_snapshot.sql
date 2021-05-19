-- +goose Up
ALTER TABLE projects
	ADD COLUMN asset_owner_id uuid,
	ADD COLUMN asset_esco_id uuid,
	ADD COLUMN asset_area integer,
	ADD COLUMN asset_heated_area integer,
	ADD COLUMN asset_billing_area integer,
	ADD COLUMN asset_flats integer,
	ADD COLUMN asset_floors integer,
	ADD COLUMN asset_stair_cases integer,
	ADD COLUMN asset_building_type integer,
	ADD COLUMN asset_heating_type integer,
	ADD COLUMN asset_cadastre character varying(64);

-- +goose Down
ALTER TABLE projects
	DROP COLUMN asset_owner_id,
	DROP COLUMN asset_esco_id,
	DROP COLUMN asset_area,
	DROP COLUMN asset_heated_area,
	DROP COLUMN asset_billing_area,
	DROP COLUMN asset_flats,
	DROP COLUMN asset_floors,
	DROP COLUMN asset_stair_cases,
	DROP COLUMN asset_building_type,
	DROP COLUMN asset_heating_type,
	DROP COLUMN asset_cadastre;
