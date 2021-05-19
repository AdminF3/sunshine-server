package stores

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

// TokenStore takes care for managing tokens.
type TokenStore interface {
	// Create new token with id for purpose.
	Create(ctx context.Context, purpose models.TokenPurpose, id uuid.UUID) (*models.Token, error)

	// Get token with id and purpose.
	Get(ctx context.Context, purpose models.TokenPurpose, id uuid.UUID) (*models.Token, error)

	// Invalidate token with id for purpose.
	Invalidate(ctx context.Context, purpose models.TokenPurpose, id uuid.UUID) error
}

//go:generate mockgen -package=mocks -self_package=stageai.tech/sunshine/sunshine/mocks -destination=./../mocks/token_store.go -write_package_comment=false stageai.tech/sunshine/sunshine/stores TokenStore
