-- +goose Up
-- updating users
UPDATE users SET status = 0 WHERE status = 4;
UPDATE assets SET status = 0 WHERE status = 4;


-- +goose Down
UPDATE users SET status = 4 WHERE status = 0;
UPDATE assets SET status = 4 WHERE status = 0;
