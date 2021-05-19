package contract

import (
	"fmt"
	"strconv"
	"strings"

	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

type IndoorClima struct {
	models.Value

	Project uuid.UUID `json:"project" gorm:"column:project_id"`

	// Basement Pipes fields
	BasementPipes JSONPipe `json:"basement_pipes"`
	AtticPipes    JSONPipe `json:"attic_pipes"`

	// Totals and indoor temp Ht
	TotalHT float64 `json:"total_ht" gorm:"-"`

	// Φiek - Heat gains, internal
	HeatGainsInternal float64 `json:"heatgains_internal" gorm:"-"`
	// Φsol - Heat gains, solar
	HeatGainsSolar float64 `json:"heatgains_solar" gorm:"-"`

	// This is percentage value. It should be in `NN`
	// For example `67`
	AirexWindows Period `json:"airex_windows" gorm:"-"`
	// ACH
	AirexBuilding Period `json:"airex_building" gorm:"-"`
	// Hve -Total 	air exchange
	AirexTotal Period `json:"airex_total" gorm:"-"`

	// Heated Volume of the building Vapk [m^3]
	HeatedVolumeBuilding float64 `json:"heated_volume_building" gorm:"column:heated_volume_building"`

	TotalEnergyConsumption Period `json:"total_energy_consumption" gorm:"-"`

	// TODO hot to get thta
	// Maybe from contract table `Space heating and circulation losses | Total`
	TotalEConsumptionCircLosses Period `json:"total_energy_consumption_circulation" gorm:"-"`

	// Circulation losses during the period when the space heating
	// system is not operated - Qcz,ex
	// TODO get it from project
	// Maybe from sum of may-september `Ciculation losses`
	CirculationLosses Period `json:"circulation_losses" gorm:"-"`

	// Distribution losses through space heating pipes installed
	// in the basement, Qcz, p
	DistributionLossesBasement float64 `json:"distribution_losses_basement" gorm:"-"`
	// Distribution losses through space heating pipes installed
	// in the technical attic, Qcz, b
	DistributionLossesAttic float64 `json:"distribution_losses_attic" gorm:"-"`

	TotalMeasured   Period `json:"total_measured" gorm:"-"`
	TotalCalculated Period `json:"total_calculated" gorm:"-"`
	IndoorTemp      Period `json:"indoor_temp" gorm:"-"`

	// General data
	OutdoorAirTemp Period `json:"ourdoor_air_temp" gorm:"-"`

	// GORM
	Periods []PeriodModel `json:"-" gorm:"foreignkey:IndoorClimaRef"`
	Zones   JSONZone      `json:"-"`
}

// AfterFind is GORM hook which is getting called right after SELECT queries.
func (ic *IndoorClima) AfterFind() {
	for _, p := range ic.Periods {
		switch p.Type {
		case PeriodAirexWindows:
			ic.AirexWindows = p.Period
		case PeriodAirexTotal:
			ic.AirexTotal = p.Period
		case PeriodTotalEnergyConsumption:
			ic.TotalEnergyConsumption = p.Period
		case PeriodTotalEConsumptionCircLosses:
			ic.TotalEConsumptionCircLosses = p.Period
		case PeriodIndoorTemp:
			ic.IndoorTemp = p.Period
		case PeriodOutdoorAirTemp:
			ic.OutdoorAirTemp = p.Period
		default:
			panic(fmt.Sprintf("Unexpected period type: %#v", p))
		}
	}

	for k, v := range ic.Zones {
		v.calc(ic.IndoorTemp, ic.OutdoorAirTemp, k)
		ic.Zones[k] = v
	}
}

// BeforeSave is GORM hook which is getting called right before CREATE and
// UPDATE queries.
func (ic *IndoorClima) BeforeSave() {
	ic.updatePeriod(PeriodAirexWindows, ic.AirexWindows)
	ic.updatePeriod(PeriodAirexTotal, ic.AirexTotal)
	ic.updatePeriod(PeriodTotalEnergyConsumption, ic.TotalEnergyConsumption)
	ic.updatePeriod(PeriodTotalEConsumptionCircLosses, ic.TotalEConsumptionCircLosses)
	ic.updatePeriod(PeriodIndoorTemp, ic.IndoorTemp)
	ic.updatePeriod(PeriodOutdoorAirTemp, ic.OutdoorAirTemp)
}

func (p Period) BaseYear(n int) float64 {
	switch n {
	case 0:
		return p.N
	case 1:
		return p.N1
	case 2:
		return p.N2
	}

	return 0
}

func NewIndoorClima(id uuid.UUID) *IndoorClima {
	return &IndoorClima{
		Project: id,
		Zones:   make(JSONZone),
	}
}

func (IndoorClima) Kind() string {
	return "indoorclima"
}

func (ic IndoorClima) Key() string {
	return ic.Project.String()
}

func (ic IndoorClima) Dependencies() []config.Dependency {
	return []config.Dependency{}
}

// Calculate the Indoor temperature according the formula in the
// LABEEF FINANCIAL AND TECHNICAL RULES AND GUIDELINES FOR ENERGY
// EFFICIENCY MEASURES:
//	QApk,cz – Qcz,ex – Qcz,p – Qcz,b = [(HT+Hve)⨯(T1 –T3) – Φiek – Φsol] ⨯ DApk ⨯ 24 x 10-6
//	T1 - average indoor temperature for the baseyear (from Annex III)
//	T3 - daily average outdoor temperature for the baseyear
//	DApk - heating days (from annex III)
func (ic *IndoorClima) Calculate(tables Tables, Aapk, floors float64) {
	cn2, _ := tables["baseyear_n_2"].Total(0)
	dapkBaseYearN2, _ := cn2.Float64()

	cn1, _ := tables["baseyear_n_1"].Total(0)
	dapkBaseYearN1, _ := cn1.Float64()

	cn, _ := tables["baseyear"].Total(0)
	dapkBaseYear, _ := cn.Float64()

	t1BaseYear, _ := strconv.ParseFloat(tables["baseline"].Row(5).Cell(5).String(), 64)
	t1BaseYearN1, _ := strconv.ParseFloat(tables["baseline"].Row(5).Cell(4).String(), 64)
	t1BaseYearN2, _ := strconv.ParseFloat(tables["baseline"].Row(5).Cell(3).String(), 64)

	ic.IndoorTemp.N = t1BaseYear
	ic.IndoorTemp.N1 = t1BaseYearN1
	ic.IndoorTemp.N2 = t1BaseYearN2

	heatingLoss, _ := strconv.ParseFloat(tables["baseline"].Row(4).Cell(5).String(), 64)
	heatingLossN1, _ := strconv.ParseFloat(tables["baseline"].Row(4).Cell(4).String(), 64)
	heatingLossN2, _ := strconv.ParseFloat(tables["baseline"].Row(4).Cell(3).String(), 64)

	ic.TotalEConsumptionCircLosses.N = heatingLoss
	ic.TotalEConsumptionCircLosses.N1 = heatingLossN1
	ic.TotalEConsumptionCircLosses.N2 = heatingLossN2

	qcz, _ := strconv.ParseFloat(tables["baseline"].Row(2).Cell(5).String(), 64)
	qczn1, _ := strconv.ParseFloat(tables["baseline"].Row(2).Cell(4).String(), 64)
	qczn2, _ := strconv.ParseFloat(tables["baseline"].Row(2).Cell(3).String(), 64)

	ic.HeatGainsInternal = 6 * Aapk

	var totalWindowArea float64

	for k, v := range ic.Zones {
		if strings.Contains(k, "window") {
			totalWindowArea += v.Area
		}
		ic.TotalHT += v.HeatLossCoeff
	}
	ic.HeatGainsSolar = 12.8 * totalWindowArea

	Qt, _ := tables["baseyear"].Total(2)
	Qtn1, _ := tables["baseyear_n_1"].Total(2)
	Qtn2, _ := tables["baseyear_n_2"].Total(2)

	Qtf, _ := Qt.Float64()
	Qtf1, _ := Qtn1.Float64()
	Qtf2, _ := Qtn2.Float64()

	ic.TotalEnergyConsumption.N = Qtf
	ic.TotalEnergyConsumption.N1 = Qtf1
	ic.TotalEnergyConsumption.N2 = Qtf2

	ic.CirculationLosses.N = qcz
	ic.CirculationLosses.N1 = qczn1
	ic.CirculationLosses.N2 = qczn2

	ic.calcTHL()

	ic.AirexBuilding.N = (-0.0083*floors-0.175)*(ic.AirexWindows.N/100) + 0.0167*floors + 0.55
	ic.AirexBuilding.N1 = (-0.0083*floors-0.175)*(ic.AirexWindows.N1/100) + 0.0167*floors + 0.55
	ic.AirexBuilding.N2 = (-0.0083*floors-0.175)*(ic.AirexWindows.N2/100) + 0.0167*floors + 0.55

	ic.AirexTotal.N = 0.34 * ic.AirexBuilding.N * ic.HeatedVolumeBuilding
	ic.AirexTotal.N1 = 0.34 * ic.AirexBuilding.N1 * ic.HeatedVolumeBuilding
	ic.AirexTotal.N2 = 0.34 * ic.AirexBuilding.N2 * ic.HeatedVolumeBuilding

	ic.TotalCalculated.N = ic.calcBaseYear(0, dapkBaseYear, t1BaseYear)
	ic.TotalCalculated.N1 = ic.calcBaseYear(1, dapkBaseYearN1, t1BaseYearN1)
	ic.TotalCalculated.N2 = ic.calcBaseYear(2, dapkBaseYearN2, t1BaseYearN2)

	ic.TotalMeasured.N = ic.TotalEConsumptionCircLosses.N -
		ic.CirculationLosses.N -
		ic.DistributionLossesBasement -
		ic.DistributionLossesAttic
	ic.TotalMeasured.N1 = ic.TotalEConsumptionCircLosses.N1 -
		ic.CirculationLosses.N1 -
		ic.DistributionLossesBasement -
		ic.DistributionLossesAttic
	ic.TotalMeasured.N2 = ic.TotalEConsumptionCircLosses.N2 -
		ic.CirculationLosses.N2 -
		ic.DistributionLossesBasement -
		ic.DistributionLossesAttic
}

func (ic *IndoorClima) calcTHL() {
	ic.DistributionLossesBasement = 0
	ic.DistributionLossesAttic = 0

	for k, v := range ic.BasementPipes {
		v.CalcQuality(heatLossBasement)
		ic.BasementPipes[k] = v

		ic.DistributionLossesBasement += ic.BasementPipes[k].HeatLossYear
	}

	for k, v := range ic.AtticPipes {
		v.CalcQuality(heatLossAttic)
		ic.AtticPipes[k] = v

		ic.DistributionLossesAttic += ic.AtticPipes[k].HeatLossYear
	}
}

func (ic *IndoorClima) calcBaseYear(year int, dapk, t1 float64) float64 {
	var (
		tHT    = ic.TotalHT
		airexT = ic.AirexTotal.BaseYear(year)

		// T3 - actual average outdoor temperature during the Baseline Period (oC).
		t3     = ic.OutdoorAirTemp.BaseYear(year)
		heatGI = ic.HeatGainsInternal
		heatGS = ic.HeatGainsSolar
	)

	return ((tHT+airexT)*(t1-t3) - heatGI - heatGS) * dapk * (24*10 ^ -6)
}

func (IndoorClima) TableName() string {
	return "indoor_climas"
}
