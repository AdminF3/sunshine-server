package contract

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestEndingBalance(t *testing.T) {
	cases := []struct {
		name      string
		bbal      string // beginning balance
		payment   string
		interest  string
		eurobor   string
		iteration int
		exp       string
	}{
		{
			name:      "ok",
			bbal:      "5000",
			payment:   "34.39",
			interest:  "5.5",
			iteration: 5,
			exp:       "4942.1",
		},
		{
			name:      "failing",
			bbal:      "dfd",
			payment:   "34.39",
			interest:  "5.5",
			iteration: 5,
			exp:       "0",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var res = c.bbal
			for i := 0; i < c.iteration; i++ {
				res = endingBalance(res, c.payment, c.interest, decimal.NewFromInt(0))
			}

			if res != c.exp {
				t.Fatalf("exp: %v, but got: %v", c.exp, res)
			}
		})
	}
}
