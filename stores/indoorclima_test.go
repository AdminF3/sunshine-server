package stores

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
)

var (
	period = contract.Period{
		N:  1.11,
		N1: 1.011,
		N2: 1.21,
	}

	pipe = contract.Pipe{
		Quality:         1,
		InstalledLength: 32,
		Diameter:        32.2,
		HeatLossUnit:    113,
		HeatLossYear:    123.2,
	}

	zone = contract.ZoneModel{
		Area:          123.3,
		UValue:        333,
		TempDiff:      period,
		HeatLossCoeff: 223.3,
	}

	BaselineInputDataColumns = []contract.Column{
		{Name: "Month", Kind: contract.Name, Headers: []string{"", "Symbol", "Unit"}},

		{Name: "Heating days", Kind: contract.Count, Headers: []string{"n-2", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: contract.Energy, Headers: []string{"n-2", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: contract.Volume, Headers: []string{"n-2", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: contract.Temperature, Headers: []string{"n-2", "θkū", "°C"}},

		{Name: "Heating days", Kind: contract.Count, Headers: []string{"n-1", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: contract.Energy, Headers: []string{"n-1", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: contract.Volume, Headers: []string{"n-1", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: contract.Temperature, Headers: []string{"n-1", "θkū", "°C"}},

		{Name: "Heating days", Kind: contract.Count, Headers: []string{"n", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: contract.Energy, Headers: []string{"n", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: contract.Volume, Headers: []string{"n", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: contract.Temperature, Headers: []string{"n", "θkū", "°C"}},
	}

	BaselineInputDataRows = []contract.Row{ // using months like that in order to keep tabular outline
		{contract.Cell(time.Month(1).String()), "5", "10.1231", "45", "60.0", "11", "20", "105", "120", "20", "40", "200", "240"},
		{contract.Cell(time.Month(2).String()), "6", "11.3412", "44", "61.0", "10", "21", "30", "121", "21", "41", "201", "241"},
		{contract.Cell(time.Month(3).String()), "7", "12.0001", "43", "62.2", "12", "19", "102", "122", "22", "42", "202", "239"},
		{contract.Cell(time.Month(4).String()), "6", "13.0000", "42", "61.9", "15", "24", "101", "121", "21", "41", "201", "246"},
		{contract.Cell(time.Month(5).String()), "9", "14.1235", "41", "62.0", "12", "25", "102", "122", "22", "42", "202", "242"},
		{contract.Cell(time.Month(6).String()), "8", "15.0000", "40", "63.0", "13", "23", "103", "123", "23", "43", "203", "243"},
	}
)

func TestIndoorClima(t *testing.T) {
	db := models.NewTestGORM(t)
	store := NewIndoorClimaStore(db, validate)

	StoreTest{
		store: store,
		entity: &contract.IndoorClima{
			Project: NewTestProject(t, store).ID,
		},
		invalidate: func(valid models.Entity) models.Entity {
			var (
				ic      = valid.(*contract.IndoorClima)
				invalIC = *ic
			)

			invalIC.Project = uuid.Nil
			return &invalIC
		},
		update: func(doc *models.Document) models.Entity {
			var (
				ic     = doc.Data.(*contract.IndoorClima)
				copyIC = *ic
			)

			return &copyIC
		},
		duplicate: func(e models.Entity) models.Entity {
			var (
				ic     = e.(*contract.IndoorClima)
				copyIC = *ic
			)

			copyIC.Value.ID = uuid.Nil
			copyIC.Project = NewTestProject(t, store).ID

			return &copyIC
		},
		searchBy:   func(models.Entity) string { return "" },
		postCreate: func(models.Entity) error { return nil },
		memberUUID: func(*testing.T, models.Entity) uuid.UUID { return uuid.Nil },
	}.Run(t)
}

func TestICFields(t *testing.T) {
	db := models.NewTestGORM(t)

	store := NewIndoorClimaStore(db, validate)

	prj := NewTestProject(t, store)
	ic := &contract.IndoorClima{
		Project:                     prj.ID,
		AirexWindows:                period,
		AirexTotal:                  period,
		HeatedVolumeBuilding:        1.33,
		TotalEnergyConsumption:      period,
		TotalEConsumptionCircLosses: period,
		IndoorTemp:                  period,
	}

	tables := contract.NewTables()
	table, _ := contract.NewTable(BaselineInputDataColumns, BaselineInputDataRows...)
	tables["baseyear_n_2"] = table
	tables["baseyear_n_1"] = table
	tables["baseyear"] = table
	tables["baseline"] = table
	ic.Calculate(tables, 2, 1)

	doc, err := store.Create(ctx, ic)
	if err != nil {
		t.Errorf("error on saving: %v", err)
	}

	setPipes(ic)

	if _, err = store.Update(ctx, doc); err != nil {
		t.Errorf("error on updating: %v", err)
	}

	_, err = store.GetByIndex(ctx, prj.ID.String())
	if err != nil {
		t.Errorf("error on getting: %v", err)
	}
}

func setPipes(ic *contract.IndoorClima) {
	ic.BasementPipes = contract.JSONPipe{
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
	}

	ic.AtticPipes = contract.JSONPipe{
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
		pipe,
	}
}
