package stores

import (
	"testing"
	"time"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func TestOrganizationStore(t *testing.T) {
	pdb := models.NewTestGORM(t)
	store := NewOrganizationStore(pdb, validate)

	StoreTest{
		store: store,
		entity: &models.Organization{
			Name:               "Goo Corporation",
			VAT:                "6388047V",
			Address:            "End of the world",
			Telephone:          "+359888123456",
			Website:            "https://goocorp.example",
			LegalForm:          models.LegalFormPublicOrganization,
			Registered:         time.Date(2017, time.October, 2, 12, 30, 10, 5, time.UTC),
			Roles:              models.OrgRoles{LEAR: NewTestUser(t, store).ID},
			Email:              "notfakeorg@real.com",
			Country:            "Latvia",
			RegistrationNumber: "some_random_reg_num",
		},
		invalidate: func(valid models.Entity) models.Entity {
			var (
				vorg = valid.(*models.Organization)
				iorg = *vorg
			)
			iorg.Name = ""
			return &iorg
		},
		update: func(doc *models.Document) models.Entity {
			var (
				orig = doc.Data.(*models.Organization)
				org  = *orig
			)
			org.Name = "Evil " + org.Name
			org.VAT = "GB" + uuid.New().String()
			return &org
		},
		duplicate: func(e models.Entity) models.Entity {
			var (
				org  = e.(*models.Organization)
				dupl = *org
			)
			dupl.ID = uuid.Nil
			dupl.VAT = "GB117223643"
			dupl.Name = "Good Company"
			return &dupl
		},
		searchBy: func(e models.Entity) string {
			return e.(*models.Organization).Name
		},
		postCreate: func(models.Entity) error { return nil },
		memberUUID: func(t *testing.T, e models.Entity) uuid.UUID {
			var (
				user = NewTestUser(t, store)
				o    = e.(*models.Organization)
			)
			o.ID = uuid.New()
			o.Roles.LEAR = user.ID
			o.VAT = uuid.New().String()
			if _, err := store.Create(ctx, o); err != nil {
				t.Error(err)
			}
			return user.ID
		},
		beforeSave: func(e models.Entity) { e.(*models.Organization).ConvertRoles() },
	}.Run(t)
}
