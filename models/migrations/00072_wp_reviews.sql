-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'commissioning report';

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'independent energy audit measurement and verification report';

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'insurance policies';

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'defects declarations';

CREATE TABLE wp_reviews (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	wp_id UUID REFERENCES work_phase NOT NULL,
	author UUID REFERENCES users,
	approved BOOLEAN,
	comment TEXT,
	type INTEGER,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE wp_reviews;

UPDATE attachments SET upload_type = 'general leaflet'  WHERE upload_type = 'commissioning report';

UPDATE attachments SET upload_type = 'general leaflet'  WHERE upload_type = 'independent energy audit measurement and verification report';

UPDATE attachments SET upload_type = 'general leaflet'  WHERE upload_type = 'insurance policies';

UPDATE attachments SET upload_type = 'general leaflet' WHERE upload_type = 'defects declarations';

