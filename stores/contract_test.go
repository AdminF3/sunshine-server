package stores

import (
	"testing"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func TestContractStore(t *testing.T) {
	db := models.NewTestGORM(t)
	store := NewContractStore(db, validate)

	StoreTest{
		store:  store,
		entity: contract.New(NewTestProject(t, store).ID),
		invalidate: func(valid models.Entity) models.Entity {
			var (
				vc = valid.(*contract.Contract)
				ic = *vc
			)
			ic.Project = uuid.Nil
			return &ic
		},
		update: func(doc *models.Document) models.Entity {
			var (
				orig  = doc.Data.(*contract.Contract)
				contr = *orig
			)
			contr.Project = NewTestProject(t, store).ID
			return &contr
		},
		duplicate: func(e models.Entity) models.Entity {
			var (
				contr = e.(*contract.Contract)
				dupl  = *contr
			)
			dupl.Value.ID = uuid.Nil
			dupl.Project = NewTestProject(t, store).ID
			return &dupl
		},
		searchBy:   func(e models.Entity) string { return "" },
		postCreate: func(models.Entity) error { return nil },
		memberUUID: func(*testing.T, models.Entity) uuid.UUID { return uuid.Nil },
	}.Run(t)
}
