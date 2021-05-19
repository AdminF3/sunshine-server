package controller

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestICCalc(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewContract(e)

	// test data
	_, prj := stores.NewTestInClima(t, e.IndoorClimaStore)
	pmID := prj.Data.(*models.Project).Roles.PM[0]
	pm, err := e.UserStore.Get(context.Background(), pmID)
	if err != nil {
		t.Fatalf("fetch PM: %v", err)
	}
	ctx := services.NewTestContext(t, e, pm)

	// update indoorclima
	if _, err := contr.UpdateIndoorClima(ctx, prj.ID, dec(t)); err != nil {
		t.Errorf("error occurs while updating indoorclima: %v", err)
	}

	// get the updated indoorclima
	res, err := contr.GetIndoorClima(ctx, prj.ID)
	if err != nil {
		t.Errorf("error occurs while getting indoorclima: %v", err)
	}

	// check for updated values
	ii := res.Data.(*contract.IndoorClima)

	// // heat loss coeff
	for k, v := range ii.Zones {
		if strings.Contains(k, "attic_zone1") {
			if v.HeatLossCoeff != 1120 {
				// area X UValue per zone
				t.Errorf("heat loss coeff is not calc properly, expected (area X UValue), got: %v", v.HeatLossCoeff)

			}

			// temp diff by zones
			if v.TempDiff.N != 11.21 {
				t.Errorf("temp diff is not calc properly, expected (Tin - Tout), got: %v", v.TempDiff.N)
			}
		}
	}

	if ii.TotalHT != 2065.5 {
		// Total sum of all Heat loss coeff in zones.
		t.Errorf("totalHT got: %v, but expected: %v", ii.TotalHT, 2065.5)
	}

	if ii.HeatGainsSolar != 5888 {
		// 12.8 * sum(window's area)
		t.Errorf("heat gain solar got: %v, but expected: %v", ii.HeatGainsSolar, 5888)
	}

	if ii.DistributionLossesAttic != 10.89 {
		t.Errorf("attic pipes exp: %v, got %v", 10.89, ii.DistributionLossesAttic)
	}

	if ii.DistributionLossesBasement != 1.0355 {
		t.Errorf("basement pipes exp: %v, got %v", 1.0355, ii.DistributionLossesBasement)
	}

	if ii.AirexTotal.N != 2228.21681784 || ii.AirexTotal.N != ii.AirexTotal.N1 {
		t.Errorf("airex total exp: %v, got %v", 2228.21, ii.AirexTotal)
	}

}

func dec(t *testing.T) decode {
	t.Helper()

	var buff bytes.Buffer

	window1 := contract.ZoneModel{Area: 332.5, UValue: 1.5}
	window2 := contract.ZoneModel{Area: 127.5, UValue: 2.5}
	attz1 := contract.ZoneModel{Area: 112, UValue: 10}
	extDoor1 := contract.ZoneModel{Area: 10, UValue: 10}
	extDoor2 := contract.ZoneModel{Area: 8, UValue: 3.5}

	bp1 := contract.Pipe{
		Quality:         contract.QualityPoor,
		InstalledLength: 23.5,
		Diameter:        12.8,
	}
	bp2 := contract.Pipe{
		Quality:         contract.QualityGood,
		InstalledLength: 10,
		Diameter:        15,
	}

	ap1 := contract.Pipe{
		Quality:         contract.QualityNo,
		InstalledLength: 100,
		Diameter:        12,
	}
	ap2 := contract.Pipe{
		Quality:         contract.QualityGood,
		InstalledLength: 90,
		Diameter:        10,
	}

	values := contract.IndoorClima{
		Zones: make(contract.JSONZone),
		OutdoorAirTemp: contract.Period{
			N:  1.79,
			N1: 2.99,
			N2: 3.84,
		},
		BasementPipes: contract.JSONPipe{bp1, bp2},
		AtticPipes:    contract.JSONPipe{ap1, ap2},
	}

	values.Zones["window_env1_zone1"] = window1
	values.Zones["window_env2_zone2"] = window2
	values.Zones["external_door_env1_zone2"] = extDoor2
	values.Zones["external_door_env1_zone1"] = extDoor1
	values.Zones["attic_num1_zone1"] = attz1

	enc := gob.NewEncoder(&buff)
	dec := gob.NewDecoder(&buff)

	e := models.Entity(values)
	if err := enc.Encode(e); err != nil {
		t.Errorf("fail to encode buffer: %v", err)
	}

	return MarshalGOB(dec)
}

func TestZoneType(t *testing.T) {
	cases := []struct {
		name string
		key  string
		err  error
	}{
		{
			name: "attic",
			key:  "attic_num1_zone1",
			err:  nil,
		},
		{
			name: "basement_ceiling",
			key:  "basement_ceiling_num5_zone1",
			err:  nil,
		},
		{
			name: "ground",
			key:  "ground_num2_zone2",
			err:  nil,
		},
		{
			name: "roof",
			key:  "roof_num1_zone1",
			err:  nil,
		},
		{
			name: "basewall",
			key:  "basewall_num1_zone1",
			err:  nil,
		},
		{
			name: "exernall_door",
			key:  "external_door_num5_zone1",
			err:  nil,
		},
		{
			name: "window",
			key:  "window_num2_zone2",
			err:  nil,
		},
		{
			name: "external_wall",
			key:  "external_wall_num1_zone1",
			err:  nil,
		},
		{
			name: "zone_exceeds_2",
			key:  "attic_num1_zone3",
			err:  ErrBadInput,
		},
		{
			name: "typo",
			key:  "atic_num1_zone1",
			err:  ErrBadInput,
		},
		{
			name: "no_num",
			key:  "roof_num_zone1",
			err:  ErrBadInput,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ic := contract.IndoorClima{
				Zones: make(contract.JSONZone),
			}
			ic.Zones[c.key] = contract.ZoneModel{}

			if err := validateZones(&ic); !errors.Is(err, c.err) {
				t.Errorf("indoorclima validation fails with: %v but expect BadInput", err)
			}
		})
	}
}
