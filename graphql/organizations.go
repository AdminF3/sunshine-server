package graphql

import (
	"context"
	"fmt"
	"strings"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/google/uuid"
)

const alterOrgRole = `{"position": %q, "user": %q}`

func (r *mutationResolver) AddOrganizationRole(ctx context.Context, oid, user uuid.UUID, role models.OrganizationRole) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}
	body := strings.NewReader(fmt.Sprintf(alterOrgRole, role.Position, user))
	_, _, err := r.org.AddRole(ctx, oid, body)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) RemoveOrganizationRole(ctx context.Context, oid, user uuid.UUID, role models.OrganizationRole) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}
	body := strings.NewReader(fmt.Sprintf(alterOrgRole, role.Position, user))
	_, _, err := r.org.RemoveRole(ctx, oid, body)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) RequestOrganizationMembership(ctx context.Context, oid uuid.UUID) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}
	err := r.org.RequestOrganizationMembership(ctx, oid)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *queryResolver) ListOrganizationReports(ctx context.Context, first, offset *int) (*PaginatedList, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, controller.ErrUnauthorized
	}

	if first == nil {
		first = new(int)
	}
	if offset == nil {
		offset = new(int)
	}

	reports, total, err := r.org.GetReport(ctx, *first, *offset)

	result := make([]Entity, len(reports))
	for i, r := range reports {
		result[i] = r
	}

	return &PaginatedList{
		Entities:   result,
		TotalCount: total,
	}, err
}

func (r *mutationResolver) AcceptLEARApplication(ctx context.Context, uid, oid uuid.UUID, comment string, filename string, approved bool) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}

	err := r.org.AcceptLEARApplication(ctx, uid, oid, comment, filename, approved)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}
