package stores

import (
	"fmt"
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func TestUserStore(t *testing.T) {
	db := models.NewTestGORM(t)
	store := NewUserStore(db, validate)

	StoreTest{
		store: store,
		entity: &models.User{
			Name:     "John Doe",
			Email:    "john@doe.com",
			Password: "foo",
			Country:  "Latvia",
		},
		invalidate: func(valid models.Entity) models.Entity {
			var (
				vu = valid.(*models.User)
				iu = *vu
			)
			iu.Email = "not_an_email.com"
			return &iu
		},
		update: func(doc *models.Document) models.Entity {
			var (
				orig = doc.Data.(*models.User)
				user = *orig
			)
			user.Name += " Junior"
			return &user
		},
		duplicate: func(e models.Entity) models.Entity {
			var (
				user = e.(*models.User)
				dupl = *user
			)
			dupl.Value.ID = uuid.Nil
			dupl.Email = "another@example.com"
			return &dupl
		},
		postCreate: func(created models.Entity) error {
			var password = created.(*models.User).Password
			if password == "foo" || len(password) != 60 {
				return fmt.Errorf(
					"Expected hashed password; got %s",
					password,
				)
			}
			return nil
		},
		searchBy: func(e models.Entity) string {
			return e.(*models.User).Email
		},
		memberUUID: func(*testing.T, models.Entity) uuid.UUID { return uuid.Nil },
	}.Run(t)
}
