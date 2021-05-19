-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE meeting_type ADD VALUE IF NOT EXISTS 'participation_event';

-- +goose Down
SELECT 1;
