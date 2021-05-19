package graphql

import (
	"context"
	"errors"
	"time"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

func (r *mutationResolver) CreateForfaitingApplication(ctx context.Context,
	fa models.ForfaitingApplication) (*models.ForfaitingApplication, error) {

	return r.fa.Create(ctx, fa)
}

func (r *queryResolver) GetForfaitingApplication(ctx context.Context,
	id uuid.UUID) (*models.ForfaitingApplication, error) {

	return r.fa.Get(ctx, id)
}

func (r *queryResolver) ListForfaitingApplications(ctx context.Context,
	countries []string) ([]models.ForfaitingApplication, error) {
	cs := make([]models.Country, len(countries))
	for i, c := range countries {
		cs[i] = models.Country(c)
	}

	return r.fa.ListByCountries(ctx, cs)
}

func (r *mutationResolver) ReviewForfaitingApplication(ctx context.Context,
	id uuid.UUID, review models.FAReview) (*Message, error) {

	return messageResult(r.fa.Review(ctx, id, review))
}

func (r *mutationResolver) UpdateForfaitingApplication(ctx context.Context, id uuid.UUID, fa models.ForfaitingApplication) (*models.ForfaitingApplication, error) {
	return r.fa.Update(ctx, id, fa)
}

func (r *queryResolver) GetForfaitingApplicationByProject(ctx context.Context, project uuid.UUID) (*models.ForfaitingApplication, error) {
	fa, err := r.fa.GetByProject(ctx, project)

	// This is "feature-request" by the front-end gurus.
	if errors.Is(err, controller.ErrNotFound) {
		err = controller.ErrNotFound
	}

	return fa, err
}

func (r *mutationResolver) CreateForfaitingPayment(ctx context.Context, transferValue int, c models.Currency, pid uuid.UUID, td *time.Time) (*models.ForfaitingPayment, error) {
	return r.fa.CreateFP(ctx, transferValue, c, pid, td)
}

func (r *queryResolver) GetForfaitingPayment(ctx context.Context, fpid uuid.UUID, pid uuid.UUID) (*models.ForfaitingPayment, error) {
	return r.fa.GetFP(ctx, fpid, pid)
}

func (r *mutationResolver) UpdateForfaitingPayment(ctx context.Context, faid, pid uuid.UUID, transferValue *int, c *models.Currency, td *time.Time) (*models.ForfaitingPayment, error) {
	return r.fa.UpdateFP(ctx, faid, pid, transferValue, c, td)
}
