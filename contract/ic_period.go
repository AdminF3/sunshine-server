package contract

import (
	"database/sql/driver"
	"fmt"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

type PeriodType string

const (
	PeriodAirexWindows                PeriodType = "airex_windows"
	PeriodAirexTotal                  PeriodType = "airex_total"
	PeriodTotalEnergyConsumption      PeriodType = "total_energy_consumption"
	PeriodTotalEConsumptionCircLosses PeriodType = "total_energy_consumption_circulation"
	PeriodIndoorTemp                  PeriodType = "indoor_temp"
	PeriodOutdoorAirTemp              PeriodType = "outdoor_air_temp"
)

// Scan implements the database/sql.Scanner interface.
func (p *PeriodType) Scan(value interface{}) error {
	var v, ok = value.([]byte)
	if !ok {
		return fmt.Errorf("invalid period type: %v", v)
	}

	*p = PeriodType(v)
	return nil
}

// Value implements the database/sql/driver.Valuer interface.
func (p PeriodType) Value() (driver.Value, error) {
	return string(p), nil
}

type Period struct {
	N  float64 `json:"baseyear_n"`
	N1 float64 `json:"baseyear_n_1"`
	N2 float64 `json:"baseyear_n_2"`
}

type PeriodModel struct {
	models.Value
	Period

	IndoorClimaRef uuid.UUID  `json:"indoor_clima" gorm:"column:indoorclima_id"`
	Type           PeriodType `json:"type" gorm:"type"`
}

func (PeriodModel) TableName() string {
	return "ic_periods"
}

func (ic *IndoorClima) updatePeriod(t PeriodType, v Period) {
	for i, p := range ic.Periods {
		if p.Type == t {
			ic.Periods[i].Period = v
			return
		}
	}

	// No period with such type has been found.
	ic.Periods = append(ic.Periods, PeriodModel{
		Period:         v,
		IndoorClimaRef: ic.ID,
		Type:           t,
	})
}
