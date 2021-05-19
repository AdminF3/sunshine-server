package models

import "github.com/google/uuid"

type ProjectCreationRequest struct {
	Value

	Asset        uuid.UUID  `gorm:"column:asset_id"`
	Organization uuid.UUID  `gorm:"column:organization_id"`
	User         uuid.UUID  `gorm:"column:user_id"`
	Token        *uuid.UUID `gorm:"column:token_id"`
	Status       ProjectCreationRequestStatus
}

func (ProjectCreationRequest) TableName() string {
	return "create_project_request"
}

func (ProjectCreationRequest) Dependencies() []Dependency { return nil }
func (r ProjectCreationRequest) Key() string              { return r.ID.String() }
func (ProjectCreationRequest) Kind() string               { return "project_creation_request" }
