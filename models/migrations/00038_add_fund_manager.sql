-- +goose Up
ALTER TABLE projects ADD COLUMN fund_manager UUID REFERENCES users;
-- +goose Down
ALTER TABLE projects DROP COLUMN fund_manager;