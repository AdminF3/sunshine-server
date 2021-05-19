-- +goose Up
CREATE TABLE forfaiting_applications (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	project_id UUID REFERENCES projects NOT NULL,
	accepted BOOLEAN DEFAULT FALSE,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE fa_reviews (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	forfaiting_application_id UUID REFERENCES forfaiting_applications NOT NULL,
	author UUID REFERENCES users NOT NULL,
	status transition_request_status,
	comment TEXT,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE fa_reviews;
DROP TABLE forfaiting_applications;
