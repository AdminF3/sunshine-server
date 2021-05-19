-- +goose Up
ALTER TABLE indoor_climas
	DROP COLUMN total_ht_zone1,
	DROP COLUMN total_ht_zone2,
	DROP COLUMN total_ht,
	DROP COLUMN heat_gains_internal,
	DROP COLUMN heat_gains_solar,
	DROP COLUMN airex_building_n,
	DROP COLUMN airex_building_n1,
	DROP COLUMN airex_building_n2,
	DROP COLUMN circulation_losses_n,
	DROP COLUMN circulation_losses_n1,
	DROP COLUMN circulation_losses_n2,
	DROP COLUMN distribution_losses_basement,
	DROP COLUMN distribution_losses_attic,
	DROP COLUMN total_measured_n,
	DROP COLUMN total_measured_n1,
	DROP COLUMN total_measured_n2,
	DROP COLUMN total_calculated_n,
	DROP COLUMN total_calculated_n1,
	DROP COLUMN total_calculated_n2;

ALTER TABLE pipes
	DROP COLUMN heat_loss_unit,
	DROP COLUMN heat_loss_year;

-- +goose Down
ALTER TABLE indoor_climas
	ADD COLUMN total_ht_zone1 NUMERIC,
	ADD COLUMN total_ht_zone2 NUMERIC,
	ADD COLUMN total_ht NUMERIC,
	ADD COLUMN heat_gains_internal NUMERIC,
	ADD COLUMN heat_gains_solar NUMERIC,
	ADD COLUMN airex_building_n NUMERIC,
	ADD COLUMN airex_building_n1 NUMERIC,
	ADD COLUMN airex_building_n2 NUMERIC,
	ADD COLUMN circulation_losses_n NUMERIC,
	ADD COLUMN circulation_losses_n1 NUMERIC,
	ADD COLUMN circulation_losses_n2 NUMERIC,
	ADD COLUMN distribution_losses_basement NUMERIC,
	ADD COLUMN distribution_losses_attic NUMERIC,
	ADD COLUMN total_measured_n NUMERIC,
	ADD COLUMN total_measured_n1 NUMERIC,
	ADD COLUMN total_measured_n2 NUMERIC,
	ADD COLUMN total_calculated_n NUMERIC,
	ADD COLUMN total_calculated_n1 NUMERIC,
	ADD COLUMN total_calculated_n2 NUMERIC;

ALTER TABLE pipes
	ADD COLUMN heat_loss_unit NUMERIC,
	ADD COLUMN heat_loss_year NUMERIC;
