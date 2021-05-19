-- +goose Up
ALTER TABLE tokens
      ALTER COLUMN id SET DEFAULT PUBLIC.gen_random_uuid(),
      DROP COLUMN expires_at,
      ADD COLUMN ttl BIGINT NOT NULL;

-- +goose Down
ALTER TABLE tokens
      DROP COLUMN ttl,
      ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
      ALTER COLUMN id DROP DEFAULT;
