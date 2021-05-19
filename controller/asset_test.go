package controller

import (
	"context"
	"io"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

const fmtValidateAsset = `{"valid": 3}`

func TestUpdateValidateAsset(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	a := NewAsset(e)
	ustore := a.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	lear := stores.NewTestUser(t, ustore)
	u := stores.NewTestUser(t, ustore)

	o := stores.NewTestOrg(t, a.store.FromKind("organization"), lear.ID)
	a1 := stores.NewTestAsset(t, a.store.FromKind("asset"), stores.TAWithOrg(o.ID))
	a2 := stores.NewTestAsset(t, a.store.FromKind("asset"))

	cases := []struct {
		name     string
		ctx      context.Context
		asset    uuid.UUID
		expected error
		valid    models.ValidationStatus
		body     io.Reader
	}{
		{
			name:     "lear validates owned asset",
			ctx:      services.NewTestContext(t, e, lear),
			asset:    a1.ID,
			expected: nil,
			valid:    models.ValidationStatusPending,
			body:     strings.NewReader(fmtValidateAsset),
		},
		{
			name:     "admin validates",
			ctx:      services.NewTestContext(t, e, su),
			asset:    a2.ID,
			expected: nil,
			valid:    models.ValidationStatusDeclined,
			body:     strings.NewReader(fmtValidateAsset),
		},
		{
			name:     "lear validates other asset",
			ctx:      services.NewTestContext(t, e, lear),
			asset:    a2.ID,
			expected: ErrUnauthorized,
			body:     strings.NewReader(fmtValidateAsset),
		},
		{
			name:     "random user validates",
			ctx:      services.NewTestContext(t, e, u),
			asset:    a2.ID,
			expected: ErrUnauthorized,
			body:     strings.NewReader(fmtValidateAsset),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			updated, _, err := a.Update(c.ctx, c.asset, c.body)
			if err != c.expected {
				t.Errorf("expected: %v, got: %v", c.expected, err)
			}
			if err == nil && updated.Data.(*models.Asset).Valid != c.valid {
				t.Errorf("expexted status: %v, got status: %v", c.valid, updated.Data.(*models.Asset).Valid)
			}
		})
	}
}

func TestValidateAsset(t *testing.T) {
	e := services.NewTestEnv(t)

	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	any := gomock.Any()
	not.EXPECT().Broadcast(any, any, any, any, any, any, any, any).AnyTimes()

	a := NewAsset(e)
	ustore := a.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	ca := stores.NewTestPortfolioRole(t, ustore, models.CountryAdminRole, models.CountryLatvia)
	lear := stores.NewTestUser(t, ustore)

	org := stores.NewTestOrg(t, a.store.FromKind("organization"), lear.ID)
	ast := stores.NewTestAsset(t, a.store.FromKind("asset"), stores.TAWithOrg(org.ID))

	cmnt := "da"
	cases := []struct {
		name      string
		ctx       context.Context
		asset     uuid.UUID
		status    models.ValidationStatus
		oldStatus models.ValidationStatus
		expected  error
		comment   *string
	}{
		{
			name:     "ok ca",
			ctx:      services.NewTestContext(t, e, ca),
			asset:    ast.ID,
			status:   models.ValidationStatusRegistered,
			expected: nil,
		},
		{
			name:     "ok su",
			ctx:      services.NewTestContext(t, e, su),
			asset:    ast.ID,
			status:   models.ValidationStatusDeclined,
			expected: nil,
		},
		{
			name:      "lear unauth",
			ctx:       services.NewTestContext(t, e, lear),
			asset:     ast.ID,
			status:    models.ValidationStatusValid,
			oldStatus: models.ValidationStatusDeclined,
			expected:  ErrUnauthorized,
		},
		{
			name:     "with comment",
			ctx:      services.NewTestContext(t, e, ca),
			asset:    ast.ID,
			status:   models.ValidationStatusRegistered,
			expected: nil,
			comment:  &cmnt,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := a.Validate(c.ctx, c.asset, c.status, c.comment)
			if err != c.expected {
				t.Errorf("expected %v err; got %v", c.expected, err)
			}

			upd, _, err := a.Get(c.ctx, c.asset)
			if err != nil {
				t.Errorf("could not fetch updated asset; got %v", err)
			}
			if c.expected != nil {
				c.status = c.oldStatus
			}
			if upd.Data.(*models.Asset).Valid != c.status {
				t.Errorf("expected status %v; got %v", c.status, upd.Data.(*models.Asset).Valid)
			}
		})
	}
}

func TestListReports(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewAsset(e)

	// anm
	anm := stores.NewTestAdminNwManager(t, e.UserStore)

	// random user
	u := stores.NewTestUser(t, e.UserStore)

	// user countries of CA role
	ca := stores.NewTestPortfolioRole(t, e.UserStore, models.CountryAdminRole, models.CountryBulgaria)

	lear := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)

	tl := 10
	makeTestAssets(t, e.AssetStore, tl, org.ID)

	cases := []struct {
		name       string
		ctx        context.Context
		err        error
		total, exp int
	}{
		{
			name: "unauth",
			ctx:  services.NewTestContext(t, e, u),
			err:  ErrUnauthorized,
		},
		{
			name:  "admin network manager",
			ctx:   services.NewTestContext(t, e, anm),
			total: tl,
			exp:   tl,
		},
		{
			name:  "lear",
			ctx:   services.NewTestContext(t, e, lear),
			total: 4,
			exp:   4,
		},
		{
			name:  "ca",
			ctx:   services.NewTestContext(t, e, ca),
			total: 4,
			exp:   4,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			docs, total, err := contr.Reports(c.ctx, stores.Filter{})
			if err != c.err {
				t.Errorf("exp err: %v, but got: %v", c.err, err)
			}

			if total != c.total {
				t.Fatalf("expected count: %d, but got: %d", c.total, total)
			}

			if len(docs) != c.exp {
				t.Fatalf("expected docs: %d, but got: %d", c.exp, len(docs))
			}
		})
	}
}

func TestListUserAssets(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewAsset(e)

	// random user
	u := stores.NewTestUser(t, e.UserStore)

	lear := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)

	tl := 4
	makeTestAssets(t, e.AssetStore, tl, org.ID)

	cases := []struct {
		name       string
		ctx        context.Context
		uid        uuid.UUID
		filter     stores.Filter
		err        error
		total, exp int
	}{
		{
			name:   "normaluser",
			ctx:    services.NewTestContext(t, e, u),
			uid:    u.ID,
			filter: stores.Filter{},
			total:  0,
			exp:    0,
		},
		{
			name:   "lear",
			ctx:    services.NewTestContext(t, e, lear),
			uid:    lear.ID,
			filter: stores.Filter{NullFields: []string{"esco"}},
			total:  2,
			exp:    2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			docs, _, total, err := contr.List(c.ctx, c.uid, c.filter)
			if err != c.err {
				t.Errorf("exp err: %v, but got: %v", c.err, err)
			}

			if total != c.total {
				t.Fatalf("expected count: %d, but got: %d", c.total, total)
			}

			if len(docs) != c.exp {
				t.Fatalf("expected docs: %d, but got: %d", c.exp, len(docs))
			}
		})
	}

}

func makeTestAssets(t *testing.T, st stores.Store, count int, orgid uuid.UUID) []models.Asset {
	assets := make([]models.Asset, count)
	for i := range assets {
		var asset *models.Document

		if i%3 == 0 {
			asset = stores.NewTestAsset(t, st,
				stores.TAWithOrg(orgid),
				stores.TAWithCountry(models.CountryBulgaria),
				stores.TAWithEsco(nil),
			)
		} else {
			asset = stores.NewTestAsset(t, st)
		}
		assets[i] = *asset.Data.(*models.Asset)
	}

	return assets
}
