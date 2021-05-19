-- +goose Up
ALTER TABLE organizations ALTER COLUMN vat drop not null;
ALTER TABLE organizations ALTER COLUMN telephone drop not null;
ALTER TABLE organizations ALTER COLUMN email drop not null;
ALTER TABLE organizations ALTER COLUMN registered_at drop not null;


-- +goose Down
ALTER TABLE organizations ALTER COLUMN vat set not null;
ALTER TABLE organizations ALTER COLUMN telephone set not null;
ALTER TABLE organizations ALTER COLUMN email set not null;
ALTER TABLE organizations ALTER COLUMN registered_at set not null;

