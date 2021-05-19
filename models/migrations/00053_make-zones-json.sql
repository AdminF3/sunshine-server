-- +goose Up
DROP TABLE ic_zones;
DROP TABLE basement_pipes;
DROP TABLE attic_pipes;

ALTER TABLE indoor_climas ADD COLUMN zones JSONB;
ALTER TABLE indoor_climas ADD COLUMN basement_pipes JSONB;
ALTER TABLE indoor_climas ADD COLUMN attic_pipes JSONB;

-- +goose Down
ALTER TABLE indoor_climas DROP COLUMN zones;
ALTER TABLE indoor_climas DROP COLUMN basement_pipes;
ALTER TABLE indoor_climas DROP COLUMN attic_pipes;

CREATE TABLE basement_pipes (
id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
indoorclima_id UUID REFERENCES indoor_climas ON DELETE CASCADE NOT NULL,

quality INTEGER,
installed_length NUMERIC,
diameter NUMERIC,
heat_loss_unit NUMERIC,
heat_loss_year NUMERIC,

created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE attic_pipes (
id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
indoorclima_id UUID REFERENCES indoor_climas ON DELETE CASCADE NOT NULL,

quality INTEGER,
installed_length NUMERIC,
diameter NUMERIC,
heat_loss_unit NUMERIC,
heat_loss_year NUMERIC,

created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE ic_zones (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	indoorclima_id UUID REFERENCES indoor_climas ON DELETE CASCADE NOT NULL,

	type TEXT NOT NULL,
	area NUMERIC DEFAULT 0,
	uvalue NUMERIC DEFAULT 0,
	outdoor_temp_n NUMERIC DEFAULT 0,
	outdoor_temp_n1 NUMERIC DEFAULT 0,
	outdoor_temp_n2 NUMERIC DEFAULT 0,
	temp_diff_n NUMERIC DEFAULT 0,
	temp_diff_n1 NUMERIC DEFAULT 0,
	temp_diff_n2 NUMERIC DEFAULT 0,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE (indoorclima_id, type)
);

