-- +goose Up
UPDATE organizations SET status = status + 1;
UPDATE assets SET status = status + 1;
UPDATE users SET status = status + 1;

-- +goose Down
UPDATE organizations SET status = status - 1;
UPDATE assets SET status = status - 1;
UPDATE users SET status = status - 1;
