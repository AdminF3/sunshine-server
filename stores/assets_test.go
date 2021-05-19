package stores

import (
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func TestAssetStore(t *testing.T) {
	db := models.NewTestGORM(t)
	store := NewAssetStore(db, validate)

	StoreTest{
		store: store,
		entity: &models.Asset{
			Owner:        NewTestOrg(t, store).ID,
			Address:      "221B Baker str., London, UK",
			Coordinates:  models.Coords{Lat: 9.83012, Lng: 37.75721},
			Area:         9000,
			Cadastre:     "666666666666",
			Country:      "Latvia",
			BuildingType: models.BuildingType104,
		},
		invalidate: func(valid models.Entity) models.Entity {
			var (
				va = valid.(*models.Asset)
				ia = *va
			)
			ia.Area = 0
			ia.Owner = NewTestOrg(t, store).ID
			return &ia
		},
		update: func(doc *models.Document) models.Entity {
			var (
				orig  = doc.Data.(*models.Asset)
				asset = *orig
			)
			asset.Area = 9001
			return &asset
		},
		duplicate: func(e models.Entity) models.Entity {
			var (
				asset = e.(*models.Asset)
				dupl  = *asset
			)
			dupl.ID = uuid.Nil
			dupl.Coordinates = models.Coords{Lat: 10.83012, Lng: 73.75721}
			dupl.Address = "33a Dragan Tzankov bul., Sofia, BG"
			dupl.Cadastre = "666666666667"
			return &dupl
		},
		searchBy: func(e models.Entity) string {
			return e.(*models.Asset).Address
		},
		postCreate: func(models.Entity) error { return nil },
		memberUUID: func(t *testing.T, e models.Entity) uuid.UUID {
			var (
				org   = NewTestOrg(t, store)
				asset = e.(*models.Asset)
			)
			asset.ID = uuid.New()
			asset.Coordinates.Lat += 3.14
			asset.Owner = org.ID
			asset.Cadastre = "666666666668"
			if _, err := store.Create(ctx, asset); err != nil {
				t.Error(err)
			}
			return org.ID
		},
	}.Run(t)

}
