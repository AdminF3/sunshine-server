package stores

import (
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func TestForfaitingApplication(t *testing.T) {
	db := models.NewTestGORM(t)

	store := NewForfaitingApplicationStore(db, validate)
	manager := NewTestUser(t, store)
	StoreTest{
		store: store,
		entity: &models.ForfaitingApplication{
			Project:   NewTestProject(t, store).ID,
			ManagerID: manager.ID,
			Manager:   *manager.Data.(*models.User),
		},
		invalidate: func(doc models.Entity) models.Entity {
			invalid := *doc.(*models.ForfaitingApplication)
			invalid.Project = uuid.Nil
			return &invalid
		},
		update: func(doc *models.Document) models.Entity {
			newDoc := *doc.Data.(*models.ForfaitingApplication)
			return &newDoc
		},
		duplicate: func(e models.Entity) models.Entity {
			dupl := *e.(*models.ForfaitingApplication)
			dupl.ID = uuid.Nil
			dupl.Project = NewTestProject(t, store).ID
			return &dupl
		},
		searchBy:   func(e models.Entity) string { return "" },
		postCreate: func(models.Entity) error { return nil },
		memberUUID: func(t *testing.T, e models.Entity) uuid.UUID {
			proj := NewTestProject(t, store)
			tr := e.(*models.ForfaitingApplication)
			tr.ID = uuid.New()
			tr.Project = proj.ID
			if _, err := store.Create(ctx, tr); err != nil {
				t.Error(err)
			}

			return proj.ID
		},
	}.Run(t)
}
