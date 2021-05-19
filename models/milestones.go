package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type Milestone string

const (
	MilestoneZero               Milestone = "zero"
	MilestoneAcquisitionMeeting Milestone = "acquisition_meeting"
	MilestoneFeasibilityStudy   Milestone = "feasibility_study"
	MilestoneCommitmentStudy    Milestone = "commitment_study"
	MilestoneProjectDesign      Milestone = "project_design"
	MilestoneProjectPreparation Milestone = "project_preparation"
	MilestoneKickOffMeeting     Milestone = "kick_off_meeting"
	MilestoneWorkPhase          Milestone = "work_phase"
	MilestoneMonitoringPhase    Milestone = "monitoring_phase"
	MilestoneCommissioning      Milestone = "commissioning"
	MilestoneForfaitingPayment  Milestone = "forfaiting_payment"
)

func (m Milestone) Index() int8 {
	switch m {
	case MilestoneZero:
		return 0
	case MilestoneAcquisitionMeeting:
		return 1
	case MilestoneFeasibilityStudy:
		return 2
	case MilestoneCommitmentStudy:
		return 3
	case MilestoneProjectDesign:
		return 4
	case MilestoneProjectPreparation:
		return 5
	case MilestoneKickOffMeeting:
		return 6
	case MilestoneWorkPhase:
		return 7
	case MilestoneMonitoringPhase:
		return 8
	case MilestoneCommissioning:
		return 9
	case MilestoneForfaitingPayment:
		return 10
	default:
		panic("Unexpected milestone value: " + m)
	}
}

// Scan implements the database/sql.Scanner interface.
func (m *Milestone) Scan(src interface{}) error {
	var err error
	switch v := src.(type) {
	case []byte:
		*m = Milestone(v)
	case string:
		*m = Milestone(v)
	case fmt.Stringer:
		*m = Milestone(v.String())
	default:
		err = fmt.Errorf("bad value: %[1]T(%[1]s)", src)
	}

	return err
}

// Value implements the database/sql/driver.Valuer interface.
func (m Milestone) Value() (driver.Value, error) { return string(m), nil }

type WorkPhase struct {
	Value
	Project     uuid.UUID    `gorm:"column:project_id"`
	Attachments []Attachment `gorm:"foreignKey:owner_id;preload:false"`
	Reviews     []WPReview   `gorm:"foreignkey:WPID;preload:false"`
}

func (WorkPhase) Kind() string               { return "work_phase" }
func (wp WorkPhase) Key() string             { return wp.ID.String() }
func (WorkPhase) TableName() string          { return "work_phase" }
func (WorkPhase) Dependencies() []Dependency { return nil }

type WPReview struct {
	Value

	WPID     uuid.UUID
	Author   *uuid.UUID
	Approved bool
	Comment  string
	Type     WPReviewType
}

func (WPReview) TableName() string          { return "wp_reviews" }
func (WPReview) Kind() string               { return "wp_review" }
func (wp WPReview) Key() string             { return wp.ID.String() }
func (WPReview) Dependencies() []Dependency { return nil }

type WPReviewType int8

const (
	_ WPReviewType = iota
	WPReviewTypeFinancial
	WPReviewTypeTechnical
	WPReviewTypeBankAccount
	WPReviewTypeExecutive
	WPReviewTypeMaintenance
)

type MonitoringPhase struct {
	Value
	Project     uuid.UUID    `gorm:"column:project_id"`
	Attachments []Attachment `gorm:"foreignKey:owner_id;preload:false"`
	Reviews     []MPReview   `gorm:"foreignkey:MPID;PRELOAD:true"`
}

func (MonitoringPhase) Kind() string               { return "monitoring_phase" }
func (mp MonitoringPhase) Key() string             { return mp.ID.String() }
func (MonitoringPhase) TableName() string          { return "monitoring_phase" }
func (MonitoringPhase) Dependencies() []Dependency { return nil }

type MPReview struct {
	Value

	MPID     uuid.UUID
	Author   *uuid.UUID
	Approved bool
	Comment  string

	Type MPReviewType
}

func (MPReview) TableName() string          { return "mp_reviews" }
func (MPReview) Kind() string               { return "mp_review" }
func (mp MPReview) Key() string             { return mp.ID.String() }
func (MPReview) Dependencies() []Dependency { return nil }

type MPReviewType int8

const (
	_ MPReviewType = iota
	MPReviewTypeForfaiting
)
