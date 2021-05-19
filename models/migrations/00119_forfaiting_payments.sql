-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'proof of transfer';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'forfaiting_payment';

CREATE TYPE currency AS ENUM ('EUR', 'ALL', 'AMD', 'BYN', 'BAM', 'BGN', 'HRK', 'CSJ', 'DKK', 'GEL', 'HUF', 'ISK',
	'CHF', 'MDL', 'NOK', 'PLN', 'RON', 'RUB', 'RSD', 'SEK', 'TRY', 'UAH', 'GBP');

CREATE TABLE forfaiting_payments (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),
	transfer_value INTEGER,
	currency currency,
	project_id UUID REFERENCES projects NOT NULL,
	transfer_date TIMESTAMP WITH TIME ZONE,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE forfaiting_payments;
DROP TYPE currency;
