-- +goose Up
ALTER TYPE user_action RENAME TO old_user_action;

CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign', 'gdpr', 'request_membership');

ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;

DROP TYPE old_user_action;

-- +goose Down
UPDATE notifications SET action = 'assign' WHERE action = 'request_membership';

ALTER TYPE user_action RENAME TO old_user_action;

CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign', 'gdpr');

ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;

DROP TYPE old_user_action;
