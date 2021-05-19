-- +goose Up
CREATE TABLE portfolio_directors (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	country TEXT NOT NULL,
	user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE (country, user_id)
);

-- +goose Down
DROP TABLE portfolio_directors;
