-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'signed_constructions_and_subcontracting_contracts';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'renovation_and_engineering_works';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'commissioning';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'technical_closing';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'forfaiting_payout';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'results_monitoring_and_analysis';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'payment_of_loans';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'forfaiting_annual_check';

ALTER TYPE milestone ADD VALUE IF NOT EXISTS 'building_maintenance';

-- +goose Down
-- adding too many new types for a sane revert.
SELECT 1;
