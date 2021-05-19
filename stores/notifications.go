package stores

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

type Notifier interface {

	// Broadcast deduces and creates multiple needed notifications, based on
	// userAction and target type.
	Broadcast(ctx context.Context, ua models.UserAction, u models.User, target models.Document, o, n string, issuer uuid.UUID, comment *string)

	// Notify creates new notification for given action by given entity.
	Notify(ctx context.Context, n *models.Notification) error

	// List all unseen notifications that are for user with ID=recp.
	List(ctx context.Context, recp uuid.UUID, action *models.UserAction) ([]models.Notification, error)

	// Filter notifications based on several possible criteria.
	Filter(ctx context.Context, recp uuid.UUID, offset, limit int, action []models.UserAction,
		userID, targetID *uuid.UUID, seen *bool, key *string, etype *models.EntityType,
		country *models.Country) ([]models.Notification, error)

	// Count number of notifications matching given filters.
	Count(ctx context.Context, recp uuid.UUID, action []models.UserAction,
		userID, targetID *uuid.UUID, seen *bool, key *string, etype *models.EntityType,
		country *models.Country) (int, error)

	// See marks a given notification as already seen by user with ID=recp.
	See(ctx context.Context, id uuid.UUID, recp uuid.UUID) error

	// Get returns a single notification given its ID for user with ID=recp.
	Get(ctx context.Context, id uuid.UUID, recp uuid.UUID) (*models.Notification, error)

	// Get non-notification Document by its id and kind.
	GetDocument(ctx context.Context, id uuid.UUID, kind models.EntityType) (*models.Document, error)
}

//go:generate mockgen -package=mocks -self_package=stageai.tech/sunshine/sunshine/mocks -destination=./../mocks/notifications.go -write_package_comment=false stageai.tech/sunshine/sunshine/stores Notifier
