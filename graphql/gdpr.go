package graphql

import (
	"context"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

type GDPRRequest struct {
	models.GDPRRequest
	Attachments []models.Attachment
	Files       []graphql.Upload
}

func (GDPRRequest) IsEntity() {}

func (r *mutationResolver) SendGDPRRequest(ctx context.Context, req GDPRRequest) (*GDPRRequest, error) {
	ups := make([]controller.Upload, len(req.Files))
	for i := range req.Files {
		u := controller.Upload{
			File:        req.Files[i].File,
			Filename:    req.Files[i].Filename,
			Size:        req.Files[i].Size,
			ContentType: req.Files[i].ContentType,
		}

		ups[i] = u
	}

	return &req, r.gdpr.SendRequest(ctx, &req.GDPRRequest, ups)
}

func (r *queryResolver) ListGDPRRequests(ctx context.Context,
	first, offset *int) (*PaginatedList, error) {
	if first == nil {
		first = new(int)
	}
	if offset == nil {
		offset = new(int)
	}

	docs, total, err := r.gdpr.List(ctx, *first, *offset)
	if err != nil {
		return nil, err
	}
	var res = make([]Entity, len(docs))
	for i, doc := range docs {
		res[i] = newGDPRRequest(&doc)
	}
	return &PaginatedList{
		Entities:   res,
		TotalCount: total,
	}, err
}

func (r *queryResolver) GetGDPRRequest(ctx context.Context, rID uuid.UUID) (*GDPRRequest, error) {
	doc, err := r.gdpr.Get(ctx, rID)
	return newGDPRRequest(doc), err
}

func newGDPRRequest(doc *models.Document) *GDPRRequest {
	if doc == nil || doc.Kind != "gdpr_request" {
		return nil
	}

	g := doc.Data.(*models.GDPRRequest)
	atts := make([]models.Attachment, 0, len(doc.Attachments))
	for _, at := range doc.Attachments {
		atts = append(atts, at)
	}

	return &GDPRRequest{GDPRRequest: *g, Attachments: atts}
}
