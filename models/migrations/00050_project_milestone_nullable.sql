-- +goose NO TRANSACTION
-- +goose Up

-- +goose StatementBegin
ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'zero' BEFORE 'acquisition_meeting';
-- +goose StatementEnd

ALTER TABLE projects ALTER COLUMN milestone SET DEFAULT 'zero';

-- +goose Down
ALTER TABLE projects ALTER COLUMN milestone SET DEFAULT 'acquisition_meeting';
