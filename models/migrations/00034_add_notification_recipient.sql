-- +goose Up
ALTER TABLE notifications ADD COLUMN recipient UUID REFERENCES users;

-- +goose Down
ALTER TABLE notifications DROP COLUMN recipient;
