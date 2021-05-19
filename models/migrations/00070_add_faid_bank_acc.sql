-- +goose Up
ALTER TABLE bank_accounts ADD COLUMN fa_id UUID REFERENCES forfaiting_applications;

ALTER TABLE forfaiting_applications DROP COLUMN bank_account_id;

-- +goose Down
ALTER TABLE bank_accounts DROP COLUMN fa_id;

ALTER TABLE forfaiting_applications ADD COLUMN bank_account_id UUID REFERENCES bank_accounts;
