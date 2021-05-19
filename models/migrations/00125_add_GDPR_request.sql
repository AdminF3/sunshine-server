-- +goose Up
-- +goose NO TRANSACTION

ALTER TABLE gdpr_requests
   ADD COLUMN requester_name    TEXT,
   ADD COLUMN requester_phone   TEXT,
   ADD COLUMN requester_email   TEXT,
   ADD COLUMN requester_address TEXT,
   ADD COLUMN requester_country TEXT,
   ADD COLUMN name              TEXT,
   ADD COLUMN phone             TEXT,
   ADD COLUMN email             TEXT,
   ADD COLUMN address           TEXT;

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'gdpr request';

-- +goose Down

ALTER TABLE gdpr_requests
   DROP COLUMN requester_name,
   DROP COLUMN requester_phone,
   DROP COLUMN requester_email,
   DROP COLUMN requester_address,
   DROP COLUMN requester_country,
   DROP COLUMN name,
   DROP COLUMN phone,
   DROP COLUMN email,
   DROP COLUMN address;
