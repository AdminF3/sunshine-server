-- +goose Up
CREATE TABLE notifications (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	action TEXT,
	user_id UUID,
	user_key TEXT,
	target_id UUID,
	target_key TEXT,
	target_type TEXT,
	old TEXT,
	new TEXT,
	seen BOOLEAN,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE notifications;
