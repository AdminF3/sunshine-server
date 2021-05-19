package stores

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/models"
)

func TestMeetingsStore(t *testing.T) {
	db := models.NewTestGORM(t)

	store := NewMeetingsStore(db, validate)

	StoreTest{
		store: store,
		entity: &models.Meeting{
			Name:        "Init meeting",
			Host:        NewTestOrg(t, store).ID,
			Location:    "Tintyava 15-17",
			Date:        time.Now(),
			Objective:   "Review new assets",
			Stakeholder: models.LegalFormNGO,
			Stage:       "First date",
			Notes:       "Can be skipped",
		},
		invalidate: func(valid models.Entity) models.Entity {
			vm := valid.(*models.Meeting)
			ivm := *vm

			ivm.Name = ""
			ivm.Objective = ""

			return &ivm
		},
		update: func(doc *models.Document) models.Entity {
			orig := doc.Data.(*models.Meeting)
			m := *orig

			m.Name = "BJJ strategy"

			return &m
		},
		duplicate: func(e models.Entity) models.Entity {
			m := e.(*models.Meeting)
			dupl := *m

			dupl.ID = uuid.Nil
			dupl.Name = "Team building review"

			return &dupl
		},
		searchBy:   func(e models.Entity) string { return "" },
		postCreate: func(models.Entity) error { return nil },
		memberUUID: func(*testing.T, models.Entity) uuid.UUID { return uuid.Nil },
	}.Run(t)
}
