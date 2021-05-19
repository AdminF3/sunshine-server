package contract

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

var (
	baselineInputDataColumns = []Column{
		{Name: "Month", Kind: Name, Headers: []string{"", "Symbol", "Unit"}},

		{Name: "Heating days", Kind: Count, Headers: []string{"n-2", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: Energy, Headers: []string{"n-2", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: Volume, Headers: []string{"n-2", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: Temperature, Headers: []string{"n-2", "θkū", "°C"}},

		{Name: "Heating days", Kind: Count, Headers: []string{"n-1", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: Energy, Headers: []string{"n-1", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: Volume, Headers: []string{"n-1", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: Temperature, Headers: []string{"n-1", "θkū", "°C"}},

		{Name: "Heating days", Kind: Count, Headers: []string{"n", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: Energy, Headers: []string{"n", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: Volume, Headers: []string{"n", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: Temperature, Headers: []string{"n", "θkū", "°C"}},
	}

	baselineInputDataRows = []Row{ // using months like that in order to keep tabular outline
		{Cell(time.Month(1).String()), "5", "10.1231", "45", "60.0", "11", "20", "105", "120", "20", "40", "200", "240"},
		{Cell(time.Month(2).String()), "6", "11.3412", "44", "61.0", "10", "21", "30", "121", "21", "41", "201", "241"},
		{Cell(time.Month(3).String()), "7", "12.0001", "43", "62.2", "12", "19", "102", "122", "22", "42", "202", "239"},
		{Cell(time.Month(4).String()), "6", "13.0000", "42", "61.9", "15", "24", "101", "121", "21", "41", "201", "246"},
		{Cell(time.Month(5).String()), "9", "14.1235", "41", "62.0", "12", "25", "102", "122", "22", "42", "202", "242"},
		{Cell(time.Month(6).String()), "8", "15.0000", "40", "63.0", "13", "23", "103", "123", "23", "43", "203", "243"},
	}
)

func TestNewTable(t *testing.T) {
	var cases = []struct {
		name    string
		columns []Column
		pre     []Row
		post    []Row
		preerr  bool
		posterr []bool
	}{
		{name: "empty"},
		{
			name:   "empty bad pre",
			pre:    []Row{[]Cell{}, []Cell{"foo"}},
			preerr: true,
		},
		{
			name:    "empty bad post",
			post:    []Row{[]Cell{}, []Cell{"foo"}},
			posterr: []bool{false, true},
		},
		{
			name:    "Baseline input data",
			columns: baselineInputDataColumns,
			pre:     baselineInputDataRows[0:3],
			post:    baselineInputDataRows[3:],
			posterr: []bool{false, false, false},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var (
				table, err = NewTable(c.columns, c.pre...)
				experr     error
			)

			if c.preerr {
				experr = errWrongCellCount
			}

			if err != experr {
				t.Fatalf("Expected %q on NewTable, got %q", experr, err)
			}

			if c.preerr {
				if table.Len() == len(c.pre) {
					t.Fatalf("All rows got added even though error occured")
				}
			} else if len(c.pre) > 0 {
				if table.Len() != len(c.pre) {
					t.Errorf("Expected %d, got %d", len(c.pre), table.Len())
				}
				for i := 0; i < table.Len(); i++ {
					tableGetRowCmp(t, &table, i, c.pre[i])
				}
			}

			for i, row := range c.post {
				var experr error
				if c.posterr[i] {
					experr = errWrongCellCount
				}

				if err := table.AddRow(row); err != experr {
					t.Errorf("Expected %q on %dth AddRow, got %q",
						experr, i, err)
				}

				if !c.posterr[i] {
					tableGetRowCmp(t, &table, i+len(c.pre), row)
				}
			}

			empty := table.Empty()
			if empty.Len() > 0 {
				t.Errorf("Expected empty table got one with %d rows", empty.Len())
			}

			if !reflect.DeepEqual(table.columns, empty.columns) {
				t.Errorf("Empty created table with different columns.\nExpect:\t%#v\nGot:\t%#v",
					table.columns, empty.columns,
				)
			}

			t.Run("out of range", func(t *testing.T) {
				t.Run("negative", func(t *testing.T) {
					defer func(t *testing.T) {
						if r := recover(); r == nil {
							t.Error("Expected table.Row to fail on negative index")
						}
					}(t)
					table.Row(-1)
				})
				t.Run("positive", func(t *testing.T) {
					defer func(t *testing.T) {
						if r := recover(); r == nil {
							t.Error("Expected table.Row to fail on too big index")
						}
					}(t)

					table.Row(38924)
				})
			})
		})
	}
}

func tableGetRowCmp(t *testing.T, table *Table, i int, exp Row) {
	defer func(t *testing.T) {
		if r := recover(); r != nil {
			t.Fatalf("Failed to get %dth row: %s", i, r)
		}
	}(t)

	var row = table.Row(i)

	if !reflect.DeepEqual(row, exp) {
		t.Fatalf("Bad %dth row.\nExpected:\t%s\nGot:\t\t%s",
			i, exp, row)
	}
}

func TestTableCalculations(t *testing.T) {
	var table = newTable(t)

	for i := 0; i < table.ColumnLen(); i++ {
		t.Run(table.columns[i].Name, func(t *testing.T) {
			defer func(t *testing.T) {
				if r := recover(); r != nil {
					t.Fatalf("table.Column(%d) failed with %q", i, r)
				}
			}(t)

			var column = table.Column(i)

			total, err := table.Total(i)
			if err != nil && CanSum(column) {
				t.Fatalf("Unexpected error on summable column %q: %q",
					column.Name, err)
			}

			avg, err := table.Average(i)
			if err != nil && CanAvg(column) {
				t.Fatalf("Unexpected error on averable column %q: %q",
					column.Name, err)
			}

			// specific manually calculated cases
			if i == 2 {
				exp, _ := decimal.NewFromString("75.5879")
				if !total.Equal(exp) {
					t.Errorf("Expected total to be %s, got %s", exp, total)
				}
			}

			if i == 4 {
				exp, _ := decimal.NewFromString("61.68")
				if !avg.Round(4).Equal(exp) {
					t.Errorf("Expected average to be %s, got %s", exp, avg)
				}
			}
		})
	}

	t.Run("out of range", func(t *testing.T) {
		t.Run("negative", func(t *testing.T) {
			var err error
			_, err = table.Total(-1)
			if err == nil {
				t.Error("Expected table.Total to fail on negative index")
			}
			_, err = table.Average(-2)
			if err == nil {
				t.Error("Expected table.Average to fail on negative index")
			}
		})
		t.Run("positive", func(t *testing.T) {
			var err error
			_, err = table.Total(123)
			if err == nil {
				t.Error("Expected table.Total to fail on too big index")
			}
			_, err = table.Average(128)
			if err == nil {
				t.Error("Expected table.Average to fail on too big index")
			}
		})
	})
}

func TestJoin(t *testing.T) {
	var (
		t1 = newTable(t)
		t2 = newTable(t)
		t3 = newTable(t)

		join = Join(*t1, *t2, *t3)

		// Sum of tables' column minus non-template name columns
		exColumns = t1.ColumnLen()*3 - 2
	)

	validateTable(t, join)

	if join.ColumnLen() != exColumns {
		t.Errorf("join.ColumnLen() = %d; want = %d", join.ColumnLen(), exColumns)
	}

	if join.Len() != t1.Len() {
		t.Errorf("join.Len() = %d; want = %d", join.Len(), t1.Len())
	}

	for i, r := range join.rows {
		for j, c := range r {
			var (
				table  *Table
				column int
			)

			switch {
			case j < t1.ColumnLen():
				table = t1
				column = j
			case j < t1.ColumnLen()+t2.ColumnLen()-1:
				table = t2
				column = j - t1.ColumnLen() + 1
			case j < t1.ColumnLen()+t2.ColumnLen()+t3.ColumnLen()-2:
				table = t3
				column = j - t1.ColumnLen() - t2.ColumnLen() + 2
			default:
				t.Errorf("Cell too far on row %d: %d, %v", i, j, c)
			}

			if c != table.Row(i).Cell(column) {
				t.Errorf("Column %d on row %d is %q; want %d=>%q",
					j, i, table.Row(i).Cell(column), column, c)
			}
		}
	}

	t.Run("different", func(t *testing.T) {
		var (
			table      = newTable(t)
			small, err = NewTable(baselineInputDataColumns[:4],
				[]Row{ // using months like that in order to keep tabular outline
					{Cell(time.Month(1).String()), "5", "10.1231", "45"},
					{Cell(time.Month(2).String()), "6", "11.3412", "44"},
				}...)

			join = Join(*table, small)
		)

		if err != nil {
			t.Fatalf("Small NewTable: %s", err)
		}

		if join.ColumnLen() > 16 {
			t.Errorf("Join of different tables has %d columns; want 16", join.ColumnLen())
		}

		if join.Len() > 12 {
			t.Errorf("Join of different tables has %d rows; want %d", join.Len(), table.Len())
		}
		validateTable(t, join)

	})

	t.Run("no tables", func(t *testing.T) {
		var join = Join()

		if join.ColumnLen() > 0 {
			t.Errorf("Join of no tables has %d columns; want 0", join.ColumnLen())
		}

		if join.Len() > 0 {
			t.Errorf("Join of no tables has %d rows; want 0", join.Len())
		}

		validateTable(t, join)
	})
}

func newTable(t *testing.T) *Table {
	var table, err = NewTable(baselineInputDataColumns, baselineInputDataRows...)
	if err != nil {
		t.Fatalf("NewTable has failed with %q", err)
	}

	return &table
}

func validateTable(t *testing.T, table Table) {
	if _, err := NewTable(table.columns, table.rows...); err != nil {
		t.Errorf("Invalid table: %s", err)
	}
}

func TestTableSort(t *testing.T) {
	var cases = []struct {
		column int
		first  Cell
		last   Cell
		ok     bool
	}{
		{column: 0, first: "April", last: "May", ok: true},
		{column: 1, first: "January", last: "May", ok: true},
		{column: 2, first: "January", last: "June", ok: true},
		{column: 3, first: "June", last: "January", ok: true},
		{column: 4, first: "January", last: "June", ok: true},
		{column: 5, first: "February", last: "April", ok: true},
		{column: 6, first: "March", last: "May", ok: true},
		{column: 7, first: "February", last: "January", ok: true},
		{column: 12, first: "March", last: "April", ok: true},
		{column: 13, ok: false},
		{column: -1, ok: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c.column), func(t *testing.T) {
			var (
				table = newTable(t)
				rows  = make([]Row, len(table.rows))
				n     = copy(rows, table.rows)
			)
			if n != len(table.rows) {
				t.Fatalf("Exected to copy %d rows, copied %d instead", n, len(table.rows))
			}

			if ok := table.SetSortByColumn(c.column); ok != c.ok {
				t.Fatalf("Expected SetSortByColumn to return %t, got %t",
					c.ok, ok)
			}

			if !c.ok {
				return
			}
			sort.Sort(table)

			if c.first != table.rows[0].Cell(0) ||
				c.last != table.rows[table.Len()-1].Cell(0) {
				t.Fatalf("Expected first row to be %s and last %s, got %s",
					c.first, c.last, table.rows)
			}

		})
	}

	t.Run("invalid sort by column", func(t *testing.T) {
		var (
			table = newTable(t)
			rows  = make([]Row, len(table.rows))
			n     = copy(rows, table.rows)
		)
		if n != len(table.rows) {
			t.Fatalf("Exected to copy %d rows, copied %d instead", n, len(table.rows))
		}

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Sorting on invalid column should've panicked")
			}
		}()

		table.sortby = 153
		sort.Sort(table)
	})

	t.Run("empty table", func(t *testing.T) {
		var table, err = NewTable([]Column{})
		if err != nil {
			t.Fatalf("NewTable has failed with %q", err)
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Sorting empty table panicked: %s", err)
			}
		}()
		sort.Sort(&table)
	})
}

func TestRow(t *testing.T) {
	var (
		max   = 2
		cells = []Cell{"Foo", "Bar", "Baz"}
		row   = Row(cells)
	)

	for i := -2; i < 5; i++ {
		cell := row.Cell(i)
		if (i < 0 || i > max) && cell != Cell("") {
			t.Errorf("Expected empty cell on invalid column, got %s", cell)
		}
		if i > 0 && i < max && cell != cells[i] {
			t.Errorf("Expected Row.Cell(%d) == %s, got %s", i, cells[i], cell)
		}
	}
}

func TestColumn(t *testing.T) {
	var cases = []struct {
		name string
		kind Kind
		sum  bool
		avg  bool
	}{
		{
			name: "Rows",
			kind: Name,
			sum:  false,
			avg:  false,
		},
		{
			name: "Heating days",
			kind: Count,
			sum:  true,
			avg:  false,
		},
		{
			name: "Total heat energy consumption",
			kind: Energy,
			sum:  true,
			avg:  false,
		},
		{
			name: "Domestic hot water consumption",
			kind: Volume,
			sum:  true,
			avg:  false,
		},
		{
			name: "Domestic hot water temperature",
			kind: Temperature,
			sum:  false,
			avg:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var column = Column{Name: c.name, Kind: c.kind}
			if c.avg != CanAvg(column) {
				t.Errorf("Expected CanAvg() to be %t", c.avg)
			}
			if c.sum != CanSum(column) {
				t.Errorf("Expected CanSum() to be %t", c.sum)
			}
		})
	}
}

func TestCellFormat(t *testing.T) {
	var cases = []struct {
		cell   Cell
		kind   Kind
		result string
	}{
		{"3.14", 0, "3.14"},
		{"3.14", Name, "3.14"},
		{"3.14", String, "3.14"},
		{"3.14", Decimal, "3.14"},
		{"3.14", Money, "3.14€"},
		{"3.14", Temperature, "3.14°C"},
		{"3.14", Count, "3"},
		{"3.14", Volume, "3.14m³"},
		{"3.14", Energy, "3.14MWh"},
		{"five", Name, "five"},
		{"five", String, "five"},
		{"five", Decimal, "0"},
		{"five", Energy, "0MWh"},
	}

	for _, c := range cases {
		if result := c.cell.Format(c.kind); result != c.result {
			t.Errorf("Expected Cell(%s).Format(%v) == %s, got %s",
				c.cell, c.kind, c.result, result)
		}
	}
}

func TestJSON(t *testing.T) {
	var table = newTable(t)

	t.Run("marshal and then unmarshal", func(t *testing.T) {
		b, err := json.Marshal(table)
		if err != nil {
			t.Fatal(err)
		}

		var utable Table
		if err := json.Unmarshal(b, &utable); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(table, &utable) {
			t.Errorf("Got different tables.\nPre:\t%#v\nPost:\t%#v",
				table, utable)
		}
	})

	t.Run("null", func(t *testing.T) {
		var table Table
		if err := json.Unmarshal([]byte("null"), &table); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("{}", func(t *testing.T) {
		var table Table
		if err := json.Unmarshal([]byte("{}"), &table); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("truncated", func(t *testing.T) {
		var table Table
		if err := json.Unmarshal([]byte(`{"columns": `), &table); err == nil {
			t.Fatal("Expected error on truncated input, got nil instead")
		}
	})
}

func TestTableTitle(t *testing.T) {
	table := newTable(t)
	if table.title != "" {
		t.Fatalf("Expected table title to be empty string but got %s", table.title)
	}
	if table.title != table.Title() {
		t.Fatalf("Expected table title to be %s string but got %s", table.Title(), table.title)
	}
	title := "Nanana batman"
	table.SetTitle(title)
	if table.Title() != title {
		t.Fatalf("Expected table title to be %s but got %s", title, table.Title())
	}
}
