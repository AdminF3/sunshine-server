-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'lear apply';

ALTER TYPE user_action RENAME TO old_user_action;
CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign', 'gdpr',
	'request_membership', 'lear_apply', 'claim_residency','request_project_creation',
	'accept_lear_application', 'remove', 'forfaiting_application', 'reject', 'reject_lear_application');
ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;
DROP TYPE old_user_action;

-- +goose Down
UPDATE notifications SET action = 'accept_lear_application' WHERE action = 'reject_lear_application';
ALTER TYPE user_action RENAME TO old_user_action;
CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign', 'gdpr',
	'request_membership','lear_apply', 'claim_residency', 'request_project_creation',
	'accept_lear_application', 'remove', 'forfaiting_application', 'reject');
ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;
DROP TYPE old_user_action;
