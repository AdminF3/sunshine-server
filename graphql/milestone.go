package graphql

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

// WorkPhase encapsulates models.WorkPhase with its attachments.
type WorkPhase struct {
	models.WorkPhase
	Attachments []models.Attachment
}

func newWP(doc *models.Document) *WorkPhase {
	if doc == nil || doc.Kind != "work_phase" {
		return nil
	}
	wp := doc.Data.(*models.WorkPhase)
	atts := make([]models.Attachment, 0, len(doc.Attachments))
	for _, v := range doc.Attachments {
		atts = append(atts, v)
	}

	return &WorkPhase{WorkPhase: *wp, Attachments: atts}
}

func (r *mutationResolver) AdvanceProjectToWorkPhase(ctx context.Context, pid uuid.UUID) (*WorkPhase, error) {
	doc, err := r.wp.AdvanceToWorkPhase(ctx, pid)
	if err != nil {
		return nil, err
	}
	return newWP(doc), nil
}

func (r *queryResolver) GetWorkPhase(ctx context.Context, wpid uuid.UUID) (*WorkPhase, error) {
	doc, _, err := r.wp.GetWP(ctx, wpid)
	if err != nil {
		return nil, err
	}
	return newWP(doc), nil
}

// MonitoringPhase encapsulates models.MonitoringPhase with its attachments.
type MonitoringPhase struct {
	models.MonitoringPhase
	Attachments []models.Attachment
}

func newMP(doc *models.Document) *MonitoringPhase {
	if doc == nil || doc.Kind != "monitoring_phase" {
		return nil
	}
	mp := doc.Data.(*models.MonitoringPhase)
	atts := make([]models.Attachment, 0, len(doc.Attachments))
	for _, v := range doc.Attachments {
		atts = append(atts, v)
	}

	return &MonitoringPhase{MonitoringPhase: *mp, Attachments: atts}
}

func (r *mutationResolver) AdvanceProjectToMonitoringPhase(ctx context.Context, pid uuid.UUID) (*MonitoringPhase, error) {
	doc, err := r.mp.AdvanceToMonitoringPhase(ctx, pid)
	if err != nil {
		return nil, err
	}
	return newMP(doc), nil
}

func (r *queryResolver) GetMonitoringPhase(ctx context.Context, mpid uuid.UUID) (*MonitoringPhase, error) {
	doc, _, err := r.mp.GetMP(ctx, mpid)
	if err != nil {
		return nil, err
	}
	return newMP(doc), nil
}

func (r *mutationResolver) ReviewWorkPhase(ctx context.Context, id uuid.UUID, review models.WPReview) (*Message, error) {
	return messageResult(r.wp.ReviewWP(ctx, id, review))
}

func (r *mutationResolver) ReviewMonitoringPhase(ctx context.Context, id uuid.UUID, review models.MPReview) (*Message, error) {
	return messageResult(r.mp.ReviewMP(ctx, id, review))
}
