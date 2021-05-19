-- +goose Up
CREATE TABLE attachments (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),

	name CHARACTER VARYING(255) NOT NULL,
	owner UUID NOT NULL,
	content_type CHARACTER VARYING(255),
	size integer NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE attachments;
