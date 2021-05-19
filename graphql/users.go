package graphql

import (
	"context"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

// ResendActivationEmail alaways returns nill error.
// To keep system information hidden from non-logged users
func (r *mutationResolver) ResendActivationEmail(ctx context.Context, email string) (*Message, error) {
	r.user.ResendActivationEmail(ctx, email)

	return msgOK, nil
}

func (r *queryResolver) ListAdmins(ctx context.Context,
	first, offset *int) (*PaginatedList, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, controller.ErrUnauthorized
	}
	// the -1 is needed for gorm limit
	if first == nil {
		first = new(int)
		*first = -1
	}
	if offset == nil {
		offset = new(int)
	}

	filter := stores.Filter{
		Offset: *offset,
		Limit:  *first,
	}

	adms, total, err := r.user.AdminList(ctx, filter)
	entities := make([]Entity, len(adms))
	for i, adms := range adms {
		entities[i] = adms

	}

	return &PaginatedList{
		TotalCount: total,
		Entities:   entities,
	}, err
}
