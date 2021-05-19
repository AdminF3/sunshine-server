-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'commissioning';

-- +goose Down
-- adding too many new types for a sane revert.
SELECT 1;
