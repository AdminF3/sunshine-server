-- +goose Up
ALTER TYPE user_action RENAME TO old_user_action;
CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign', 'gdpr',
	'request_membership', 'lear_apply', 'claim_residency','request_project_creation',
	'accept_lear_application', 'remove', 'forfaiting_application', 'reject',
	'approve_forfaiting_application', 'approve_forfaiting_payment');
ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;
DROP TYPE old_user_action;

-- +goose Down
UPDATE notifications SET action = 'forfaiting_application' WHERE action = 'approve_forfaiting_application';
UPDATE notifications SET action = 'create' WHERE action = 'approve_forfaiting_payment';

ALTER TYPE user_action RENAME TO old_user_action;
CREATE TYPE user_action AS ENUM ('create', 'update', 'upload', 'assign', 'gdpr',
	'request_membership','lear_apply', 'claim_residency', 'request_project_creation',
	'accept_lear_application', 'remove', 'forfaiting_application', 'reject');
ALTER TABLE notifications ALTER COLUMN action TYPE user_action USING action::TEXT::user_action;
DROP TYPE old_user_action;
