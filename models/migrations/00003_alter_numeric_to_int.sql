-- +goose Up
ALTER TABLE users ALTER COLUMN status TYPE INTEGER using status::integer;

ALTER TABLE organizations
ALTER COLUMN legal_status TYPE INTEGER[] using legal_status::integer[],
ALTER COLUMN legal_form TYPE INTEGER using legal_form::integer,
ALTER COLUMN status TYPE INTEGER using status::integer;

ALTER TABLE assets
ALTER COLUMN area TYPE INTEGER using area::integer,
ALTER COLUMN heated_area TYPE INTEGER using heated_area::integer,
ALTER COLUMN billing_area TYPE INTEGER using billing_area::integer,
ALTER COLUMN flats TYPE INTEGER using flats::integer,
ALTER COLUMN floors TYPE INTEGER using floors::integer,
ALTER COLUMN stair_cases TYPE INTEGER using stair_cases::integer,
ALTER COLUMN building_type TYPE INTEGER using building_type::integer,
ALTER COLUMN heating_type TYPE INTEGER using heating_type::integer,
ALTER COLUMN status TYPE INTEGER using status::integer;

ALTER TABLE projects
ALTER COLUMN status TYPE INTEGER using status::integer,
ALTER COLUMN contract_term TYPE INTEGER using contract_term::integer,
ALTER COLUMN first_year TYPE INTEGER using first_year::integer;

-- +goose Down
ALTER TABLE users ALTER COLUMN status TYPE NUMERIC using status::numeric;

ALTER TABLE organizations
ALTER COLUMN legal_status TYPE NUMERIC[] using legal_status::numeric[],
ALTER COLUMN legal_form TYPE NUMERIC using legal_form::numeric,
ALTER COLUMN status TYPE NUMERIC using status::numeric;

ALTER TABLE assets
ALTER COLUMN area TYPE NUMERIC using area::numeric,
ALTER COLUMN heated_area TYPE NUMERIC using heated_area::numeric,
ALTER COLUMN billing_area TYPE NUMERIC using billing_area::numeric,
ALTER COLUMN flats TYPE NUMERIC using flats::numeric,
ALTER COLUMN floors TYPE NUMERIC using floors::numeric,
ALTER COLUMN stair_cases TYPE NUMERIC using stair_cases::numeric,
ALTER COLUMN building_type TYPE NUMERIC using building_type::numeric,
ALTER COLUMN heating_type TYPE NUMERIC using heating_type::numeric,
ALTER COLUMN status TYPE NUMERIC using status::numeric;

ALTER TABLE projects
ALTER COLUMN status TYPE NUMERIC using status::numeric,
ALTER COLUMN contract_term TYPE NUMERIC using contract_term::numeric,
ALTER COLUMN first_year TYPE NUMERIC using first_year::numeric;
