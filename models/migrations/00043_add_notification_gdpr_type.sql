-- +goose Up
ALTER TYPE user_action RENAME TO old_user_action;

CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign', 'gdpr');

ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;

DROP TYPE old_user_action;

-- +goose Down
UPDATE notifications SET action = 'create' WHERE action = 'gdpr';

ALTER TYPE user_action RENAME TO old_user_action;

CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign');

ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;

DROP TYPE old_user_action;
