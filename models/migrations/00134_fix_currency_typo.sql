-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE currency RENAME VALUE 'CSJ' TO 'CZK';

-- +goose Down
ALTER TYPE currency RENAME VALUE 'CZK' TO 'CSJ';
