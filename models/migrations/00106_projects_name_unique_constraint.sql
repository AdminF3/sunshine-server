-- +goose Up
WITH ct AS
(
  SELECT id , row_number() OVER (PARTITION by name ORDER BY created_at ASC) rn
  FROM   projects
)
UPDATE projects
SET    name = name || CASE WHEN ct.rn = 1 THEN '' ELSE '-' || (ct.rn-1)::text END
FROM   ct
WHERE  ct.id = projects.id;

ALTER TABLE projects ADD CONSTRAINT projects_name_ukey UNIQUE (name);

-- +goose Down
ALTER TABLE projects DROP CONSTRAINT projects_name_ukey;
