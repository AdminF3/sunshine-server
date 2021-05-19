-- +goose Up
CREATE TYPE country AS ENUM ('Austria', 'Bulgaria', 'Latvia', 'Poland', 'Romania', 'Slovakia');
ALTER TABLE assets ALTER country TYPE country USING INITCAP(country)::country;
ALTER TABLE organizations ALTER country TYPE country USING INITCAP(country)::country;
ALTER TABLE projects ALTER country TYPE country USING INITCAP(country)::country;
ALTER TABLE users ALTER country TYPE country USING INITCAP(country)::country;

-- +goose Down
ALTER TABLE assets ALTER country TYPE text USING country::text;
ALTER TABLE organizations ALTER country TYPE text USING country::text;
ALTER TABLE projects ALTER country TYPE text USING country::text;
ALTER TABLE users ALTER country TYPE text USING country::text;
DROP TYPE country;

