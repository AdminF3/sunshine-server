-- +goose Up
ALTER TABLE attachments DROP COLUMN upload_type;

DROP TYPE upload_type;

CREATE TYPE upload_type AS ENUM (
	'general leaflet',
	'process leaflet',
	'aquisition protocol meeting',
	'aquisition protocol survey',
	'energy audit report',
	'technical inspection report',
	'esco tender',
	'commitment protocol meeting',
	'commitment protocol survey',
	'pre epc agreement',
	'cooperation agreement',
	'construction project',
	'procurement construction installation',
	'finincing offer and altum',
	'draft epc contract',
	'kickoff protocol meeting',
	'kickoff protocol survey',
	'signed epc',
	'agreement altum bank loan',
	'agreement altum grand agreement',
	'project management contract',
	'construction works company contract',
	'engineering networks company contract',
	'supervision contract',
	'maintenance company contract',
	'land owners contract',
	'house heating contract',
	'windows contract',
	'kick-off meeting document',
	'initial information meeting document',
	'weekly meeting report document',
	'informative residents meeting document',
	'other residents meeting document',
	'monthly construction company report',
	'expenses document',
	'tama comments document',
	'building audit document',
	'work acceptance document',
	'construction managers final meeting mom',
	'residents final meeting mom',
	'building user guide'
);

ALTER TABLE attachments ADD COLUMN upload_type upload_type;

-- +goose Down
-- taken from 00048 migration
ALTER TABLE attachments DROP COLUMN upload_type;

DROP TYPE upload_type;

CREATE TYPE upload_type AS ENUM (
	'general leaflet',
	'process leaflet',
	'aquisition protocol meeting',
	'aquisition protocol survey',
	'energy audit report',
	'technical inspection report',
	'esco tender',
	'commitment protocol meeting',
	'commitment protocol survey',
	'pre epc agreement',
	'cooperation agreement',
	'construction project',
	'procurement construction installation',
	'finincing offer and altum',
	'draft epc contract',
	'kickoff protocol meeting',
	'kickoff protocol survey',
	'signed epc',
	'agreement altum bank loan',
	'agreement altum grand agreement'
);

ALTER TABLE attachments ADD COLUMN upload_type upload_type;

