-- +goose Up
ALTER TABLE assets DROP CONSTRAINT assets_area_check;
ALTER TABLE assets DROP CONSTRAINT assets_heated_area_check;
ALTER TABLE assets DROP CONSTRAINT assets_billing_area_check;
ALTER TABLE assets DROP CONSTRAINT assets_flats_check;
ALTER TABLE assets DROP CONSTRAINT assets_floors_check;
ALTER TABLE assets DROP CONSTRAINT assets_stair_cases_check;

ALTER TABLE assets ADD CONSTRAINT assets_area_check CHECK (area >= 0);
ALTER TABLE assets ADD CONSTRAINT assets_heated_area_check CHECK (heated_area >= 0);
ALTER TABLE assets ADD CONSTRAINT assets_billing_area_check CHECK (billing_area >= 0);
ALTER TABLE assets ADD CONSTRAINT assets_flats_check CHECK (flats >=0);
ALTER TABLE assets ADD CONSTRAINT assets_floors_check CHECK (floors >=0);
ALTER TABLE assets ADD CONSTRAINT assets_stair_cases_check CHECK (stair_cases >=0);


-- +goose Down
ALTER TABLE assets DROP CONSTRAINT assets_area_check;
ALTER TABLE assets DROP CONSTRAINT assets_heated_area_check;
ALTER TABLE assets DROP CONSTRAINT assets_billing_area_check;
ALTER TABLE assets DROP CONSTRAINT assets_flats_check;
ALTER TABLE assets DROP CONSTRAINT assets_floors_check;
ALTER TABLE assets DROP CONSTRAINT assets_stair_cases_check;

ALTER TABLE assets ADD CONSTRAINT assets_area_check CHECK (area >= 0 AND area < 65535);
ALTER TABLE assets ADD CONSTRAINT assets_heated_area_check CHECK (heated_area >= 0 AND heated_area < 65535);
ALTER TABLE assets ADD CONSTRAINT assets_billing_area_check CHECK (billing_area >= 0 AND billing_area < 65535);
ALTER TABLE assets ADD CONSTRAINT assets_flats_check CHECK (flats >=0 AND flats < 255);
ALTER TABLE assets ADD CONSTRAINT assets_floors_check CHECK (floors >=0 AND floors < 255);
ALTER TABLE assets ADD CONSTRAINT assets_stair_cases_check CHECK (stair_cases >=0 AND stair_cases < 255);
