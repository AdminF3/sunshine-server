-- +goose Up
CREATE TYPE asset_category AS ENUM (
	'nonresidential_educational_facilities',
	'nonresidential_cultural_facilities',
	'nonresidential_medical_facilities',
	'nonresidential_sports_facilities',
	'nonresidential_office_buildings',
	'nonresidential_transportation_facilities');

ALTER TABLE assets ADD COLUMN category asset_category;

-- +goose Down
ALTER TABLE assets DROP COLUMN category;

DROP TYPE asset_category;
