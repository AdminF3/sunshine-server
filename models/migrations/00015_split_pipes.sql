-- +goose Up
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

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migratePipes() RETURNS VOID AS $$
DECLARE
	icid UUID;
BEGIN
FOR icid IN
	SELECT id FROM indoor_climas ORDER BY id LOOP
	FOR i IN 1..10 LOOP
		INSERT INTO basement_pipes (indoorclima_id, quality, installed_length, diameter, heat_loss_unit, heat_loss_year) VALUES (icid,0,0,0,0,0);
		INSERT INTO attic_pipes (indoorclima_id, quality, installed_length, diameter, heat_loss_unit, heat_loss_year) VALUES (icid,0,0,0,0,0);
	END LOOP;
END LOOP;
RETURN;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migratePipes();
DROP TABLE pipes;

-- +goose Down
CREATE TABLE pipes (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	indoorclima_id UUID REFERENCES indoor_climas ON DELETE CASCADE NOT NULL,

	quality INTEGER,
	installed_length NUMERIC,
	diameter NUMERIC,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);


-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migratePipes() RETURNS VOID AS $$
DECLARE
	icid UUID;
BEGIN
FOR icid IN
	SELECT id FROM indoor_climas ORDER BY id LOOP
	FOR i IN 1..20 LOOP
		INSERT INTO pipes (indoorclima_id, quality, installed_length, diameter, heat_loss_unit, heat_loss_year) VALUES (icid,0,0,0,0,0);
	END LOOP;
END LOOP;
RETURN;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migratePipes();
DROP TABLE basement_pipes;
DROP TABLE attic_pipes;

