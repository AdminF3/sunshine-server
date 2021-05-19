-- +goose Up
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

ALTER TABLE meetings ADD COLUMN topic meeting_type;

-- +goose Down
ALTER TABLE meetings DROP COLUMN topic;

DROP TYPE meeting_type;
