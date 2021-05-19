-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'energy management system company';

-- +goose Down
SELECT 1;
