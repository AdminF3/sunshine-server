-- +goose Up
ALTER TYPE meeting_type RENAME TO old_meeting_type;

CREATE TYPE meeting_type AS ENUM (
	'organization_conference',
	'organization_workshop',
	'press_release',
	'popularised_publication',
	'exhibition',
	'training',
	'communication_campaign',
	'participation_conference',
	'participation_workshop',
	'brokerage_event',
	'pitch_event',
	'trade_fair',
	'eu_project_activity',
	'other',
	'acquisition',
	'acquisition_commitment',
	'acquisition_kick_off',
	'works_kick_off',
	'works_initial_information',
	'works_weekly_report',
	'works_renovation_informative',
	'works_communication',
	'works_construction_managers_final',
	'works_final_information');

ALTER TABLE meetings ALTER COLUMN  topic TYPE meeting_type USING topic::TEXT::meeting_type;
DROP TYPE old_meeting_type;

-- +goose Down
UPDATE meetings SET topic = 'training' WHERE topic = 'acquisition';
UPDATE meetings SET topic = 'training' WHERE topic = 'acquisition_commitment';
UPDATE meetings SET topic = 'training' WHERE topic = 'acquisition_kick_off';
UPDATE meetings SET topic = 'training' WHERE topic = 'works_kick_off';
UPDATE meetings SET topic = 'training' WHERE topic = 'works_initial_information';
UPDATE meetings SET topic = 'training' WHERE topic = 'works_weekly_report';
UPDATE meetings SET topic = 'training' WHERE topic = 'works_renovation_informative';
UPDATE meetings SET topic = 'training' WHERE topic = 'works_communication';
UPDATE meetings SET topic = 'training' WHERE topic = 'works_construction_managers_final';
UPDATE meetings SET topic = 'training' WHERE topic = 'works_final_information';

ALTER TYPE meeting_type RENAME TO old_meeting_type;
CREATE TYPE meeting_type AS ENUM (
	'organization_conference',
	'organization_workshop',
	'press_release',
	'popularised_publication',
	'exhibition',
	'training',
	'communication_campaign',
	'participation_conference',
	'participation_workshop',
	'video',
	'brokerage_event',
	'pitch_event',
	'trade_fair',
	'eu_project_activity',
	'other');
ALTER TABLE meetings ALTER COLUMN  topic TYPE meeting_type USING topic::TEXT::meeting_type;
DROP TYPE old_meeting_type;
