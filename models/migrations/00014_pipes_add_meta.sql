-- +goose Up
ALTER TABLE pipes
	ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- +goose Down
ALTER TABLE pipes
	DROP COLUMN created_at,
	DROP COLUMN updated_at,
	DROP COLUMN deleted_at;
