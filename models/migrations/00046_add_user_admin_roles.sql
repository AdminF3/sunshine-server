-- +goose Up
ALTER TABLE users ADD COLUMN platform_manager BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN admin_network_manager BOOLEAN DEFAULT FALSE;

-- +goose Down
ALTER TABLE users DROP COLUMN platform_manager;
ALTER TABLE users DROP COLUMN admin_network_manager;
