-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'payment of loans';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'investment invoices';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'annual check other financials';

-- +goose Down
SELECT 1;