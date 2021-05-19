-- +goose Up
-- +goose NO TRANSACTION
-- this is due to forgotten enum value from previous migrations.
ALTER TYPE meeting_type ADD VALUE IF NOT EXISTS 'internal_meeting';

UPDATE meetings SET topic = 'internal_meeting' WHERE topic in (
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
	'other',
	'participation_event'
);

ALTER TYPE meeting_type RENAME TO meeting_type_old;
CREATE TYPE meeting_type AS ENUM (
	'internal_meeting',
	'conference',
	'workshop',
	'event',
	'training',
	'eu_project_meeting',
	'eu_project_activity',

	'acquisition',
	'acquisition_commitment',
	'acquisition_kick_off',
	'works_kick_off',
	'works_initial_information',
	'works_weekly_report',
	'works_renovation_informative',
	'works_communication',
	'works_construction_managers_final',
	'works_final_information',
        'other'
);

ALTER TABLE meetings ALTER COLUMN topic TYPE meeting_type USING topic::TEXT::meeting_type;
DROP TYPE meeting_type_old;

-- +goose Down
ALTER TYPE meeting_type RENAME TO meeting_type_old;
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
	'participation_event',
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
	'works_final_information',
	'internal_meeting',
	'conference',
	'workshop',
	'event',
	'eu_project_meeting'
        'other',
);

ALTER TABLE meetings ALTER COLUMN topic TYPE meeting_type USING topic::TEXT::meeting_type;
DROP TYPE meeting_type_old;
