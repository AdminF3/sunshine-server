package graphql

import (
	"context"
	"time"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

// ctxt is a dummy type for storing context values.
//
// For more information see the documentation of `context.WithValue`.
type ctxt string

const ctxkey ctxt = "stageai.tech/sunshine/sunshine/graphql"

// dataloader stores dataloaders intended to be stored in a request-scoped
// context and be accessible by resolvers.
type dataloader struct {
	Project                 *ProjectLoader
	AssetResidentCount      *AssetResidentCountLoader
	AssetOwnerName          *AssetOwnerLoader
	FAReviews               *FAReviewsLoader
	User                    *UserLoader
	FABankAttachments       *FAAttachmentsLoader
	FAStatementsAttachments *FAAttachmentsLoader
	WPReviews               *WPReviewsLoader
	MPReviews               *MPReviewsLoader
	Organization            *OrganizationLoader
}

//go:generate dataloaden ProjectLoader github.com/google/uuid.UUID stageai.tech/sunshine/sunshine/models.Project
//go:generate dataloaden OrganizationLoader github.com/google/uuid.UUID stageai.tech/sunshine/sunshine/models.Organization
//go:generate dataloaden UserLoader github.com/google/uuid.UUID stageai.tech/sunshine/sunshine/models.User
//go:generate dataloaden AssetResidentCountLoader github.com/google/uuid.UUID int
//go:generate dataloaden AssetOwnerLoader github.com/google/uuid.UUID string
//go:generate dataloaden FAReviewsLoader github.com/google/uuid.UUID []stageai.tech/sunshine/sunshine/models.FAReview
//go:generate dataloaden WPReviewsLoader github.com/google/uuid.UUID []stageai.tech/sunshine/sunshine/models.WPReview
//go:generate dataloaden MPReviewsLoader github.com/google/uuid.UUID []stageai.tech/sunshine/sunshine/models.MPReview
//go:generate dataloaden FAAttachmentsLoader github.com/google/uuid.UUID []stageai.tech/sunshine/sunshine/models.Attachment
//go:generate dataloaden UserLoader github.com/google/uuid.UUID stageai.tech/sunshine/sunshine/models.User

func dataloaderCtx(ctx context.Context, r *Resolver) context.Context {
	return context.WithValue(ctx, ctxkey, dataloader{
		Project: NewProjectLoader(
			ProjectLoaderConfig{
				MaxBatch: 20,
				Wait:     10 * time.Millisecond,
				Fetch: func(ids []uuid.UUID) ([]models.Project, []error) {
					return fetchProjects(ctx, r.project, ids)
				},
			}),
		Organization: NewOrganizationLoader(
			OrganizationLoaderConfig{
				MaxBatch: 20,
				Wait:     10 * time.Millisecond,
				Fetch: func(ids []uuid.UUID) ([]models.Organization, []error) {
					return fetchOrgs(ctx, r.org, ids)
				},
			}),
		AssetResidentCount: NewAssetResidentCountLoader(
			AssetResidentCountLoaderConfig{
				MaxBatch: 500,
				Wait:     100 * time.Millisecond,
				Fetch: func(ids []uuid.UUID) ([]int, []error) {
					return fetchAssetResidentCount(ctx, r.asset, ids)
				},
			}),
		AssetOwnerName: NewAssetOwnerLoader(
			AssetOwnerLoaderConfig{
				MaxBatch: 500,
				Wait:     100 * time.Millisecond,
				Fetch: func(keys []uuid.UUID) ([]string, []error) {
					return fetchAssetsName(ctx, r.asset, keys)
				},
			}),
		FAReviews: NewFAReviewsLoader(FAReviewsLoaderConfig{
			MaxBatch: 20,
			Wait:     10 * time.Millisecond,
			Fetch: func(ids []uuid.UUID) ([][]models.FAReview, []error) {
				return fetchReviews(ctx, r.fa, ids)
			},
		}),
		WPReviews: NewWPReviewsLoader(WPReviewsLoaderConfig{
			MaxBatch: 20,
			Wait:     10 * time.Millisecond,
			Fetch: func(ids []uuid.UUID) ([][]models.WPReview, []error) {
				return fetchWPReviews(ctx, r.wp, ids)
			},
		}),
		MPReviews: NewMPReviewsLoader(MPReviewsLoaderConfig{
			MaxBatch: 20,
			Wait:     10 * time.Millisecond,
			Fetch: func(ids []uuid.UUID) ([][]models.MPReview, []error) {
				return fetchMPReviews(ctx, r.mp, ids)
			},
		}),
		User: NewUserLoader(
			UserLoaderConfig{
				MaxBatch: 20,
				Wait:     10 * time.Millisecond,
				Fetch: func(ids []uuid.UUID) ([]models.User, []error) {
					return fetchUsers(ctx, r.user, ids)
				},
			},
		),
		FABankAttachments: NewFAAttachmentsLoader(FAAttachmentsLoaderConfig{
			MaxBatch: 20,
			Wait:     10 * time.Millisecond,
			Fetch: func(ids []uuid.UUID) ([][]models.Attachment, []error) {
				return fetchFAAttachments(ctx, r.fa, "fa bank confirmation", ids)
			},
		}),
		FAStatementsAttachments: NewFAAttachmentsLoader(FAAttachmentsLoaderConfig{
			MaxBatch: 20,
			Wait:     10 * time.Millisecond,
			Fetch: func(ids []uuid.UUID) ([][]models.Attachment, []error) {
				return fetchFAAttachments(ctx, r.fa, "fa financial statements", ids)
			},
		}),
	})
}

func fetchAssetResidentCount(ctx context.Context, c *controller.Asset, ids []uuid.UUID) ([]int, []error) {
	m, err := c.ResidentsCount(ctx, ids...)

	if err != nil {
		// multiply the error for each id
		errs := make([]error, len(ids))
		for i := range ids {
			errs[i] = err
		}
		return nil, errs
	}

	result := make([]int, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func fetchAssetsName(ctx context.Context, c *controller.Asset, ids []uuid.UUID) ([]string, []error) {
	m, err := c.OwnerNames(ctx, ids...)

	if err != nil {
		// multiply the error for each id
		errs := make([]error, len(ids))
		for i := range ids {
			errs[i] = err
		}
		return nil, errs
	}

	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func multiplyErr(err error, n int) []error {
	errs := make([]error, n)
	for i := range errs {
		errs[i] = err
	}
	return errs
}

func fetchProjects(ctx context.Context,
	c *controller.Project, ids []uuid.UUID) ([]models.Project, []error) {

	m, err := c.ListByIDs(ctx, ids...)
	if err != nil {
		return nil, multiplyErr(err, len(ids))
	}

	result := make([]models.Project, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func fetchOrgs(ctx context.Context,
	c *controller.Organization, ids []uuid.UUID) ([]models.Organization, []error) {

	m, err := c.ListByIDs(ctx, ids...)
	if err != nil {
		return nil, multiplyErr(err, len(ids))
	}

	result := make([]models.Organization, len(ids))
	for i, id := range ids {
		result[i] = m[id]
	}
	return result, nil
}

func fetchReviews(ctx context.Context,
	c *controller.ForfaitingAgreement, ids []uuid.UUID) ([][]models.FAReview, []error) {

	rs, err := c.FetchReviews(ctx, ids...)
	if err != nil {
		return nil, multiplyErr(err, len(ids))
	}

	rmap := make(map[uuid.UUID][]models.FAReview, len(rs))
	for _, r := range rs {
		rmap[r.ForfaitingApplicationID] = append(rmap[r.ForfaitingApplicationID], r)
	}

	result := make([][]models.FAReview, len(ids))
	for i, id := range ids {
		result[i] = rmap[id]
	}
	return result, nil
}

func fetchWPReviews(ctx context.Context,
	c *controller.WorkPhase, ids []uuid.UUID) ([][]models.WPReview, []error) {

	rs, err := c.FetchReviews(ctx, ids...)
	if err != nil {
		return nil, multiplyErr(err, len(ids))
	}

	rmap := make(map[uuid.UUID][]models.WPReview, len(rs))
	for _, r := range rs {
		rmap[r.WPID] = append(rmap[r.WPID], r)
	}

	result := make([][]models.WPReview, len(ids))
	for i, id := range ids {
		result[i] = rmap[id]
	}
	return result, nil
}

func fetchMPReviews(ctx context.Context,
	c *controller.MonitoringPhase, ids []uuid.UUID) ([][]models.MPReview, []error) {

	rs, err := c.FetchReviews(ctx, ids...)
	if err != nil {
		return nil, multiplyErr(err, len(ids))
	}

	rmap := make(map[uuid.UUID][]models.MPReview, len(rs))
	for _, r := range rs {
		rmap[r.MPID] = append(rmap[r.MPID], r)
	}

	result := make([][]models.MPReview, len(ids))
	for i, id := range ids {
		result[i] = rmap[id]
	}
	return result, nil
}

func fetchFAAttachments(
	ctx context.Context,
	c *controller.ForfaitingAgreement,
	attType string,
	ids []uuid.UUID) ([][]models.Attachment, []error) {
	rs, err := c.FetchAttachments(ctx, attType, ids...)
	if err != nil {
		return nil, multiplyErr(err, len(ids))
	}

	rmap := make(map[uuid.UUID][]models.Attachment, len(rs))
	for _, r := range rs {
		rmap[r.Owner] = append(rmap[r.Owner], r)
	}

	result := make([][]models.Attachment, len(ids))
	for i, id := range ids {
		result[i] = rmap[id]
	}
	return result, nil
}

func fetchUsers(ctx context.Context, contr *controller.User, ids []uuid.UUID) ([]models.User, []error) {
	users, err := contr.FetchUsers(ctx, ids)
	if err != nil {
		return nil, multiplyErr(err, len(ids))
	}

	umap := make(map[uuid.UUID]models.User, len(users))
	for _, user := range users {
		umap[user.ID] = user
	}

	result := make([]models.User, len(ids))
	for i, id := range ids {
		result[i] = umap[id]
	}

	return result, nil
}
