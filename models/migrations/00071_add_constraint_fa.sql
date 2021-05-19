-- +goose Up
ALTER TABLE forfaiting_applications ADD UNIQUE (project_id);
-- +goose Down
ALTER TABLE forfaiting_applications DROP CONSTRAINT forfaiting_applications_project_id_key;
