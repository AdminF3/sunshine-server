package controller

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestCountry(t *testing.T) {
	t.Run("setVat", setVat)
	t.Run("getCountry", getCountry)
}

func setVat(t *testing.T) {
	e := services.NewTestEnv(t)
	ct := NewCountry(e)

	admin := stores.NewTestAdmin(t, e.UserStore)

	cases := []struct {
		name string
		err  error
		nv   int

		ctx context.Context
	}{
		{
			name: "ok admin",
			ctx:  services.NewTestContext(t, e, admin),
			nv:   42,
		},
		{
			name: "unauth",
			ctx:  context.Background(),
			nv:   42,
			err:  ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			upd, err := ct.SetVat(c.ctx, models.CountryBulgaria, c.nv)
			if err != c.err {
				t.Errorf("Could not set vat; err: %v", err)
			}

			if err == nil && upd.VAT != c.nv {
				t.Errorf("VAT was not updated correctly; got: %v, want: %v", upd.VAT, c.nv)
			}
		})
	}
}

func getCountry(t *testing.T) {
	e := services.NewTestEnv(t)
	ct := NewCountry(e)

	admin := stores.NewTestAdmin(t, e.UserStore)
	random := stores.NewTestUser(t, e.UserStore)

	cases := []struct {
		name string
		err  error
		ctx  context.Context
	}{
		{
			name: "ok admin",
			ctx:  services.NewTestContext(t, e, admin),
		},
		{
			name: "fail random guy",
			ctx:  services.NewTestContext(t, e, random),
			err:  ErrUnauthorized,
		},
		{
			name: "unauth",
			ctx:  context.Background(),
			err:  ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bg, err := ct.GetCountry(c.ctx, models.CountryBulgaria)
			if err != c.err {
				t.Errorf("Got unexpected err; exp: %v, got: %v", c.err, err)
			}
			if err == nil && (bg.VAT != 20 || bg.Country != models.CountryBulgaria) {
				t.Errorf("Got unexpected country; %v", bg)
			}
		})
	}
}
