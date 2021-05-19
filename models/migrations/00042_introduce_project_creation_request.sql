-- +goose Up
CREATE TYPE project_creation_request_status AS ENUM (
	'opened',
	'accepted',
        'rejected'
);

CREATE TABLE create_project_request (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	asset_id UUID REFERENCES assets NOT NULL,
        organization_id UUID REFERENCES organizations NOT NULL,
        user_id UUID REFERENCES users NOT NULL,
	status project_creation_request_status NOT NULL DEFAULT 'opened',
        token_id UUID REFERENCES tokens,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
        deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE(asset_id, organization_id, user_id, status)
);

ALTER TYPE token_purpose RENAME TO old_token_purpose;
CREATE TYPE token_purpose AS ENUM ('session', 'create', 'resetpwd', 'createprj');
ALTER TABLE tokens ALTER COLUMN purpose TYPE token_purpose using purpose::TEXT::token_purpose;

DROP TYPE old_token_purpose;

-- +goose Down
ALTER TYPE token_purpose RENAME TO old_token_purpose;
CREATE TYPE token_purpose AS ENUM ('session', 'create', 'resetpwd');
ALTER TABLE tokens ALTER COLUMN purpose TYPE token_purpose using purpose::TEXT::token_purpose;

DROP TYPE old_token_purpose;

DROP TABLE create_project_request;
DROP TYPE project_creation_request_status;
