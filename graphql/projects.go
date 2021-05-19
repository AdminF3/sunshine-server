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

const alterProjRole = `{"position": %q, "user": %q}`

func (r *mutationResolver) AssignPm(ctx context.Context, pid uuid.UUID, pm []uuid.UUID) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}

	return msgOK, r.project.AssignProjectRoles(ctx, pid, pm, controller.AssignPM)
}

func (r *mutationResolver) AddProjectRole(ctx context.Context, pid, user uuid.UUID, role models.ProjectRole) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}
	body := strings.NewReader(fmt.Sprintf(alterProjRole, role.Position, user))
	_, _, err := r.project.AddRole(ctx, pid, body)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) RemoveProjectRole(ctx context.Context, pid, user uuid.UUID, role models.ProjectRole) (*Message, error) {
	if !services.FromContext(ctx).Authorized() {
		return msgErr, controller.ErrUnauthorized
	}
	body := strings.NewReader(fmt.Sprintf(alterProjRole, role.Position, user))
	_, _, err := r.project.RemoveRole(ctx, pid, body)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) RequestProjectCreation(ctx context.Context, assetID, orgID uuid.UUID) (*Message, error) {
	err := r.project.RequestProjectCreation(ctx, assetID, orgID)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) ProcessProjectCreation(ctx context.Context, userID, assetID uuid.UUID, approve bool) (*Message, error) {
	err := r.project.ProcessProjectRequest(ctx, userID, assetID, approve)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}

func (r *mutationResolver) CommentProject(ctx context.Context, id uuid.UUID, content string, topic *string) (*models.Project, error) {
	return r.project.CommentProject(ctx, id, content, topic)
}

func (r *mutationResolver) AdvanceToMilestone(ctx context.Context, projectID uuid.UUID, nextMilestone models.Milestone) (*Message, error) {
	err := r.project.AdvanceToMilestone(ctx, projectID, nextMilestone)
	if err != nil {
		return msgErr, err
	}
	return msgOK, nil
}
