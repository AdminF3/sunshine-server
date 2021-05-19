package stores

import (
	"testing"

	"stageai.tech/sunshine/sunshine/models"
)

func TestCountryStats(t *testing.T) {
	db := models.NewTestGORM(t)
	store := NewUserStore(db, validate)

	empty := Stats{}
	for _, c := range models.Countries() {
		s, err := CountryStats(ctx, store, "")
		if err != nil {
			t.Errorf("CountryStats error: %v", err)
		}
		if len(s) != 0 {
			t.Errorf("Without data CountryStats(%v) = %v; expected %v", c, s, empty)
		}
	}

	NewTestUser(t, NewUserStore(db, validate))
	NewTestOrg(t, NewOrganizationStore(db, validate))
	NewTestAsset(t, NewAssetStore(db, validate))
	NewTestProject(t, NewProjectStore(db, validate))

	// NewTest* helpers create everything in Latvia.
	allStats, _ := CountryStats(ctx, store, "")
	latviaStats, _ := CountryStats(ctx, store, models.CountryLatvia)
	if allStats[models.CountryLatvia] != latviaStats[models.CountryLatvia] {
		t.Errorf("Got different result for the same country:\n")
		t.Errorf("CountryStats(ctx, s, nil) =\t%#v\n", allStats["Latvia"])
		t.Errorf("CountryStats(ctx, s, Latvia) =\t%#v\n", latviaStats["Latvia"])
	}
	nonZeroStats(t, models.CountryLatvia, allStats[models.CountryLatvia])
}

func TestCountryStats_invalid(t *testing.T) {
	db := models.NewTestGORM(t)
	store := NewUserStore(db, validate)

	_, err := CountryStats(ctx, store, "Neverland")
	if err == nil {
		t.Fatal("Expected error got nil")
	}
	t.Log(err.Error())
}

func nonZeroStats(t *testing.T, c models.Country, s Stats) {
	t.Helper()
	if s.Assets == 0 {
		t.Errorf("Zero assets in %v", c)
	}
	if s.Organizations == 0 {
		t.Errorf("Zero organizations in %v", c)
	}
	if s.Projects == 0 {
		t.Errorf("Zero projects in %v", c)
	}
	if s.Users == 0 {
		t.Errorf("Zero users in %v", c)
	}
}
