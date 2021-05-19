package stores

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	validator "gopkg.in/go-playground/validator.v9"
)

func TestPDStore(t *testing.T) {
	db := models.NewTestGORM(t)

	st := NewUserStore(db, validator.New())
	pf := NewPortfolioStore(db)

	u1 := NewTestPortfolioRole(t, st, models.PortfolioDirectorRole, "Bulgaria")
	u2 := NewTestPortfolioRole(t, st, models.PortfolioDirectorRole, "Bulgaria")

	pf.Put(ctx, u1.ID, "Bulgaria", models.PortfolioDirectorRole)
	pf.Put(ctx, u1.ID, "Greece", models.PortfolioDirectorRole)
	pf.Put(ctx, u2.ID, "Bulgaria", models.PortfolioDirectorRole)

	c := pf.GetPDCountries(context.Background(), u1.ID)

	if len(c) != 2 {
		t.Fatalf("expected %d countries, found %d", 2, len(c))
	}
	if c[0] != "Bulgaria" {
		t.Fatalf("expected to find %s, got %s", c[0], "Bulgaria")
	}

	_, err := pf.GetPortfolioRole(ctx, "Latvia", models.PortfolioDirectorRole)
	if err != nil {
		t.Fatal("expected to legit uuid, found null")
	}

	_, err = pf.GetPortfolioRole(ctx, "Greece", models.PortfolioDirectorRole)
	if err != nil {
		t.Fatal("expected to legit uuid, found null")
	}

	c = pf.GetPDCountries(context.Background(), u1.ID)
	oldLen := len(c)
	err = pf.Remove(context.Background(), u1.ID, "Bulgaria", models.PortfolioDirectorRole)
	if err != nil {
		t.Fatal("Store returns an error from remove:", err)
	}
	c = pf.GetPDCountries(context.Background(), u1.ID)
	newLen := len(c)
	if newLen == oldLen {
		t.Fatalf("Expected one record to be removed, got no records removed expected: %v got: %v", oldLen-1, newLen)
	}
}
