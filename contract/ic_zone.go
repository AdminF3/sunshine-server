package contract

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type JSONZone map[string]ZoneModel

type ZoneModel struct {
	Area          float64 `json:"area"`
	UValue        float64 `json:"uvalue"`
	TempDiff      Period  `json:"tempdiff"`
	HeatLossCoeff float64 `json:"heat_loss_coeff"`
}

// calc calculates the values that are not inputs.
// Ti - indoor temperature
func (m *ZoneModel) calc(Tin, Tout Period, mtype string) {
	m.HeatLossCoeff = m.UValue * m.Area

	if strings.Contains(mtype, "zone1") {
		// zone 1 calculation
		m.TempDiff.N = Tin.N - Tout.N
		m.TempDiff.N1 = Tin.N1 - Tout.N1
		m.TempDiff.N2 = Tin.N2 - Tout.N2
	} else {
		// zone 2 calculation
		m.TempDiff.N = (Tin.N - Tout.N) - 3
		m.TempDiff.N1 = (Tin.N1 - Tout.N1) - 3
		m.TempDiff.N2 = (Tin.N2 - Tout.N2) - 3
	}
}

func (z JSONZone) Value() (driver.Value, error) {
	return json.Marshal(z)
}

func (z *JSONZone) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &z)
}

func (z *JSONZone) UnmarshalGQL(v interface{}) error {
	switch b := v.(type) {
	case []byte:
		return json.Unmarshal(b, z)
	case string:
		return json.Unmarshal([]byte(b), z)
	default:
		return fmt.Errorf("%[1]T(%[1]v) is not json zone", v)
	}
}

func (z JSONZone) MarshalGQL(w io.Writer) {
	json.NewEncoder(w).Encode(z)
}
