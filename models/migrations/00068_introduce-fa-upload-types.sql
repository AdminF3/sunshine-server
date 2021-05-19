-- +goose NO TRANSACTION
-- +goose Up

-- +goose StatementBegin
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'fa financial statements';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'fa bank confirmation';
-- +goose StatementEnd

DROP TABLE fa_attachments_financial_statement;
DROP TABLE fa_attachments_bank_confirmation;
-- +goose Down
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

