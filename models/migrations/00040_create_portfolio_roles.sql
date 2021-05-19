-- +goose Up
CREATE TYPE portfolio_roles AS ENUM ('portfolio_director', 'fund_manager', 'country_admin', 'data_protection_officer');

CREATE TABLE country_roles (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	country TEXT NOT NULL,
	user_id UUID REFERENCES users NOT NULL,
	role PORTFOLIO_ROLES NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

	UNIQUE(country, role, user_id)
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migratePDs() RETURNS VOID AS $$
DECLARE
	cnt TEXT;
        uid UUID;
BEGIN
FOR cnt, uid IN
		SELECT  country, user_id FROM portfolio_directors
        LOOP
		INSERT INTO country_roles (country, user_id, role) VALUES (cnt, uid, 'portfolio_director');
	END LOOP;
RETURN;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migratePDs();
DROP FUNCTION migratePDs;
DROP TABLE portfolio_directors;

-- +goose Down

-- migration from 00027#goose_down
CREATE TABLE portfolio_directors (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	country TEXT NOT NULL,
	user_id UUID REFERENCES users NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE (country, user_id)
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migrateRoles() RETURNS VOID AS $$
DECLARE
	cnt TEXT;
	uid UUID;
BEGIN
FOR cnt, uid IN
		SELECT  country, user_id FROM country_roles WHERE role = 'portfolio_director'
	LOOP
		INSERT INTO portfolio_directors (country, user_id) VALUES (cnt, uid);
	END LOOP;
RETURN;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migrateRoles();
DROP FUNCTION migrateRoles;
DROP TABLE country_roles;
DROP TYPE portfolio_roles;
