-- +goose Up
ALTER TABLE meetings ALTER COLUMN objective DROP NOT NULL;
ALTER TABLE meetings ALTER COLUMN participant DROP NOT NULL;
ALTER TABLE meetings ALTER COLUMN stage DROP NOT NULL;

ALTER TABLE meetings ADD COLUMN guest_email TEXT;
ALTER TABLE meetings ADD COLUMN guest_phone TEXT;
ALTER TABLE meetings ADD COLUMN actions_taken TEXT;
ALTER TABLE meetings ADD COLUMN next_contact TIMESTAMP WITH TIME ZONE;

-- +goose Down
ALTER TABLE meetings ALTER COLUMN objective SET NOT NULL;
ALTER TABLE meetings ALTER COLUMN participant SET NOT NULL;
ALTER TABLE meetings ALTER COLUMN stage SET NOT NULL;

ALTER TABLE meetings DROP COLUMN guest_email;
ALTER TABLE meetings DROP COLUMN guest_phone;
ALTER TABLE meetings DROP COLUMN actions_taken;
ALTER TABLE meetings DROP COLUMN next_contact;
