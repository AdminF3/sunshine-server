-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE meeting_type ADD VALUE IF NOT EXISTS 'conference';
ALTER TYPE meeting_type ADD VALUE IF NOT EXISTS 'workshop';
ALTER TYPE meeting_type ADD VALUE IF NOT EXISTS 'event';
ALTER TYPE meeting_type ADD VALUE IF NOT EXISTS 'eu_project_meeting';

-- +goose Down
SELECT 1;
