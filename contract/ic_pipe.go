package contract

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type (
	JSONPipe []Pipe
	calc     func(diameter float64, pipeQuality Quality) float64
)

type Pipe struct {
	Quality         Quality `json:"quality"`
	InstalledLength float64 `json:"installed_length"`
	Diameter        float64 `json:"diameter"`
	HeatLossUnit    float64 `json:"heat_loss_unit"`
	HeatLossYear    float64 `json:"heat_loss_year"`
}

type Quality int8

const (
	_ Quality = iota
	QualityGood
	QualityPoor
	QualityNo
)

func (p *Pipe) CalcQuality(c calc) {
	if p.Diameter <= 0.0 {
		return
	}
	p.HeatLossUnit = c(p.Diameter, p.Quality)
	p.HeatLossYear = p.HeatLossUnit * p.InstalledLength
}

func heatLossAttic(d float64, q Quality) float64 {
	return heatLoss(d, q+2)
}

func heatLossBasement(d float64, q Quality) float64 {
	return heatLoss(d, q-1)
}

func heatLoss(d float64, row Quality) float64 {
	var (
		col    int8
		values = [6][6]float64{
			// basement
			// Good
			{0.026, 0.039, 0.05, 0.053, 0.057, 0.066},
			// Poor
			{0.033, 0.048, 0.062, 0.064, 0.07, 0.081},
			// No
			{0.069, 0.183, 0.274, 0.366, 0.411, 0.503},

			// attic
			//Good
			{0.031, 0.046, 0.059, 0.063, 0.068, 0.078},
			// Poor
			{0.039, 0.057, 0.074, 0.076, 0.083, 0.096},
			// No
			{0.081, 0.217, 0.326, 0.434, 0.489, 0.597},
		}
	)

	switch {
	case d < 30:
		col = 0
	case d < 50:
		col = 1
	case d < 70:
		col = 2
	case d < 90:
		col = 3
	case d < 110:
		col = 4
	default:
		col = 5
	}

	return values[row][col]
}

func (p JSONPipe) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *JSONPipe) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}

func (p *JSONPipe) UnmarshalGQL(v interface{}) error {
	switch b := v.(type) {
	case []byte:
		return json.Unmarshal(b, p)
	case string:
		return json.Unmarshal([]byte(b), p)
	default:
		return fmt.Errorf("%[1]T(%[1]v) is not json pipe", v)
	}
}

func (p JSONPipe) MarshalGQL(w io.Writer) {
	json.NewEncoder(w).Encode(p)
}
