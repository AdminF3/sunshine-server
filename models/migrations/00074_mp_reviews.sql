-- +goose Up
CREATE TABLE mp_reviews (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	mp_id UUID REFERENCES monitoring_phase NOT NULL,
	author UUID REFERENCES users,
	approved BOOLEAN,
	comment TEXT,
	type INTEGER,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

DROP TABLE task_comments;

DROP TABLE tasks;

-- +goose Down
DROP TABLE mp_reviews;

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

