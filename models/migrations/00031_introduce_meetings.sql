-- +goose Up
CREATE TABLE meetings (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	name TEXT NOT NULL,
	host UUID NOT NULL,
	location TEXT,
	date TIMESTAMP WITH TIME ZONE NOT NULL,
	objective TEXT NOT NULL,
	guest TEXT NOT NULL,
	stakeholder NUMERIC NOT NULL,
	participant TEXT[] NOT NULL,
	stage TEXT NOT NULL,
	notes TEXT,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE meetings;
