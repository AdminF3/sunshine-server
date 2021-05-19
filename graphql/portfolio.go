package graphql

import (
	"context"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/models"
)

func (r *mutationResolver) AddPortfolioRole(ctx context.Context, uid uuid.UUID, country string, role models.PortfolioRole) (*Message, error) {
	err := r.pf.AddPortfolioRole(ctx, uid, models.Country(country), role)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) RemovePortfolioRole(ctx context.Context, uid uuid.UUID, country string, role models.PortfolioRole) (*Message, error) {
	if err := r.pf.RemovePortfolioRole(ctx, uid, models.Country(country), role); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) AddCountryAdmin(ctx context.Context, uid uuid.UUID, country string) (*Message, error) {
	if err := r.pf.AddCountryAdmin(ctx, uid, models.Country(country)); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}
func (r *mutationResolver) RemoveCountryAdmin(ctx context.Context, uid uuid.UUID, country string) (*Message, error) {
	if err := r.pf.RemoveCountryAdmin(ctx, uid, models.Country(country)); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) AddAdminNetworkManager(ctx context.Context, uid uuid.UUID) (*Message, error) {
	if err := r.pf.AddAdminNetworkManager(ctx, uid); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) RemoveAdminNetworkManager(ctx context.Context, uid uuid.UUID) (*Message, error) {
	if err := r.pf.RemoveAdminNetworkManager(ctx, uid); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) ValidateOrganization(ctx context.Context, org uuid.UUID, status models.ValidationStatus, comment *string) (*Message, error) {
	if err := r.org.Validate(ctx, org, status, comment); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) ValidateAsset(ctx context.Context, asset uuid.UUID, status models.ValidationStatus, comment *string) (*Message, error) {
	if err := r.asset.Validate(ctx, asset, status, comment); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) ValidateUser(ctx context.Context, user uuid.UUID, status models.ValidationStatus, comment *string) (*Message, error) {
	if err := r.user.Validate(ctx, user, status, comment); err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *queryResolver) GetDPOs(ctx context.Context, country *string) ([]models.User, error) {
	return r.pf.GetPortfolioUsersPerCountry(ctx, country, models.DataProtectionOfficerRole)
}
