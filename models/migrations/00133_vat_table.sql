-- +goose Up

CREATE TABLE countries (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	country country,
	vat INTEGER,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

INSERT INTO countries (country, vat) VALUES ('Latvia', 21), ('Austria',20),('Bulgaria', 20),('Poland',23),('Romania',19),('Slovakia', 20);

-- +goose Down
DROP TABLE countries;
