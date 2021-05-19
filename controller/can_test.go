package controller

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func TestCan(t *testing.T) {
	e := services.NewTestEnv(t)

	ctx := context.Background()

	admin := stores.NewTestAdmin(t, e.UserStore)
	leaa := stores.NewTestUser(t, e.UserStore)
	tU := stores.NewTestUser(t, e.UserStore)
	user := stores.NewTestUser(t, e.UserStore)
	pdirector := stores.NewTestPortfolioRole(t, e.UserStore, models.PortfolioDirectorRole, models.CountryBulgaria)

	org := stores.NewTestOrg(t, e.OrganizationStore, user.ID, leaa.ID)
	proj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(user.ID))

	// The PD must be valid member of any org
	stores.NewTestOrg(t, e.OrganizationStore, pdirector.ID)

	var cases = []struct {
		name    string
		ctx     context.Context
		action  Action
		target  uuid.UUID
		result  bool
		country models.Country
	}{
		{
			name:   "Update user as superuser",
			ctx:    services.NewTestContext(t, e, admin),
			action: UpdateUser,
			target: leaa.ID,
			result: true,
		},
		{
			name:   "Update user as self",
			ctx:    services.NewTestContext(t, e, user),
			action: UpdateUser,
			target: user.ID,
			result: true,
		},
		{
			name:   "Unauthorized user update",
			ctx:    ctx,
			action: UpdateUser,
			target: leaa.ID,
			result: false,
		},
		{
			name:   "Update org as lear",
			ctx:    services.NewTestContext(t, e, user),
			action: UpdateOrganization,
			target: org.ID,
			result: true,
		},
		{
			name:   "Update proj as pm",
			ctx:    services.NewTestContext(t, e, user),
			action: UpdateProject,
			target: proj.ID,
			result: true,
		},
		{
			name:   "Assign pm as pm",
			ctx:    services.NewTestContext(t, e, user),
			action: AssignPM,
			target: proj.ID,
			result: true,
		},
		{
			name:    "Update org as non-lear",
			ctx:     services.NewTestContext(t, e, tU),
			action:  UpdateOrganization,
			target:  org.ID,
			result:  false,
			country: org.Data.(*models.Organization).Country,
		},
		{
			name:   "Update asset as lear",
			ctx:    services.NewTestContext(t, e, user),
			action: UpdateAsset,
			target: org.ID,
			result: true,
		},
		{
			name:    "authorized pd",
			ctx:     services.NewTestContext(t, e, pdirector),
			action:  GetAsset,
			target:  uuid.Nil,
			result:  true,
			country: models.CountryBulgaria,
		},
		{
			name:   "unauthorized pd",
			ctx:    services.NewTestContext(t, e, pdirector),
			action: UploadAsset,
			target: uuid.Nil,
			result: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if r := Can(c.ctx, c.action, c.target, c.country); r != c.result {
				t.Errorf("Can expected %t; got %t", c.result, r)
			}
		})
	}
}
