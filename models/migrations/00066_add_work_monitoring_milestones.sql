-- +goose Up
CREATE TABLE work_phase (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	project_id UUID REFERENCES projects NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE work_phase_comments (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	work_phase_id UUID REFERENCES work_phase NOT NULL,
	author UUID REFERENCES users NOT NULL,
	content TEXT NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE monitoring_phase (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	project_id UUID REFERENCES projects NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE tasks (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	monitoring_phase_id UUID REFERENCES monitoring_phase NOT NULL,
	activity TEXT,
	company UUID REFERENCES organizations,
	planned_date TIMESTAMP WITH TIME ZONE,
	date_done TIMESTAMP WITH TIME ZONE,
	status BOOLEAN,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE task_comments (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	task_id UUID REFERENCES tasks NOT NULL,
	author UUID REFERENCES users NOT NULL,
	content TEXT NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE task_comments;
DROP TABLE work_phase_comments;
DROP TABLE tasks;
DROP TABLE work_phase;
DROP TABLE monitoring_phase;
