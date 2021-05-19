-- +goose Up
CREATE TYPE gdpr_action AS ENUM (
	'get',
	'delete'
);

CREATE TABLE gdpr_requests (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	user_id UUID,
	action gdpr_action NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE gdpr_requests;
DROP TYPE gdpr_action;
