-- +goose Up
-- +goose NO TRANSACTION

ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'acquisition other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'feasibility other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'commitment other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'design other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'preparation other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'kickoff other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'signed other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'renovation other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'commissioning other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'technical other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'forfaiting other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'results other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'payment other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'annual other';
ALTER TYPE upload_type ADD VALUE IF NOT EXISTS 'building other';


-- +goose Down
SELECT 1;