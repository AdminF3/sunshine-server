-- +goose Up
UPDATE organizations SET legal_form = 5 WHERE legal_form = 1;
UPDATE organizations SET legal_form = 4 WHERE legal_form = 2;
UPDATE organizations SET legal_form = 1 WHERE legal_form = 3;
UPDATE organizations SET legal_form = 4 WHERE legal_form = 4;
UPDATE organizations SET legal_form = 5 WHERE legal_form = 5;
UPDATE organizations SET legal_form = 5 WHERE legal_form = 6;
UPDATE organizations SET legal_form = 4 WHERE legal_form = 7;
UPDATE organizations SET legal_form = 1 WHERE legal_form = 8;
UPDATE organizations SET legal_form = 5 WHERE legal_form = 9;
UPDATE organizations SET legal_form = 4 WHERE legal_form = 10;
UPDATE organizations SET legal_form = 3 WHERE legal_form = 11;

-- +goose Down
SELECT 1;
