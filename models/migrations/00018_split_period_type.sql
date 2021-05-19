-- +goose Up
CREATE TYPE period_type AS ENUM (
	'airex_windows',
	'airex_total',
	'total_energy_consumption',
	'total_energy_consumption_circulation',
	'indoor_temp'
);
CREATE TABLE ic_periods (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	indoorclima_id UUID REFERENCES indoor_climas ON DELETE CASCADE NOT NULL,

	type period_type NOT NULL,
	n NUMERIC DEFAULT 0,
	n1 NUMERIC DEFAULT 0,
	n2 NUMERIC DEFAULT 0,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,


	UNIQUE (indoorclima_id, type)
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migrate_periods_up() RETURNS VOID AS $$
DECLARE v indoor_climas%rowtype;
BEGIN
	FOR v IN SELECT * FROM indoor_climas LOOP
		INSERT INTO ic_periods (indoorclima_id, type, n, n1, n2) VALUES (
			v.id, 'airex_windows',
			v.airex_windows_n,
			v.airex_windows_n1,
			v.airex_windows_n2);
		INSERT INTO ic_periods (indoorclima_id, type, n, n1, n2) VALUES (
			v.id, 'airex_total',
			v.airex_total_n,
			v.airex_total_n1,
			v.airex_total_n2);
		INSERT INTO ic_periods (indoorclima_id, type, n, n1, n2) VALUES (
			v.id, 'total_energy_consumption',
			v.total_energy_consumption_n,
			v.total_energy_consumption_n1,
			v.total_energy_consumption_n2);
		INSERT INTO ic_periods (indoorclima_id, type, n, n1, n2) VALUES (
			v.id, 'total_energy_consumption_circulation',
			v.total_e_consumption_circlosses_n,
			v.total_e_consumption_circlosses_n1,
			v.total_e_consumption_circlosses_n2);
		INSERT INTO ic_periods (indoorclima_id, type, n, n1, n2) VALUES (
			v.id, 'indoor_temp',
			v.indoor_temp_n,
			v.indoor_temp_n1,
			v.indoor_temp_n2);
	END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migrate_periods_up();
DROP FUNCTION migrate_periods_up;

ALTER TABLE indoor_climas
	DROP COLUMN airex_windows_n,
	DROP COLUMN airex_windows_n1,
	DROP COLUMN airex_windows_n2,
	DROP COLUMN airex_total_n,
	DROP COLUMN airex_total_n1,
	DROP COLUMN airex_total_n2,
	DROP COLUMN indoor_temp_n,
	DROP COLUMN indoor_temp_n1,
	DROP COLUMN indoor_temp_n2,
	DROP COLUMN total_energy_consumption_n,
	DROP COLUMN total_energy_consumption_n1,
	DROP COLUMN total_energy_consumption_n2,
	DROP COLUMN total_e_consumption_circlosses_n,
	DROP COLUMN total_e_consumption_circlosses_n1,
	DROP COLUMN total_e_consumption_circlosses_n2;

-- +goose Down
ALTER TABLE indoor_climas
	ADD COLUMN airex_windows_n NUMERIC,
	ADD COLUMN airex_windows_n1 NUMERIC,
	ADD COLUMN airex_windows_n2 NUMERIC,
	ADD COLUMN airex_total_n NUMERIC,
	ADD COLUMN airex_total_n1 NUMERIC,
	ADD COLUMN airex_total_n2 NUMERIC,
	ADD COLUMN indoor_temp_n NUMERIC,
	ADD COLUMN indoor_temp_n1 NUMERIC,
	ADD COLUMN indoor_temp_n2 NUMERIC,
	ADD COLUMN total_energy_consumption_n NUMERIC,
	ADD COLUMN total_energy_consumption_n1 NUMERIC,
	ADD COLUMN total_energy_consumption_n2 NUMERIC,
	ADD COLUMN total_e_consumption_circlosses_n NUMERIC,
	ADD COLUMN total_e_consumption_circlosses_n1 NUMERIC,
	ADD COLUMN total_e_consumption_circlosses_n2 NUMERIC;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migrate_periods_down() RETURNS VOID AS $$
DECLARE v indoor_climas%rowtype;
DECLARE p ic_periods%rowtype;
BEGIN
	FOR v IN (SELECT * FROM indoor_climas) LOOP
		FOR p IN (SELECT * FROM ic_periods WHERE indoorclima_id = v.id) LOOP
			IF p.type = 'airex_windows' THEN
				UPDATE indoor_climas SET
					airex_windows_n = p.n,
					airex_windows_n1 = p.n1,
					airex_windows_n2 = p.n2
					WHERE id = v.id;
			END IF;
			IF p.type = 'airex_total' THEN
				UPDATE indoor_climas SET
					airex_total_n = p.n,
					airex_total_n1 = p.n1,
					airex_total_n2 = p.n2
					WHERE id = v.id;
			END IF;
			IF p.type = 'total_energy_consumption' THEN
				UPDATE indoor_climas SET
					total_energy_consumption_n = p.n,
					total_energy_consumption_n1 = p.n1,
					total_energy_consumption_n2 = p.n2
					WHERE id = v.id;
			END IF;
			IF p.type = 'total_energy_consumption_circulation' THEN
				UPDATE indoor_climas SET
					total_e_consumption_circlosses_n = p.n,
					total_e_consumption_circlosses_n1 = p.n1,
					total_e_consumption_circlosses_n2 = p.n2
					WHERE id = v.id;
			END IF;
			IF p.type = 'indoor_temp' THEN
				UPDATE indoor_climas SET
					indoor_temp_n = p.n,
					indoor_temp_n1 = p.n1,
					indoor_temp_n2 = p.n2
					WHERE id = v.id;
			END IF;
		END LOOP;
	END LOOP;
RETURN;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migrate_periods_down();
DROP FUNCTION  migrate_periods_down;

DROP TABLE ic_periods;
DROP TYPE period_type;
