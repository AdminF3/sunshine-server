-- +goose Up
ALTER TABLE attachments ADD COLUMN comment TEXT;

-- +goose Down
ALTER TABLE attachments DROP COLUMN comment;
