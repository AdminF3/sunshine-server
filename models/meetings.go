package models

import (
	"fmt"
	"strings"
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/google/uuid"
)

type Meeting struct {
	Value

	Name string `json:"name" validate:"required"`

	// Host is the organization that creates and hosts the meeting.
	Host       uuid.UUID     `json:"host" validate:"required"`
	HostEntity *Organization `gorm:"foreignkey:Host"`

	Location     string     `json:"location" validate:"required"`
	Date         time.Time  `json:"date" validate:"required"`
	Objective    string     `json:"objective"`
	Stakeholder  LegalForm  `json:"stakeholder" validate:"required"`
	Stage        string     `json:"stage"`
	Notes        string     `json:"notes"`
	ActionsTaken string     `json:"actions_taken"`
	NextContact  *time.Time `json:"next_contact"`

	Project       *uuid.UUID `json:"project"`
	ProjectEntity *Project   `gorm:"foreignkey:Project"`

	Topic  *MeetingType   `json:"topic"`
	Guests []MeetingGuest `gorm:"foreignkey:MeetingID"`

	// InternalProject is Related Organization's internal project
	InternalProject string
}

func (Meeting) Kind() string {
	return "meeting"
}

func (m Meeting) Key() string {
	return m.ID.String()
}

func (Meeting) TableName() string {
	return "meetings"
}

func (m Meeting) Dependencies() []config.Dependency {
	return []config.Dependency{}
}

type MeetingGuest struct {
	Value

	MeetingID uuid.UUID
	Name      string
	Type      StakeholdersType
	Email     string
	Phone     string

	// Organization name
	Organization string
}

func (mg *MeetingGuest) String() string {
	return fmt.Sprintf("name: %q, type: %q, email: %q, phone: %q",
		mg.Name, mg.Type, mg.Email, mg.Phone)
}

func (MeetingGuest) TableName() string {
	return "meeting_guests"
}

// Parse2String transforms the meeting entity into array of strings
func (m *Meeting) Parse2String() []string {
	prjname := ""
	if m.ProjectEntity != nil {
		prjname = m.ProjectEntity.Name
	}

	nc := ""
	if m.NextContact != nil {
		nc = m.NextContact.Format(time.RFC822)
	}

	t := ""
	if m.Topic != nil {
		t = string(*m.Topic)
	}

	return []string{
		m.Name,
		m.HostEntity.Name,
		m.Location,
		m.Date.Format(time.RFC822),
		m.Objective,
		m.Stakeholder.String(),
		m.Stage,
		m.Notes,
		m.ActionsTaken,
		nc,
		prjname,
		t,
		guestsString(m.Guests),
		m.InternalProject,
	}
}

func guestsString(mg []MeetingGuest) string {
	var b strings.Builder

	for _, g := range mg {
		b.WriteString(g.String())
	}

	return b.String()
}
