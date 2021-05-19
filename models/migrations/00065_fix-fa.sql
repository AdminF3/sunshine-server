-- +goose Up
ALTER TABLE forfaiting_applications DROP COLUMN accepted;

CREATE TABLE bank_accounts (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),

	beneficiary_name TEXT,
        bank_name_address TEXT,
        iban TEXT,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE forfaiting_applications ADD COLUMN bank_account_id UUID REFERENCES bank_accounts;
ALTER TABLE forfaiting_applications ADD COLUMN private_bond BOOLEAN;
ALTER TABLE forfaiting_applications ADD COLUMN manager_id UUID REFERENCES users NOT NULL;
ALTER TABLE forfaiting_applications ADD COLUMN finance INTEGER;

ALTER TABLE fa_reviews DROP COLUMN status;
ALTER TABLE fa_reviews ADD COLUMN approved BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE fa_reviews ADD COLUMN type INTEGER;

CREATE TABLE fa_attachments_financial_statement (
	forfaiting_application_id UUID REFERENCES forfaiting_applications(id) NOT NULL,
        attachment_id UUID REFERENCES attachments(id) NOT NULL,

CONSTRAINT fa_attachments_financial_statement_pkey PRIMARY KEY(forfaiting_application_id, attachment_id)
);

CREATE TABLE fa_attachments_bank_confirmation (
	forfaiting_application_id UUID REFERENCES forfaiting_applications(id) NOT NULL,
        attachment_id UUID REFERENCES attachments(id) NOT NULL,

CONSTRAINT fa_attachments_bank_confirmation_pkey PRIMARY KEY(forfaiting_application_id, attachment_id)
);

-- +goose Down
ALTER TABLE forfaiting_applications ADD COLUMN accepted BOOLEAN DEFAULT FALSE;
ALTER TABLE forfaiting_applications DROP COLUMN bank_account_id;
DROP TABLE bank_accounts;
ALTER TABLE forfaiting_applications DROP COLUMN private_bond;
ALTER TABLE forfaiting_applications DROP COLUMN manager_id;
ALTER TABLE forfaiting_applications DROP COLUMN finance;

ALTER TABLE fa_reviews ADD COLUMN status transition_request_status;
ALTER TABLE fa_reviews DROP COLUMN approved;
ALTER TABLE fa_reviews DROP COLUMN type;

DROP TABLE fa_attachments_financial_statement;
DROP TABLE fa_attachments_bank_confirmation;
