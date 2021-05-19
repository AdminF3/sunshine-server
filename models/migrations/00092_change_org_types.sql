-- +goose Up
ALTER TABLE organizations drop COLUMN legal_status;
ALTER TABLE organizations ADD COLUMN registration_number TEXT;

-- +goose Down
ALTER TABLE organizations ADD COLUMN legal_status integer[];
ALTER TABLE organizations drop COLUMN registration_number;
