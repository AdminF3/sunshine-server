package stores

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

type Portfolio interface {

	// GetPDCountries fetches all countries for given user.ID that
	// he is portfolio director.
	GetPDCountries(context.Context, uuid.UUID) []string

	// GetPortfolioRole searches and returns a user.ID for the given
	// country and given role. If it is not found any - return randomly picked
	// one ID from portfolio_roles table.
	GetPortfolioRole(context.Context, models.Country, models.PortfolioRole) (uuid.UUID, error)

	// Put registers given user as Portfolio actor for given country and role.
	Put(ctx context.Context, user uuid.UUID, country models.Country, role models.PortfolioRole) error

	// Remove matching record for user, country and role.
	Remove(ctx context.Context, user uuid.UUID, country models.Country, role models.PortfolioRole) error

	// GetPortfolioRolesPerCountry returs all users as slice of
	// uuid.UUID that are the given role for that country.
	GetPortfolioRolesPerCountry(ctx context.Context, country models.Country, role models.PortfolioRole) ([]uuid.UUID, error)

	// GetPortfolioUsersPerCountry returs all users that have the given role for that country.
	GetPortfolioUsersPerCountry(ctx context.Context, country *string, role models.PortfolioRole) ([]models.User, error)
}
