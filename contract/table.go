package contract

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// Table represents an arrangement of data in rows and columns.
//
// Each column, identified by name, defines what kind of data is supposed
// to be stored in cells underneath it. A row is a n-tuple where n is the
// number of columns and the data kind of each record is determined by the
// matching column. The intersection of a column and a row is a cell. Once
// defined table columns cannot be added, removed or changed.
//
// Implements sort.Interface for sorting rows. See SetSortByColumn method.
type Table struct {
	columns []Column
	rows    []Row
	sortby  int
	title   string
}

// NewTable creates new table value with given columns and rows.
//
// Once created a table the columns are considered immutable and each row is
// supposed to have `len(columns)` number of cells.
//
// If any rows are passed they're being added to the table until an invalid one
// occurs (if any) which returns a non-nil error.
func NewTable(columns []Column, rows ...Row) (Table, error) {
	var (
		t = Table{
			rows:    make([]Row, 0, len(rows)),
			columns: columns,
		}
		err error
	)

	for _, row := range rows {
		err = t.AddRow(row)
		if err != nil {
			break
		}
	}

	return t, err
}

// AddRow adds a new row at the end of the receiver if it has the same number
// or cells as required (i.e. count of table columns).
func (t *Table) AddRow(r Row) error {
	if len(r) != len(t.columns) {
		return errWrongCellCount
	}

	t.rows = append(t.rows, r)
	return nil
}

// Empty returns new table with the same columns as receiver but without any
// rows whatsoever.
func (t Table) Empty() *Table {
	return &Table{
		columns: t.columns,
		rows:    make([]Row, 0),
	}
}

// Row returns Nth row from the receiver, starting from 0. If invalid n
// is given it panics. Consult with .Len() before using this.
func (t Table) Row(n int) Row {
	return t.rows[n]
}

// Columns returns a copy of all table columns.
func (t Table) Columns() []Column {
	var result = make([]Column, len(t.columns))
	copy(result, t.columns)
	return result
}

// Headers returns a copy of all column headers as rows.
func (t Table) Headers() []Row {
	var result = make([]Row, t.HeaderLen())
	for i := 0; i < len(result); i++ {
		var row = make(Row, t.ColumnLen())
		for j := 0; j < t.ColumnLen(); j++ {
			c := t.Column(j)
			if len(c.Headers) > i {
				row[j] = Cell(c.Headers[i])
			}
		}
		result[i] = row
	}
	return result
}

// Rows returns a copy of all table rows.
func (t Table) Rows() []Row {
	var result = make([]Row, len(t.rows))
	copy(result, t.rows)
	return result
}

// Column returns Nth column from the receiver, starting from 0. If invalid n
// is given it panics. Consult with .ColumnLen() before using this.
func (t Table) Column(n int) Column {
	return t.columns[n]
}

// Total calculates a sum of cells in given column.
func (t Table) Total(column int) (decimal.Decimal, error) {
	return t.calc(column, CanSum, decimal.Sum)
}

// Average calculates an average value of cells in given column.
func (t Table) Average(column int) (decimal.Decimal, error) {
	avg, err := t.calc(column, CanAvg, decimal.Avg)
	return avg.Round(2), err
}

// Title returns title of the table.
func (t Table) Title() string {
	return t.title
}

// SetTitle sets the table title.
func (t *Table) SetTitle(title string) {
	t.title = title
}

type canFunc func(c Column) bool
type calcFunc func(decimal.Decimal, ...decimal.Decimal) decimal.Decimal

func (t Table) calc(i int, can canFunc, do calcFunc) (decimal.Decimal, error) {
	if 0 > i || i >= t.ColumnLen() {
		return zero, errBadColumnIndex
	}

	var cells = make([]decimal.Decimal, len(t.rows))
	for j, row := range t.rows {
		cells[j] = row[i].Decimal()
	}

	if !can(t.columns[i]) {
		return zero, errInvalidAction
	}

	return do(cells[0], cells[1:]...), nil
}

// JSONTable is just a plain JSON representation of Table.
type JSONTable struct {
	Columns []Column `json:"columns"`
	Rows    []Row    `json:"rows"`
	Title   string   `json:"title"`
}

func (t Table) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONTable{
		Columns: t.columns,
		Rows:    t.rows,
		Title:   t.title,
	})
}

func (t *Table) UnmarshalJSON(data []byte) error {
	var (
		table = JSONTable{}
		err   = json.Unmarshal(data, &table)
	)

	t.columns = table.Columns
	t.rows = table.Rows
	t.title = table.Title
	return err
}

// HeaderLen is the number of most headers in a column.
func (t Table) HeaderLen() int {
	var max int

	for _, c := range t.columns {
		if len(c.Headers) > max {
			max = len(c.Headers)
		}
	}

	return max
}

// ColumnLen is the number of columns in a table.
func (t Table) ColumnLen() int {
	return len(t.columns)
}

// Len is the number of rows in a table.
func (t Table) Len() int {
	return len(t.rows)
}

// Swap swaps rows with indexes i and j.
func (t *Table) Swap(i, j int) {
	t.rows[i], t.rows[j] = t.rows[j], t.rows[i]
}

// Less reports whether the rows with index i should sort before the row with
// index j, using the chosen column. If the column is holding numerical kind of
// values, less compares them by value. Otherwise lexicographical less-than
// comparison is used.
//
// It is unsafe to call this method with out of range i and/or j. However, the
// sort package makes sure not to do that by invoking Len() first. If invalid
// column is chosen for sorting by this always returns false, thus
// preserving the current row order.
func (t Table) Less(i, j int) bool {
	var (
		cmp    int
		irow   = t.rows[i].Cell(t.sortby)
		jrow   = t.rows[j].Cell(t.sortby)
		column = t.Column(t.sortby)
	)

	if column.Kind.IsNumber() {
		cmp = irow.Decimal().Cmp(jrow.Decimal())
	} else {
		cmp = strings.Compare(string(irow), string(jrow))
	}

	return cmp == -1
}

// SetSortByColumn marks the default column to sort by. If invalid column index
// is given the state doesn't change and the method returns false.
func (t *Table) SetSortByColumn(n int) bool {
	if n < 0 || n >= t.ColumnLen() {
		return false
	}
	t.sortby = n
	return true
}

// Join several tables into one using the first one as a template. Join of zero
// tables gives an empty table.
//
// Template table is important because:
//
// - Name column is omitted from all tables but the template.
//
// - The result table has as much rows as the template.
//
// - If there's a table with less rows than the template empty cells are
// inserted there.
//
// - If there's a table with more rows than the template they are discarded.
func Join(tables ...Table) Table {
	if len(tables) == 0 {
		return Table{}
	}

	var (
		from int
		tmpl = tables[0]

		// assume that all tables have the same amount of columns.
		columns = make([]Column, 0, tmpl.ColumnLen()*len(tables))
		rows    = make([]Row, tmpl.Len())
	)

	for i, t := range tables {
		columns = append(columns, t.columns[from:]...)

		// omit first column on all tables after the first one.
		if i == 0 {
			from = 1
		}

	}

	for i := range tmpl.rows {
		var row = make(Row, 0, len(columns))
		for j, t := range tables {
			if j == 0 {
				from = 0
			} else {
				from = 1
			}
			var subrow Row
			if t.Len() > i {
				subrow = t.rows[i][from:]
			} else {
				// this table is smaller than the template, so
				// just append empty cells.
				subrow = make(Row, t.ColumnLen()-1)
			}
			row = append(row, subrow...)
		}
		rows[i] = row
	}

	return Table{
		columns: columns,
		rows:    rows,
		sortby:  tables[0].sortby,
	}

}

// Cell holds a value in table cell and can be represented as any kind.
type Cell string

// String returns a plain string value of cell.
func (c Cell) String() string {
	return string(c)
}

// FormatFloat converts the floating-point number f to a Cell.
func FormatFloat(f float64) Cell {
	if math.IsNaN(f) {
		f = 0
	}
	return Cell(strconv.FormatFloat(f, 'f', 2, 64))
}

// Format cell's value as given kind.
func (c Cell) Format(k Kind) string {
	var (
		unit  string
		value = c.Decimal().Round(4)
	)

	switch k {
	case Decimal:
		// .Decimal() is called anyways.
	case Money:
		unit = "€"
	case Temperature:
		unit = "°C"
	case Volume:
		unit = "m³"
	case Energy:
		unit = "MWh"
	case Count:
		value = value.Round(0)
	case Area:
		unit = "m²"
	case Percent:
		unit = "%"
	default:
		return c.String()
	}

	return value.String() + unit
}

// Decimal returns a decimal number value.
//
// If no numerical value is stored in the receiver this returns zero decimal
// value.
func (c Cell) Decimal() decimal.Decimal {
	var d, err = decimal.NewFromString(string(c))
	if err != nil {
		return zero
	}

	return d
}

// Row in a table is just a slice of cells.
type Row []Cell

// Cell is a safe method for getting the Nth cell, starting from 0. If invalid
// n is given, returns empty cell.
func (r Row) Cell(n int) Cell {
	if n < 0 || n >= len(r) {
		return ""
	}

	return r[n]
}

// Column is identified by name and defines the data supposed to be stored in
// cells underneath it.
//
// Apart from name and meta-information stored (and possibly displayed) as
// headers, it describes what kind of data is supposed to be stored and
// what calculations are meaningful to be made with values in this column.
type Column struct {
	Name    string   `json:"name"`
	Kind    Kind     `json:"kind"`
	Headers []string `json:"headers"`
}

// CanSum reports whether it makes sense to sum cells of this column.
//
// The result is not based just on kind's format, but on real-world
// application. Temperature values for instance are stored as decimals and
// therefore could be summed, but it doesn't really make sense to report the
// total temperature for given period of time.
func CanSum(c Column) bool {
	switch c.Kind {
	case Decimal, Money, Count, Energy, Volume:
		return true
	default:
		return false
	}
}

// CanAvg reports whether it makes sense to calculate average cell value
// of this column.
//
// The result is not based just on kind's format, but on real-world
// application. Heating days for instance are stored as numbers and therefore
// it's possible to calculate average value of them, but it doesn't really make
// sense to report the average heating days for given period of time. Thus
// given current requirements, it makes sense to calculate only on temperature
// values.
func CanAvg(c Column) bool {
	return c.Kind == Temperature
}

// Kind represents the specific kind of value that a cell holds. The zero
// kind is a valid plain-text value.
type Kind uint

const (
	// String is plain-text value without any conversion whatsoever.
	String = iota

	// Name is generally the first column used as row description.
	Name

	// Decimal represents a fixed-point decimal (value * 10 ^ exp).
	Decimal

	// Money is decimal value of the official currency in EU (€).
	Money

	// Temperature is decimal value of temperature in degrees Celsius (°C).
	Temperature

	// Count is integer value representing count of something (e.g. days).
	Count

	// Volume is decimal value of volume in cubic meters (m³).
	Volume

	// Energy is decimal value of power consumption in megawatts multiplied
	// by the time in hours (MWh).
	Energy

	// Area is decimal value of are in square meters (m²)
	Area

	// Percent is decimal value of some percentage value
	Percent
)

// IsNumber reports whether kind is numerical value.
func (k Kind) IsNumber() bool {
	return k > 1
}

var (
	zero = decimal.Decimal{}

	errBadColumnIndex = errors.New("column index out of range")
	errWrongCellCount = errors.New("wrong cell count")
	errInvalidAction  = errors.New("invalid action on this column")
)
