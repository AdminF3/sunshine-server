-- +goose Up
DROP TABLE organization_roles;
CREATE TABLE organization_roles (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
	organization_id UUID REFERENCES organizations ON DELETE CASCADE NOT NULL,
	position organization_role NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE (user_id, organization_id, position)
);

-- +goose Down
DROP TABLE organization_roles;
CREATE TABLE organization_roles (
	user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
	organization_id UUID REFERENCES organizations ON DELETE CASCADE NOT NULL,
	position organization_role NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE,

	UNIQUE (user_id, organization_id, position)
);
