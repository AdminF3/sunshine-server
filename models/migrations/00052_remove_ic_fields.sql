-- +goose Up
ALTER TYPE period_type RENAME TO old_period_type;
CREATE TYPE period_type AS ENUM ('airex_windows',
'airex_total',
'total_energy_consumption',
'total_energy_consumption_circulation',
'indoor_temp',
'outdoor_air_temp'
);
ALTER TABLE ic_periods ALTER COLUMN type TYPE period_type USING type::TEXT::period_type;

-- this is calculable field - we do not need to store it
ALTER TABLE ic_zones DROP COLUMN heat_loss_coeff;

-- drop the zone_type and period_type. Currently the will be
-- dynamically selected, so we do not know their count therefore their
-- type
ALTER TABLE ic_zones ALTER COLUMN type SET DATA TYPE TEXT;
DROP TYPE zone_type;
DROP TYPE old_period_type;

-- +goose Down
ALTER TABLE ic_zones ADD COLUMN heat_loss_coeff NUMERIC DEFAULT 0;

CREATE TYPE zone_type AS ENUM (
'attic_zone1',
'attic_zone2',
'basement_ceiling_zone1',
'basement_ceiling_zone2',
'ground_zone1',
'ground_zone2',
'roof_zone1',
'roof_zone2',
'basewall_zone1',
'basewall_zone2',
'external_door_env1_zone1',
'external_door_env1_zone2',
'external_door_env2_zone1',
'external_door_env2_zone2',
'window_env1_zone1',
'window_env1_zone2',
'window_env2_zone1',
'window_env2_zone2',
'external_wall_env1_zone1',
'external_wall_env1_zone2',
'external_wall_env2_zone1',
'external_wall_env2_zone2',
'external_wall_env3_zone1',
'external_wall_env3_zone2',
'external_wall_env4_zone1',
'external_wall_env4_zone2'
);
ALTER TABLE ic_zones ALTER COLUMN type SET DATA TYPE zone_type USING type::zone_type;

ALTER TYPE period_type RENAME TO old_period_type;
CREATE TYPE period_type AS ENUM ('airex_windows',
'airex_total',
'total_energy_consumption',
'total_energy_consumption_circulation',
'indoor_temp');
ALTER TABLE ic_periods ALTER COLUMN type TYPE period_type USING type::TEXT::period_type;
DROP TYPE old_period_type;
