package graphql

import (
	"context"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func (r *queryResolver) ListAssets(ctx context.Context,
	filterMine *bool, first, offset *int, ascending *bool) (*PaginatedList, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, controller.ErrUnauthorized
	}
	var id uuid.UUID
	if filterMine != nil && *filterMine {
		id = cv.User.ID
	}

	return listAssets(ctx, first, offset, listAll(r.asset, id), ascending)
}

func (r *queryResolver) ListAssetReports(ctx context.Context,
	first, offset *int, ascending *bool) (*PaginatedList, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, controller.ErrUnauthorized
	}

	return listAssets(ctx, first, offset, listReports(r.asset), ascending)
}

func listAssets(ctx context.Context,
	first, offset *int,
	f listf, ascending *bool) (*PaginatedList, error) {

	if first == nil {
		first = new(int)
	}
	if offset == nil {
		offset = new(int)
	}

	filter := stores.Filter{
		Offset: *offset,
		Limit:  *first,
	}

	if ascending != nil {
		filter.Ascending = *ascending
	}
	assets, total, err := f(ctx, filter)

	entities := make([]Entity, len(assets))
	for i, asset := range assets {
		entities[i] = asset.Data.(*models.Asset)
	}

	return &PaginatedList{
		TotalCount: total,
		Entities:   entities,
	}, err
}

// listf generic type (wrapper) for listing assets.3
type listf func(cxt context.Context, f stores.Filter) (
	assets []models.Document, total int, err error)

// listAll decorates the wrapper with generic list assets.
func listAll(asset *controller.Asset, id uuid.UUID) listf {
	return func(ctx context.Context, f stores.Filter) (
		assets []models.Document, total int, err error) {
		docs, _, total, err := asset.List(ctx, id, f)
		return docs, total, err
	}
}

// listReports decorates the wrapper with list asset reports.
func listReports(asset *controller.Asset) listf {
	return func(ctx context.Context, f stores.Filter) (
		assets []models.Document, total int, err error) {
		return asset.Reports(ctx, f)
	}
}

func (r *mutationResolver) ClaimAssetResidency(ctx context.Context, assetID uuid.UUID) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}
	err := r.asset.ClaimAssetResidency(ctx, assetID)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}
