package models

import (
	"encoding/json"
	"log"
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Organization struct {
	Value

	Name               string           `json:"name" validate:"required"`
	VAT                string           `json:"vat" `
	RegistrationNumber string           `json:"registration_number"`
	Address            string           `json:"address" validate:"required"`
	Telephone          string           `json:"telephone"`
	Website            string           `json:"website" validate:"omitempty,url"`
	Logo               string           `json:"logo"`
	LegalForm          LegalForm        `json:"legal_form" validate:"required"`
	Registered         time.Time        `json:"registered"  gorm:"column:registered_at"`
	Valid              ValidationStatus `json:"valid" gorm:"column:status"`
	Roles              OrgRoles         `json:"roles" validate:"required"`
	Email              string           `json:"email" validate:"omitempty,email"`
	Country            Country          `json:"country" validate:"required"`

	OrganizationRoles []OrganizationRole `json:"-" gorm:"foreignkey:OrganizationID"`
}

// ExcerptString returns a stringified json with the fields of an
// organization that when updated invalidate it.
func (o Organization) ExcerptString() string {
	type coreInfo struct {
		Name               string
		Address            string
		LegalForm          LegalForm
		Valid              ValidationStatus
		VAT                string
		RegistrationNumber string
	}
	ci := coreInfo{Name: o.Name, Address: o.Address, LegalForm: o.LegalForm, Valid: o.Valid, VAT: o.VAT, RegistrationNumber: o.RegistrationNumber}
	out, err := json.Marshal(ci)
	if err != nil {
		return ""
	}
	return string(out)
}

func (Organization) Kind() string {
	return "organization"
}

func (o Organization) Key() string {
	return o.VAT
}

func (o Organization) Dependencies() []config.Dependency {
	return addRoles(
		[]config.Dependency{{ID: o.Roles.LEAR, Kind: "user"}},
		o.Roles.LSigns, o.Roles.LEAAs, o.Roles.Members,
	)
}

func (Organization) TableName() string {
	return "organizations"
}

func (o *Organization) MarshalJSON() ([]byte, error) {
	for _, r := range o.OrganizationRoles {
		o.Roles.add(r)
	}
	o.Roles.make()

	type orgt Organization
	return json.Marshal(orgt(*o))
}

func (o *Organization) ConvertRoles() {
	var (
		roles     = make([]OrganizationRole, 0, o.Roles.Len())
		positions = []struct {
			values   []uuid.UUID
			position string
		}{
			{values: o.Roles.LSigns, position: "lsign"},
			{values: o.Roles.LEAAs, position: "leaa"},
			{values: o.Roles.Members, position: "member"},
		}
	)

	if len(o.Roles.LEAR) > 0 {
		roles = append(roles, OrganizationRole{
			UserID:         o.Roles.LEAR,
			OrganizationID: o.ID,
			Position:       "lear",
		})
	}

	for _, sp := range positions {
		for _, id := range sp.values {
			roles = append(roles, OrganizationRole{
				UserID:         id,
				OrganizationID: o.ID,
				Position:       sp.position,
			})
		}
	}

	if len(o.OrganizationRoles) == 0 {
		o.OrganizationRoles = roles
	} else {
		o.OrganizationRoles = append(o.OrganizationRoles, roles...)
	}
}

type OrgRoles struct {
	LEAR    uuid.UUID   `json:"lear"`    // Legal Entity Appointed Representative
	LSigns  []uuid.UUID `json:"lsigns"`  // Legal Signatories
	LEAAs   []uuid.UUID `json:"leaas"`   // Legal Entity Account Administrators
	Members []uuid.UUID `json:"members"` // Regular oragnization members with no exclusive previliges
}

func (o OrgRoles) Len() int {
	return len(o.LSigns) + len(o.LEAAs) + len(o.Members) + 1
}

func (o *OrgRoles) add(v OrganizationRole) {
	switch v.Position {
	case "lear":
		o.LEAR = v.UserID
	case "lsign":
		o.LSigns = appendUnique(o.LSigns, v.UserID)
	case "leaa":
		o.LEAAs = appendUnique(o.LEAAs, v.UserID)
	case "member":
		o.Members = appendUnique(o.Members, v.UserID)
	default:
		// TODO: Report once we have sentry
		log.Printf("We got role in strange position: %v", v.Position)
	}
}

func (o *OrgRoles) make() {
	if o.LSigns == nil {
		o.LSigns = make([]uuid.UUID, 0)
	}
	if o.LEAAs == nil {
		o.LEAAs = make([]uuid.UUID, 0)
	}
	if o.Members == nil {
		o.Members = make([]uuid.UUID, 0)
	}
}

// OrganizationRole is normalized view stored in the PostgreSQL.
type OrganizationRole struct {
	// Does not embed Value because we don't want soft deletion for roles.

	ID        uuid.UUID `gorm:"type:uuid; primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Position       string    `json:"position"`
}

func (OrganizationRole) TableName() string {
	return "organization_roles"
}

// BeforeCreate ignores inserting the same role for the same user.
//
// In general we would want to DO NOTHING on conflict, but GORM appends
// "RETURNING <table_name>.<primary_key>" at the end of each query and expects
// to receive a non-empty result on INSERT. Setting DO NOTHING will not insert
// anything and GORM will not receive an ID of inserted row thus leading to
// an obscure error: "sql: no rows in result set"
func (or *OrganizationRole) BeforeCreate(scope *gorm.Scope) error {
	scope.Set("gorm:insert_option", `ON CONFLICT
		ON CONSTRAINT organization_roles_user_id_organization_id_position_key
		DO UPDATE SET updated_at = now()`)
	return nil
}
