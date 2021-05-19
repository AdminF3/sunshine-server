package stores

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/models"
)

var _ TokenStore = new(tokenStore)

func TestTokenStore(t *testing.T) {
	db := models.NewTestGORM(t)
	id := NewTestUser(t, NewUserStore(db, validate)).ID
	store := NewTokenStore(db, validate)

	ctx := context.Background()
	purposes := []models.TokenPurpose{
		models.SessionToken,
		models.CreateToken,
		models.ResetPwdToken,
	}

	for _, purpose := range purposes {
		t.Run(purpose.String(), func(t *testing.T) {

			t.Run("get", func(t *testing.T) {
				token := createToken(t, store, purpose, id)

				token, err := store.Get(ctx, purpose, token.ID)
				if err != nil {
					t.Fatalf("Get error: %s", err)
				}
				if !models.ValidToken(*token, purpose, validate) {
					t.Fatal("Got invalid token")
				}
			})

			t.Run("invalidate", func(t *testing.T) {
				token := createToken(t, store, purpose, id)

				err := store.Invalidate(ctx, purpose, token.ID)
				if err != nil {
					t.Fatalf("Invalidate error: %s", err)
				}
				_, err = store.Get(ctx, purpose, id)
				if err == nil {
					t.Error("Got token after invalidation")
				}
				err = store.Invalidate(ctx, purpose, token.ID)
				if err != nil {
					t.Fatalf("Invalidating invalid token: %v", err)
				}
			})
		})
	}
}

func createToken(t *testing.T, store TokenStore, purpose models.TokenPurpose, id uuid.UUID) *models.Token {
	token, err := store.Create(ctx, purpose, id)
	if err != nil {
		t.Fatalf("Create error: %s", err)
	}

	if !models.ValidToken(*token, purpose, validate) {
		t.Fatal("Created invalid token")
	}

	return token
}
