-- +goose Up
ALTER TABLE projects ADD COLUMN portfolio_director UUID REFERENCES users;
UPDATE projects SET portfolio_director = (select id from users where is_admin = true limit 1);
ALTER TABLE projects ALTER COLUMN portfolio_director SET NOT NULL;

-- +goose Down
ALTER TABLE projects DROP COLUMN portfolio_director;
