-- +goose Up
CREATE TABLE contracts (
       id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
       project UUID REFERENCES projects ON DELETE CASCADE UNIQUE NOT NULL,
       fields JSONB,
       agreement JSONB,
       annex1 JSONB,
       annex2 JSONB,
       annex3 JSONB,
       annex4 JSONB,
       annex5 JSONB,

       created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
       updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
       deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE contracts;
