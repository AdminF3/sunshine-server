-- +goose Up
CREATE TYPE milestone AS ENUM (
	'acquisition_meeting',
	'feasibility_study',
	'commitment_study',
	'project_design',
	'project_preparation',
	'kick_off_meeting'
);

CREATE TYPE transition_request_status AS ENUM (
	'for_review',
	'accepted',
	'rejected',
	'requested_changes'
);


CREATE TABLE transition_requests (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	project_id UUID REFERENCES projects NOT NULL,
	from_milestone milestone NOT NULL,
	to_milestone milestone NOT NULL,
	status transition_request_status NOT NULL,


	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE transition_request_comments (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	transition_request_id UUID REFERENCES transition_requests NOT NULL,
	author UUID REFERENCES users NOT NULL,
	content TEXT NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE transition_request_comments;
DROP TABLE transition_requests;
DROP TYPE transition_request_status;
DROP TYPE milestone;
