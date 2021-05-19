package controller

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func TestGetPortfolioUserPerCountry(t *testing.T) {
	e := services.NewTestEnv(t)
	pf := NewPortfolio(e)
	ustore := pf.u.FromKind("user")

	// create some dpos
	stores.NewTestPortfolioRole(t, ustore, models.DataProtectionOfficerRole, models.CountryVatican)
	stores.NewTestPortfolioRole(t, ustore, models.DataProtectionOfficerRole, models.CountryVatican)
	stores.NewTestPortfolioRole(t, ustore, models.DataProtectionOfficerRole, models.CountryVatican)
	stores.NewTestPortfolioRole(t, ustore, models.DataProtectionOfficerRole, models.CountryVatican)

	usr := stores.NewTestUser(t, ustore)
	c := "Vatican City"

	cases := []struct {
		name     string
		ctx      context.Context
		user     uuid.UUID
		country  *string
		expDPOs  int
		expected error
	}{
		{
			name:     "ok",
			ctx:      services.NewTestContext(t, e, usr),
			user:     usr.ID,
			country:  &c,
			expDPOs:  4,
			expected: nil,
		},
		{
			name: "no country",
			ctx:  services.NewTestContext(t, e, usr),
			user: usr.ID,
			// this is 4 from the above case + 6 from the default ones for every
			// consortium country that gets created with a new test env in
			// services.Env/134 line
			expDPOs:  10,
			country:  nil,
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dpos, err := pf.GetPortfolioUsersPerCountry(c.ctx, c.country,
				models.DataProtectionOfficerRole)
			if err != c.expected {
				t.Errorf("expected %v err; got %v", c.expected, err)
			}
			if len(dpos) != c.expDPOs {
				t.Errorf("expected to get %v dpos; got: %v", c.expDPOs, len(dpos))
			}
		})
	}

}
