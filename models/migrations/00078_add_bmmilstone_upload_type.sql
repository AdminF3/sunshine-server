-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'building maintenance milestone';

-- +goose Down
SELECT 1;
