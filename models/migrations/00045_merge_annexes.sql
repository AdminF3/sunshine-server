-- +goose Up
ALTER TABLE contracts
	DROP COLUMN annex1,
	DROP COLUMN annex2,
	DROP COLUMN annex3,
	DROP COLUMN annex4,
	DROP COLUMN annex5,
	ADD COLUMN tables JSONB;

-- +goose Down
ALTER TABLE contracts
	ADD COLUMN annex1 JSONB,
	ADD COLUMN annex2 JSONB,
	ADD COLUMN annex3 JSONB,
	ADD COLUMN annex4 JSONB,
	ADD COLUMN annex5 JSONB,
	DROP COLUMN tables;
