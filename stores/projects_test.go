package stores

import (
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func TestProjectStore(t *testing.T) {
	db := models.NewTestGORM(t)
	store := NewProjectStore(db, validate)

	StoreTest{
		store: store,
		entity: &models.Project{
			Name:              "Project",
			Owner:             NewTestOrg(t, store).ID,
			Asset:             NewTestAsset(t, store).ID,
			Status:            models.ProjectStatusPlanning,
			AirTemperature:    20,
			WaterTemperature:  40,
			GuaranteedSavings: 51.16,
			Roles: models.ProjRoles{
				PM:   []uuid.UUID{NewTestUser(t, store).ID},
				PaCo: []uuid.UUID{NewTestUser(t, store).ID, NewTestUser(t, store).ID},
			},
			PortfolioDirector: NewTestAdmin(t, store).ID,
			FundManager:       &NewTestUser(t, store).ID,
			Country:           "Latvia",
			Milestone:         models.MilestoneAcquisitionMeeting,
			ConsortiumOrgs:    []string{NewTestOrg(t, store).ID.String()},
		},
		invalidate: func(valid models.Entity) models.Entity {
			var (
				vproj = valid.(*models.Project)
				iproj = *vproj
			)
			iproj.Name = ""
			return &iproj
		},
		update: func(doc *models.Document) models.Entity {
			var (
				orig = doc.Data.(*models.Project)
				proj = *orig
			)
			proj.Status = models.ProjectStatusAbandoned
			proj.Roles = models.ProjRoles{
				PM: []uuid.UUID{NewTestUser(t, store).ID},
			}
			return &proj
		},
		duplicate: func(e models.Entity) models.Entity {
			var (
				proj = e.(*models.Project)
				dupl = *proj
			)
			dupl.ID = uuid.Nil
			dupl.Name = "DuplProj"
			dupl.ProjectRoles = nil
			return &dupl
		},
		searchBy: func(e models.Entity) string {
			return e.(*models.Project).Name
		},
		postCreate: func(models.Entity) error { return nil },
		memberUUID: func(t *testing.T, e models.Entity) uuid.UUID {
			var (
				user = NewTestUser(t, store)
				proj = e.(*models.Project)
			)

			NewTestProject(t, store, TPrjWithOrg(proj.Owner), TPrjWithPm(user.ID))
			NewTestProject(t, store, TPrjWithOrg(proj.Owner), TPrjWithPm(user.ID))
			NewTestProject(t, store)

			return user.ID
		},
		beforeSave: func(e models.Entity) { e.(*models.Project).ConvertRoles() },
	}.Run(t)
}
