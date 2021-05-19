package graphql

import "testing"

func TestEncodeCursor(t *testing.T) {
	cases := []struct {
		index  int
		result string
	}{
		{index: -1, result: "Y3Vyc29yOi0x"},
		{index: 0, result: "Y3Vyc29yOjA="},
		{index: 1, result: "Y3Vyc29yOjE="},
		{index: 2, result: "Y3Vyc29yOjI="},
		{index: 3, result: "Y3Vyc29yOjM="},
	}

	for _, c := range cases {
		if r := encodeCursor(c.index); r != c.result {
			t.Errorf("encodeCursor(%v) = %v (expected %v)",
				c.index, r, c.result)
		}
	}
}

func TestDecodeCursor(t *testing.T) {
	cases := []struct {
		cursor string
		result int
	}{
		{"", -1},
		{"Y3Vyc29yOA=", -1},
		{"Y3Vyc29jyOjA=", -1},
		{"Y3Vyc29yOi01MA==", -1},
		{"Y3Vyc29yOjA=", 0},
		{"Y3Vyc29yOjE=", 1},
		{"Y3Vyc29yOjEw", 10},
		{"Y3Vyc29yOjQy", 42},
	}

	for _, c := range cases {
		if r := decodeCursor(&c.cursor); r != c.result {
			t.Errorf("decodeCursor(encodeCursor(%v)) = %v (expected %v)",
				c.cursor, r, c.result)
		}
	}

	if r := decodeCursor(nil); r != -1 {
		t.Errorf("decodeCursor(nil) = %v (expected -1)", r)
	}

}

const fmtCalc = "calcBounds(first: %v, after: %v, last: %v, before: %v) = offset: %v, limit: %v (expected %v, %v)"

func TestCalcBounds(t *testing.T) {
	cases := []struct {
		first  int
		after  int
		last   int
		before int
		total  int

		offset int
		limit  int
	}{
		// zero args
		{first: 0, after: -1, last: 0, before: 0, offset: 0, limit: 0, total: 20},

		// bad input
		{first: 0, after: -1, last: 0, before: 1, offset: 0, limit: 1, total: 20},

		// skip  first
		{first: 20, after: 0, last: 0, before: 0, offset: 1, limit: 20, total: 20},
		{first: 0, after: 0, last: 0, before: 0, offset: 1, limit: 0, total: 20},

		// fetch first
		{first: 0, after: -1, last: 1, before: 1, offset: 0, limit: 1, total: 20},

		// fetching something from the middle
		{first: 10, after: 0, last: 0, before: 0, offset: 1, limit: 10, total: 20},
		{first: 0, after: 0, last: 10, before: 11, offset: 1, limit: 10, total: 20},
		{first: 10, after: 5, last: 0, before: 0, offset: 6, limit: 10, total: 20},
		{first: 0, after: -1, last: 5, before: 10, offset: 5, limit: 5, total: 20},
		{first: 0, after: 0, last: 5, before: 11, offset: 6, limit: 5, total: 20},

		// last without before
		{first: 0, after: -1, last: 5, before: 0, offset: 15, limit: 5, total: 20},

		// {first, after} ∩ {last, before} = {first, after}
		{first: 5, after: 0, last: 10, before: 10, offset: 1, limit: 5, total: 20},
		{first: 10, after: 5, last: 20, before: 20, offset: 6, limit: 10, total: 20},

		// {first, after} ∩ {last, before} = {last, before}
		{first: 20, after: 4, last: 8, before: 14, offset: 6, limit: 8, total: 20},

		// {first, after} ∩ {last, before} = ∅ => ignore {last, before}
		{first: 10, after: 20, last: 15, before: 15, offset: 21, limit: 10, total: 20},
	}

	for _, c := range cases {
		offset, limit := calcBounds(c.first, c.after, c.last, c.before, c.total)
		if c.offset != offset || c.limit != limit {
			t.Errorf(fmtCalc, c.first, c.after, c.last, c.before,
				offset, limit, c.offset, c.limit)
		}
	}
}

func TestPageInfo(t *testing.T) {
	cases := []struct {
		offset int
		limit  int
		length int
		prev   bool
		next   bool
	}{
		{0, 0, 0, false, false},
		{0, 1, 0, false, false},
		{0, 1, 1, false, false},
		{0, 100, 100, false, false},
		{0, 100, 50, false, false},
		{0, 50, 100, false, true},
		{10, 50, 100, true, true},
		{100, 100, 100, true, false},
	}

	for _, c := range cases {
		p := pageInfo(c.offset, c.limit, c.length)
		if p.HasNextPage != c.next || p.HasPreviousPage != c.prev {
			t.Errorf("pageInfo(%v, %v, %v) = {prev: %v, next: %v} (expected %v, %v)",
				c.offset, c.limit, c.length, c.prev, c.next,
				p.HasPreviousPage, p.HasNextPage)
		}
	}
}
