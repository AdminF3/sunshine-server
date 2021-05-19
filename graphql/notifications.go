package graphql

import (
	"context"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/google/uuid"
)

func (r *queryResolver) GetNotification(ctx context.Context, nID uuid.UUID) (*models.Notification, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, controller.ErrUnauthorized
	}

	return r.env.Notifier.Get(ctx, nID, cv.User.ID)
}

func (r *queryResolver) ListNotifications(ctx context.Context, action *models.UserAction) ([]models.Notification, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, controller.ErrUnauthorized
	}

	return r.env.Notifier.List(ctx, cv.User.ID, action)
}

func (r *mutationResolver) SeeNotification(ctx context.Context, nID uuid.UUID) (*Message, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return msgErr, controller.ErrUnauthorized
	}

	return msgOK, r.env.Notifier.See(ctx, nID, cv.User.ID)
}

func (r *queryResolver) NotificationListing(ctx context.Context,
	first *int, after *string,
	last *int, before *string,
	targetKey *string, targetType *models.EntityType,
	action []models.UserAction, userID *uuid.UUID, targetID *uuid.UUID, seen *bool,
	country *string,
) (*NotificationConnection, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, controller.ErrUnauthorized
	}
	firstValue := 0
	if first != nil {
		firstValue = *first
	}
	lastValue := 0
	if last != nil {
		lastValue = *last
	}

	var c *models.Country
	if country != nil {
		cc := models.Country(*country)
		c = &cc
	}
	total, err := r.env.Notifier.Count(ctx, cv.User.ID, action, userID, targetID, seen, targetKey, targetType, c)
	if err != nil {
		return nil, err
	}

	offset, limit := calcBounds(firstValue, decodeCursor(after), lastValue, decodeCursor(before), total)
	records, err := r.env.Notifier.Filter(ctx, cv.User.ID, offset, limit, action, userID, targetID, seen, targetKey, targetType, c)
	return newNotificationConnection(offset, limit, total, records), err
}

func newNotificationConnection(offset, limit, total int, records []models.Notification) *NotificationConnection {
	edges := make([]NotificationEdge, len(records))
	for i := range edges {
		index := offset + i
		edges[i] = NotificationEdge{
			Node:   &records[i],
			Cursor: encodeCursor(index),
		}
	}

	pi := pageInfo(offset, limit, total)
	if len(edges) > 0 {
		pi.StartCursor = edges[0].Cursor
		pi.EndCursor = edges[len(edges)-1].Cursor
	}

	return &NotificationConnection{Edges: edges, PageInfo: pi, TotalCount: total}
}
