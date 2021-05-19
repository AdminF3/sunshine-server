-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE user_action ADD VALUE IF NOT EXISTS 'claim_residency';

-- +goose Down

SELECT 1;
