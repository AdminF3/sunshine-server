package graphql

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

// Meeting encapsulates models.Meeting with its attachments.
type Meeting struct {
	models.Meeting
	Attachments []models.Attachment
}

func newMeeting(doc *models.Document) *Meeting {
	if doc == nil || doc.Kind != "meeting" {
		return nil
	}
	m := doc.Data.(*models.Meeting)
	atts := make([]models.Attachment, 0, len(doc.Attachments))
	for _, v := range doc.Attachments {
		atts = append(atts, v)
	}

	return &Meeting{Meeting: *m, Attachments: atts}
}

func (r *mutationResolver) CreateMeeting(ctx context.Context, meeting Meeting) (*Meeting, error) {
	doc, err := r.meet.Create(ctx, &meeting.Meeting)
	return newMeeting(doc), err
}

func (r *queryResolver) GetMeeting(ctx context.Context, mID uuid.UUID) (*Meeting, error) {
	doc, err := r.meet.Get(ctx, mID)
	return newMeeting(doc), err
}

func (r *mutationResolver) DeleteMeeting(ctx context.Context, id uuid.UUID) (*Message, error) {
	err := r.meet.Delete(ctx, id)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) UpdateMeeting(ctx context.Context, meeting Meeting) (*Meeting, error) {
	doc, err := r.meet.Update(ctx, meeting.Meeting)
	return newMeeting(doc), err
}

func (r *queryResolver) ListMeetings(ctx context.Context, id *uuid.UUID) ([]*Meeting, error) {
	docs, err := r.meet.List(ctx, id)
	if err != nil {
		return nil, err
	}
	var res = make([]*Meeting, len(docs))
	for i, doc := range docs {
		res[i] = newMeeting(&doc)
	}
	return res, err
}
