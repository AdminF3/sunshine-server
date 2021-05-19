package contract

import (
	"fmt"
	"math"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/shopspring/decimal"
)

const layout = "2006-01-02"

// Calculation holds the result of dynamic calculations done over a contract.
type Calculation struct {
	QTRef     [4]float64
	QApkRef   [4]float64
	QCzRef    [4]float64
	QApkCzRef [4]float64
	QKuRef    [4]float64
	QT1Ref    [4]float64
	GDDRef    [4]float64
	QIetG     float64
	QApkCzG   float64
	QMApkCzG  float64
	ETMFG     float64
	AMP       float64
	OM1       float64
}

// Calculate contract's dynamic data.
func calculate(contr *Contract, project models.Project) (calc *Calculation, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			debug.PrintStack()
		}
	}()
	var (
		tabl    Table
		summary [3]float64
	)

	calc = new(Calculation)
	asset := project.AssetSnapshot

	// --- baseconditions ---
	populateBasecondition(contr.Tables["baseconditions_n"], contr.Tables["baseyear_n"])
	populateBasecondition(contr.Tables["baseconditions_n_1"], contr.Tables["baseyear_n_1"])
	populateBasecondition(contr.Tables["baseconditions_n_2"], contr.Tables["baseyear_n_2"])

	// calculate QTRef
	by, by2, by3, err := qtrefTotals(contr)
	if err != nil {
		return nil, err
	}
	tabl = contr.Tables["baseline"]

	tabl.rows[0][3] = FormatFloat(by3)
	tabl.rows[0][4] = FormatFloat(by2)
	tabl.rows[0][5] = FormatFloat(by)

	// calculate GDD
	bc, bc2, bc3, err := gddTotals(contr)
	if err != nil {
		return nil, err
	}

	tabl.rows[6][3] = FormatFloat(bc3)
	tabl.rows[6][4] = FormatFloat(bc2)
	tabl.rows[6][5] = FormatFloat(bc)

	// calculate average indoor temp
	tn, tn1, tn2, err := avgIndoorTemp(contr)
	if err != nil {
		return nil, err
	}
	tabl.rows[5][3] = FormatFloat(tn2)
	tabl.rows[5][4] = FormatFloat(tn1)
	tabl.rows[5][5] = FormatFloat(tn)

	calc.QApkRef = referenceRow(tabl, 1, 3)
	calc.QCzRef = referenceRow(tabl, 2, 3)
	calc.QKuRef = referenceRow(tabl, 1, 3)
	calc.QApkCzRef = referenceRow(tabl, 4, 3)

	tabl.rows[3][3] = Cell(calcdhw(contr.Tables["baseyear_n_2"]))
	tabl.rows[3][4] = Cell(calcdhw(contr.Tables["baseyear_n_1"]))
	tabl.rows[3][5] = Cell(calcdhw(contr.Tables["baseyear_n"]))

	contr.Tables["baseline"] = tabl

	calc.QIetG = calc.QApkCzRef[3] * float64(project.GuaranteedSavings)
	calc.QApkCzG = calc.QApkCzRef[3] - calc.QIetG
	calc.QMApkCzG = calc.QApkCzG / 12

	om1, err := contr.Tables["operation_maintenance_budget"].Total(1)
	if err != nil {
		return nil, fmt.Errorf("OM1: %w", err)
	}
	calc.OM1, _ = om1.Float64()
	summary[2] = calc.OM1 / float64(asset.HeatedArea)

	tabl = contr.Tables["balancing_period_fee"]
	for i, row := range tabl.rows {
		var (
			a = rowFloat64(row, 1)
			b = rowFloat64(row, 2)
			c = a * b
			d = rowFloat64(row, 4)
			f = rowFloat64(row, 5)
		)
		row[1] = FormatFloat(a)
		row[2] = FormatFloat(b)
		row[3] = FormatFloat(c)
		row[4] = FormatFloat(d)
		row[5] = FormatFloat(f)
		row[6] = FormatFloat(f * b)

		if i == 0 {
			summary[0] = c
		}
	}

	tabl = contr.Tables["calc_energy_fee"]
	for _, row := range tabl.rows {
		var (
			a = rowFloat64(row, 1)
			b = rowFloat64(row, 2)
			c = a * b
			d = rowFloat64(row, 4)
			e = c / d
		)
		row[1] = FormatFloat(a)
		row[2] = FormatFloat(b)
		row[3] = FormatFloat(c)
		row[4] = FormatFloat(d)
		row[5] = FormatFloat(e)
	}
	contr.Tables["calc_energy_fee"] = tabl

	tabl = contr.Tables["operations_maintenance_fee"]
	for i := range tabl.rows {
		omy := rowFloat64(tabl.rows[i], 2)
		omm := omy / 12
		tabl.rows[i][1] = FormatFloat(rowFloat64(tabl.rows[i], 1))
		tabl.rows[i][2] = FormatFloat(omy)
		tabl.rows[i][3] = FormatFloat(omm)
		tabl.rows[i][4] = FormatFloat(float64(asset.HeatedArea))
		tabl.rows[i][5] = FormatFloat(omm / float64(asset.HeatedArea))
	}
	contr.Tables["operations_maintenance_fee"] = tabl

	prjDevRenovationSum, err := contr.Tables["project_development_renovations"].Total(2)
	if err != nil {
		return nil, err
	}

	constrCostsRenovationSum, err := contr.Tables["construction_costs_renovations"].Total(2)
	if err != nil {
		return nil, err
	}

	prjsupervisionSum, err := contr.Tables["project_supervision"].Total(2)
	if err != nil {
		return nil, err
	}

	finchargSum, err := contr.Tables["financial_charges"].Total(2)
	if err != nil {
		return nil, err
	}

	// total costs for renovation works
	ttcrw := prjDevRenovationSum.
		Add(constrCostsRenovationSum).
		Add(prjsupervisionSum).
		Add(finchargSum)
	ttcrwvat := ttcrw.Add(ttcrw.Mul(decimal.NewFromFloat(contr.VAT)))

	contr.Tables["renovation_financial_plan"].Row(4)[1] = Cell(ttcrwvat.Round(2).String())

	tbl := contr.Tables["renovation_overall_budget"]
	// Project development and management costs
	tbl.Row(0)[1] = Cell(prjDevRenovationSum.Round(2).String())

	// Construction costs
	tbl.Row(1)[1] = Cell(constrCostsRenovationSum.Round(2).String())

	// Project supervision costs
	tbl.Row(2)[1] = Cell(prjsupervisionSum.Round(2).String())

	// Financial charges
	tbl.Row(3)[1] = Cell(finchargSum.Round(2).String())

	// annex 7 | project_measurements_table
	tbl = contr.Tables["project_measurements_table"]

	bbal := contr.Fields["contractor_fin_contribution"]
	intr := contr.Fields["interest_rate_percent"]
	for i := range tbl.rows {
		payment := string(tbl.Row(i)[3])
		if payment == "" {
			// do not calculate further, because payments
			// need to be added.
			break
		}

		dd := convdate(contr.Fields["start_date_of_loan"])
		dd = dd.AddDate(0, i, 0)
		tbl.Row(i)[1] = Cell(dd.String())

		// beginning balance
		tbl.Row(i)[2] = Cell(bbal)

		// ending balance
		eb := endingBalance(bbal, payment, intr, contr.Eurobor)
		tbl.Row(i)[4] = Cell(eb)

		bbal = eb

		if i == 0 {
			f, _ := strconv.ParseFloat(payment, 64)
			summary[1] = f
		}
	}

	sumtbl := contr.Tables["summary"]
	for i := range sumtbl.Rows() {
		sum := summary[i]
		sumtbl.Row(i)[1] = FormatFloat(sum)
		sumtbl.Row(i)[2] = FormatFloat(sum * contr.VAT)
		sumtbl.Row(i)[3] = FormatFloat(sum + contr.VAT*sum)
	}

	return calc, nil
}

// calcdhw calculate Domestic Hot Water row based in the `baseline`:
// Qku,ref = (V * (Oku - Tw) * 4186 * Ro)/3600
//
// Where:
//      V is the DHW consumption in m3,
//      Oku - DHW temperature from tables 1, 2 ,3,
//	Tw - cold water temperature (10 °C) constant,
//	Ro - specific density of water (roughly 1000 kg/m3 ==> 999.7), constant,
func calcdhw(bytabl Table) string {
	tw := decimal.NewFromInt(10)
	ro := decimal.NewFromInt(1000)
	c := decimal.NewFromInt(4186)
	div := decimal.NewFromInt(3600)

	v := decimal.NewFromInt(0)
	oku := decimal.NewFromInt(0)

	for i := range bytabl.Rows() {
		v = v.Add(bytabl.Row(i).Cell(3).Decimal())
		oku = oku.Add(bytabl.Row(i).Cell(4).Decimal())
	}

	res := oku.Sub(tw).Mul(v).Mul(c).Mul(ro).Div(div).Round(2)

	return res.String()
}

// populateBasecondition in the basecondition tables
func populateBasecondition(bctabl, bytabl Table) {
	for i := range bytabl.Rows() {
		// heating days
		ht := bytabl.Row(i)[1]
		bctabl.Row(i)[1] = ht

		// GDD - degree days
		dapk := bctabl.Row(i)[1].Decimal()
		t1 := bctabl.Row(i)[2].Decimal()
		t3 := bctabl.Row(i)[3].Decimal()

		// gdd = dapk(T3-t1)
		gdd := t3.Sub(t1).Mul(dapk).Round(2).String()
		bctabl.Row(i)[4] = Cell(gdd)
	}
}

// Ending Balance      = beginning_balance - principal (principal: payment - (interest+eurobor))
func endingBalance(beginningBalance, payment, interest string, eurobor decimal.Decimal) string {
	bbal, _ := decimal.NewFromString(beginningBalance)
	p, _ := decimal.NewFromString(payment)
	irate, _ := decimal.NewFromString(interest)
	irate = irate.Add(eurobor)

	hundred := decimal.NewFromInt(100)
	tw := decimal.NewFromInt(12)

	// beginning balance * interest_rate (in %) per year
	interestPerYear := bbal.Mul(irate).Div(hundred).Div(tw)
	principal := p.Sub(interestPerYear)

	res := bbal.Sub(principal)
	if res.IsNegative() {
		res = decimal.Zero
	}
	return res.Round(2).String()
}

func convdate(s string) time.Time {
	date := strings.Split(s, " ")[0]
	d, _ := time.Parse(layout, date)

	return d

}

func rowFloat64(row Row, i int) float64 {
	if i < 0 || i >= len(row) {
		return 0
	}

	var f, _ = row[i].Decimal().Float64()
	if math.IsNaN(f) {
		f = 0
	}
	return f
}

func referenceRow(t Table, row, from int) [4]float64 {
	if row >= t.Len() {
		panic(fmt.Sprintf("table has %d rows, asked for index %d", t.Len(), row))
	}

	var (
		cells = t.rows[row][from : from+4]
		f     = func(d decimal.Decimal) float64 {
			var f, _ = d.Float64()
			return f
		}
	)

	return [4]float64{
		f(cells[0].Decimal()),
		f(cells[1].Decimal()),
		f(cells[2].Decimal()),
		f(cells[3].Decimal()),
	}
}

func qtrefTotals(contr *Contract) (float64, float64, float64, error) {
	b1, err := contr.Tables["baseyear_n"].Total(2)
	if err != nil {
		return 0, 0, 0, err
	}
	by, _ := b1.Float64()
	b2, err := contr.Tables["baseyear_n_1"].Total(2)
	if err != nil {
		return 0, 0, 0, err
	}
	by2, _ := b2.Float64()
	b3, err := contr.Tables["baseyear_n_2"].Total(2)
	if err != nil {
		return 0, 0, 0, err
	}
	by3, _ := b3.Float64()

	return by, by2, by3, nil
}

func gddTotals(contr *Contract) (float64, float64, float64, error) {
	bc, err := contr.Tables["baseconditions_n"].Total(4)
	if err != nil {
		return 0, 0, 0, err
	}
	gdd, _ := bc.Float64()
	bc1, err := contr.Tables["baseconditions_n_1"].Total(4)
	if err != nil {
		return 0, 0, 0, err
	}
	gdd1, _ := bc1.Float64()
	bc2, err := contr.Tables["baseconditions_n_2"].Total(4)
	if err != nil {
		return 0, 0, 0, err
	}
	gdd2, _ := bc2.Float64()
	return gdd, gdd1, gdd2, nil
}

func avgIndoorTemp(contr *Contract) (float64, float64, float64, error) {
	t, err := contr.Tables["baseconditions_n"].Average(3)
	if err != nil {
		return 0, 0, 0, err
	}
	t1, _ := t.Float64()
	tt, _ := contr.Tables["baseconditions_n_1"].Average(3)
	if err != nil {
		return 0, 0, 0, err
	}
	t2, _ := tt.Float64()
	ttt, _ := contr.Tables["baseconditions_n_2"].Average(3)
	if err != nil {
		return 0, 0, 0, err
	}
	t3, _ := ttt.Float64()
	return t1, t2, t3, nil
}
