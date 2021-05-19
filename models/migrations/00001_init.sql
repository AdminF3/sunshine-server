-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TYPE token_purpose AS ENUM ('session', 'create', 'resetpwd');
CREATE TYPE organization_role AS ENUM ('lear', 'lsign', 'leaa', 'member');
CREATE TYPE project_role AS ENUM ('pm', 'paco', 'plsign', 'tama', 'teme');

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	email CHARACTER VARYING(255) UNIQUE NOT NULL,
	password BYTEA NOT NULL,
	name CHARACTER VARYING(255),
	identity CHARACTER VARYING(64),
	address CHARACTER VARYING(255),
	avatar CHARACTER VARYING(255),
	telephone CHARACTER VARYING(16),
	is_admin BOOLEAN DEFAULT false,
	status NUMERIC,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE tokens (
	id UUID PRIMARY KEY,
	user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
	purpose token_purpose NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE social_profiles (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
	type CHARACTER VARYING(16),
	handle CHARACTER VARYING(255),

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE organizations (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	name CHARACTER VARYING(255) NOT NULL,
	vat CHARACTER VARYING(255) UNIQUE NOT NULL,
	address CHARACTER VARYING(255),
	telephone CHARACTER VARYING(16),
	website CHARACTER VARYING(255),
	logo CHARACTER VARYING(255),
	legal_status NUMERIC[],
	legal_form NUMERIC,
	registered_at TIMESTAMP WITH TIME ZONE NOT NULL,
	status NUMERIC,
	email CHARACTER VARYING(255) NOT NULL,
	lear_name CHARACTER VARYING(255),

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE organization_roles (
	user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
	organization_id UUID REFERENCES organizations ON DELETE CASCADE NOT NULL,
	position organization_role NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE (user_id, organization_id, position)
);


CREATE TABLE assets (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	cadastre CHARACTER VARYING(64) UNIQUE NOT NULL,
	owner UUID REFERENCES organizations ON DELETE CASCADE NOT NULL,
	coords REAL[],
	address CHARACTER VARYING(255) NOT NULL,
	area NUMERIC CHECK (area >= 0 AND area < 65535),
	heated_area NUMERIC CHECK (heated_area >= 0 AND heated_area < 65535),
	billing_area NUMERIC CHECK (billing_area >= 0 AND billing_area < 65535),
	flats NUMERIC CHECK (flats >=0 AND flats < 255),
	floors NUMERIC CHECK (floors >=0 AND floors < 255),
	stair_cases NUMERIC CHECK (stair_cases >=0 AND stair_cases < 255),
	building_type NUMERIC,
	heating_type NUMERIC,
	status NUMERIC,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE projects (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	name CHARACTER VARYING(255),
	organization CHARACTER VARYING(36) NOT NULL,
	esco CHARACTER VARYING(36),
	asset CHARACTER VARYING(36) NOT NULL,
	client CHARACTER VARYING(36),
	status NUMERIC,
	air_temperature DOUBLE PRECISION,
	water_temperature DOUBLE PRECISION,
	guaranteed_savings DOUBLE PRECISION,
	construction_from TIMESTAMP WITH TIME ZONE,
	construction_to TIMESTAMP WITH TIME ZONE,
	contract_term NUMERIC,
	first_year NUMERIC,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE project_roles (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
	project_id UUID REFERENCES projects ON DELETE CASCADE NOT NULL,
	position project_role NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE (user_id, project_id, position)
);

-- +goose Down
DROP TABLE tokens;
DROP TABLE social_profiles;
DROP TABLE organization_roles;
DROP TABLE project_roles;
DROP TABLE projects;
DROP TABLE assets;
DROP TABLE organizations;
DROP TABLE users;

DROP TYPE token_purpose;
DROP TYPE organization_role;
DROP TYPE project_role;
