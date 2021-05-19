-- +goose Up
DROP TABLE work_phase_comments;

CREATE TABLE project_comments (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	project_id UUID REFERENCES projects NOT NULL,
	author UUID REFERENCES users NOT NULL,
	content TEXT NOT NULL,
	topic TEXT,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE project_comments;

CREATE TABLE work_phase_comments (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	work_phase_id UUID REFERENCES work_phase NOT NULL,
	author UUID REFERENCES users NOT NULL,
	content TEXT NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

