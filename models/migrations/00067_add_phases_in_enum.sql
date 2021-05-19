
-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'work_phase';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'monitoring_phase';

-- +goose Down
-- adding too many new types for a sane revert.
SELECT 1;
