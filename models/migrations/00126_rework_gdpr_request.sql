-- +goose Up

ALTER TABLE gdpr_requests
   DROP COLUMN requester_country,
   ADD COLUMN reason TEXT,
   ADD COLUMN information TEXT;

-- +goose Down

ALTER TABLE gdpr_requests
   ADD COLUMN requester_country TEXT,
   DROP COLUMN reason,
   DROP COLUMN information;
