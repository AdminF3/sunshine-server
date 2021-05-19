package models

import (
	"encoding/json"
	"log"
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Project struct {
	Value

	Name string `json:"name" validate:"required"`

	// Owner of the project - an valid organization.
	Owner uuid.UUID `json:"owner" validate:"required"`
	Asset uuid.UUID `json:"asset" validate:"required"`

	// ConsortiumOrgs is all organizations that its staff has
	// (read) access to the project.
	ConsortiumOrgs     pq.StringArray `json:"consortium_organizations" gorm:"type:text[]"`
	Roles              ProjRoles      `json:"roles"`
	Status             ProjectStatus  `json:"status" validate:"required"`
	AirTemperature     float64        `json:"airtemp" `
	FlatAirTemperature float64        `json:"flat_airtemp" `
	WaterTemperature   float64        `json:"watertemp" `
	GuaranteedSavings  float64        `json:"savings" `
	ConstructionFrom   time.Time      `json:"construction_from"`
	ConstructionTo     time.Time      `json:"construction_to"`
	ContractTerm       int            `json:"contract_term"`
	FirstYear          int            `json:"first_year"`
	Country            Country        `json:"country" validate:"required"`
	PortfolioDirector  uuid.UUID      `json:"portfolio_director"`
	Milestone          Milestone      `json:"milestone"`
	FundManager        *uuid.UUID     `json:"fund_manager" gorm:"type:uuid; null"`
	WorkPhase          WorkPhase
	MonitoringPhase    MonitoringPhase
	CommissioningDate  time.Time

	AssetSnapshot AssetSnapshot `json:"asset_snapshot" gorm:"EMBEDDED;EMBEDDED_PREFIX:asset_"`

	ProjectRoles []ProjectRole `json:"-" gorm:"foreignkey:ProjectID"`

	ForfaitingApplication *ForfaitingApplication `gorm:"foreignkey:project_id"`
	ForfaitingPayment     *ForfaitingPayment     `gorm:"foreignkey:project_id"`
	IsFAApproved          bool                   `gorm:"-"`

	Comments []ProjectComment `gorm:"foreignkey:ProjectID;PRELOAD:true"`
}

// ProjectComment is comment of project.
type ProjectComment struct {
	Value
	ProjectID uuid.UUID
	Author    CommentAuthor `gorm:"foreignKey:UserID;PRELOAD:true"`
	UserID    uuid.UUID     `gorm:"column:author"`
	Content   string
	Topic     string
}

// CommentAuthor is author of comment.
type CommentAuthor struct {
	Value

	Name  string `gorm:"column:name"`
	Email string
}

// TableName returns name of project comment authors table.
func (CommentAuthor) TableName() string {
	return "users"
}

func (ProjectComment) Kind() string      { return "project_comment" }
func (pc ProjectComment) Key() string    { return pc.ID.String() }
func (ProjectComment) TableName() string { return "project_comments" }

func (Project) Kind() string {
	return "project"
}

func (p Project) Key() string {
	return p.Name
}

func (p Project) Dependencies() []config.Dependency {
	var dep = []config.Dependency{
		{ID: p.Owner, Kind: "organization"},
		{ID: p.Asset, Kind: "asset"},
		{ID: p.PortfolioDirector, Kind: "user"},
	}

	for _, org := range p.ConsortiumOrgs {
		dep = append(dep, config.Dependency{
			ID:   uuid.MustParse(org),
			Kind: "organization",
		})
	}

	return addRoles(dep, p.Roles.PM, p.Roles.PLSign, p.Roles.PaCo, p.Roles.TaMa, p.Roles.TeMe)
}

func (Project) TableName() string {
	return "projects"
}

func (p *Project) MarshalJSON() ([]byte, error) {
	for _, r := range p.ProjectRoles {
		p.Roles.add(r)
	}
	p.Roles.make()

	type prjt Project
	return json.Marshal(prjt(*p))
}

func (p *Project) ConvertRoles() {
	roles := make([]ProjectRole, 0, p.Roles.len())

	positions := []struct {
		values   []uuid.UUID
		position string
	}{
		{values: p.Roles.PM, position: "pm"},
		{values: p.Roles.PaCo, position: "paco"},
		{values: p.Roles.PLSign, position: "plsign"},
		{values: p.Roles.TaMa, position: "tama"},
		{values: p.Roles.TeMe, position: "teme"},
	}

	for _, sp := range positions {
		for _, id := range sp.values {
			roles = append(roles, ProjectRole{
				UserID:    id,
				ProjectID: p.ID,
				Position:  sp.position,
			})
		}
	}

	if len(p.ProjectRoles) == 0 {
		p.ProjectRoles = roles
	} else {
		p.ProjectRoles = append(p.ProjectRoles, roles...)
	}
}

func addRoles(dep []config.Dependency, roles ...[]uuid.UUID) []config.Dependency {
	for _, r := range roles {
		for _, v := range r {
			dep = append(dep, config.Dependency{ID: v, Kind: "user"})
		}
	}

	return dep
}

type ProjRoles struct {
	PLSign []uuid.UUID `json:"plsign"` // Project Legal Signatories
	PM     []uuid.UUID `json:"pm"`     // Project Manager
	PaCo   []uuid.UUID `json:"paco"`   // Participant Contact
	TaMa   []uuid.UUID `json:"tama"`   // Task manager
	TeMe   []uuid.UUID `json:"teme"`   // Team Member
}

func (p ProjRoles) len() int {
	return len(p.PM) + len(p.PaCo) + len(p.PLSign) +
		len(p.TaMa) + len(p.TeMe)
}

func (p *ProjRoles) add(v ProjectRole) {
	switch v.Position {
	case "pm":
		p.PM = appendUnique(p.PM, v.UserID)
	case "paco":
		p.PaCo = appendUnique(p.PaCo, v.UserID)
	case "plsign":
		p.PLSign = appendUnique(p.PLSign, v.UserID)
	case "tama":
		p.TaMa = appendUnique(p.TaMa, v.UserID)
	case "teme":
		p.TeMe = appendUnique(p.TeMe, v.UserID)
	default:
		// TODO: Report once we have sentry
		log.Printf("We got role in strange position: %v", v.Position)
	}
}

func (p *ProjRoles) make() {
	if p.PM == nil {
		p.PM = make([]uuid.UUID, 0)
	}
	if p.PaCo == nil {
		p.PaCo = make([]uuid.UUID, 0)
	}
	if p.PLSign == nil {
		p.PLSign = make([]uuid.UUID, 0)
	}
	if p.TaMa == nil {
		p.TaMa = make([]uuid.UUID, 0)
	}
	if p.TeMe == nil {
		p.TeMe = make([]uuid.UUID, 0)
	}
}

type ProjectRole struct {
	// Does not embed Value because we don't want soft deletion for roles.

	ID        uuid.UUID `gorm:"type:uuid; primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID    uuid.UUID `json:"user_id"`
	ProjectID uuid.UUID `json:"project_id"`
	Position  string    `json:"position"`
}

func (ProjectRole) TableName() string {
	return "project_roles"
}

// BeforeCreate ignores inserting the same role for the same user.
//
// In general we would want to DO NOTHING on conflict, but GORM appends
// "RETURNING <table_name>.<primary_key>" at the end of each query and expects
// to receive a non-empty result on INSERT. Setting DO NOTHING will not insert
// anything and GORM will not receive an ID of inserted row thus leading to
// an obscure error: "sql: no rows in result set"
func (p *ProjectRole) BeforeCreate(scope *gorm.Scope) error {
	scope.Set("gorm:insert_option", `ON CONFLICT
		ON CONSTRAINT project_roles_user_id_project_id_position_key
		DO UPDATE SET updated_at = now()`)
	return nil
}

// AfterFind is GORM hook for querying.
func (p *Project) AfterFind() (err error) {
	if p.ForfaitingApplication == nil {
		return
	}

	if p.ForfaitingApplication.Reviews == nil {
		return
	}

	for _, r := range p.ForfaitingApplication.Reviews {
		if r.Type == FAReviewTypeExecutive {
			p.IsFAApproved = r.Approved
		}
	}

	return
}

// AssetSnapshot holds copied values from Asset during project's creation.
type AssetSnapshot struct {
	Owner        uuid.UUID  `json:"owner" gorm:"column:owner_id"`
	ESCO         *uuid.UUID `json:"esco" gorm:"column:esco_id"`
	Area         int        `json:"area"`
	HeatedArea   int        `json:"heated_area"`
	BillingArea  int        `json:"billing_area"`
	Flats        int        `json:"flats"`
	Floors       int        `json:"floors"`
	StairCases   int        `json:"stair_cases"`
	BuildingType Building   `json:"building_type"`
	HeatingType  Heating    `json:"heating_type"`
	Cadastre     string     `json:"cadastre"`
}
