-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE portfolio_roles ADD VALUE IF NOT EXISTS 'investor';

-- +goose Down
SELECT 1;
