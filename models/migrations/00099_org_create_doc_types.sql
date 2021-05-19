-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'registration document';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'proof of address';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'vat document';

-- +goose Down
SELECT 1;
