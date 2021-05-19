-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'forfaiting annual check';

-- +goose Down
SELECT 1;
