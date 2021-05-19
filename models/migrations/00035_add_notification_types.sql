-- +goose Up
CREATE TYPE entity_type AS ENUM (
	'user',
	'organization',
	'asset',
	'project',
	'indoor_clima',
	'meeting'
);

CREATE TYPE user_action AS ENUM (
	'create',
	'update',
	'upload'
);

ALTER TABLE notifications DROP COLUMN target_type;
ALTER TABLE notifications ADD COLUMN target_type entity_type;

ALTER TABLE notifications DROP COLUMN action;
ALTER TABLE notifications ADD COLUMN action user_action;

-- +goose Down
ALTER TABLE notifications DROP COLUMN target_type;
ALTER TABLE notifications ADD COLUMN target_type TEXT;

ALTER TABLE notifications DROP COLUMN action;
ALTER TABLE notifications ADD COLUMN action TEXT;

DROP TYPE entity_type;
DROP TYPE user_action;
