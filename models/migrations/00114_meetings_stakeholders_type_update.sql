-- +goose Up

-- migrate stakeholders from LegalForm enum to its own (StakeholdersType)
-- PUBLIC_ORGANIZATION (3) -> CENTRAL_GOVERNMENT (3)
-- RESIDENTS_COMMUNITY (4) -> RESIDENT (7)

UPDATE meetings SET stakeholder = 7 WHERE stakeholder = 4;

-- +goose Down
UPDATE meetings SET stakeholder = 4 WHERE stakeholder = 7;
