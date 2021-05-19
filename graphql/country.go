package graphql

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"
)

func (r *mutationResolver) SetVat(ctx context.Context, country string, vat int) (*models.CountryVat, error) {
	return r.ctry.SetVat(ctx, models.Country(country), vat)
}

func (r *queryResolver) GetCountry(ctx context.Context, country string) (*models.CountryVat, error) {
	return r.ctry.GetCountry(ctx, models.Country(country))
}
