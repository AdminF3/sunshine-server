package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// OrganizationProjectsReport is a report of organization projects.
type OrganizationProjectsReport struct {
	// TotalCount shows the total count of all projects.
	TotalCount int

	// OngoingCount shows the count of ongoing projects.
	OngoingCount int

	// MonitoringPhaseCount shows count of projects in monitoring phase.
	MonitoringPhaseCount int

	// ApprovedForfaitingCount shows the total projects approved for forfaiting.
	ApprovedForfaitingCount int
}

// OrganizationReportsList represent OrganizationID to Org Report
type OrganizationReportsList map[uuid.UUID]OrganizationReport

func (list OrganizationReportsList) Transform() []OrganizationReport {
	var result = make([]OrganizationReport, len(list))

	idx := 0
	for id, report := range list {
		result[idx] = report
		result[idx].ID = id

		idx++
	}

	return result
}

// OrganizationReport is a report of organization data.
type OrganizationReport struct {
	// ID of the organization.
	ID uuid.UUID
	// Name of the organization,
	Name string
	// VAT of the organization.
	VAT string
	// Address of the organization.
	Address string
	// Telephone of the organization.
	Telephone string
	// Website of the organization.
	Website string
	// Legal form of the organization.
	LegalForm LegalForm
	// Registration date of the organization.
	Registered time.Time
	// Valid status of the organization.
	Valid ValidationStatus
	// Roles of the organization.
	Roles OrgRoles
	// Email of the organization
	Email string
	// Country of the organization
	Country Country

	// OwnProjects are projects that the organization is owner and
	// the asset for that project is also hers.
	OwnProjects OrganizationProjectsReport

	// RelatedProjects are projects that or the org is owner but
	// not asset's owner or vice versa (owns the asset but not the
	// projects).
	RelatedProjects OrganizationProjectsReport

	LearEmail string
	LearName  string

	// UsersCount shows the total count of users that have role in the organization.
	UsersCount int

	RegistrationNumber string
}

func (OrganizationReport) IsEntity() {}

type (
	set []struct {
		ID    uuid.UUID
		Name  string
		Email string
		Count int
	}

	where func(*gorm.DB) *gorm.DB
)

func (list OrganizationReportsList) PopulateLear(db *gorm.DB, ids []uuid.UUID) error {
	var s set

	err := db.Select("organizations.ID, users.Name, users.Email").
		Table("organizations").
		Joins("inner join organization_roles on organizations.ID = organization_roles.organization_id").
		Joins("inner join users on users.ID = organization_roles.user_id").
		Where("organizations.ID IN (?)", ids).
		Group("organizations.ID, users.Name, users.Email").
		Find(&s).Error

	for _, el := range s {
		prjreport := list[el.ID]
		prjreport.LearEmail = el.Email
		prjreport.LearName = el.Name

		list[el.ID] = prjreport
	}

	return err
}

func (list OrganizationReportsList) CountOngoingProjects(db *gorm.DB, ids []uuid.UUID, isOwned bool) error {
	set, err := countPrjs(db, ids, isOwned,
		func(db *gorm.DB) *gorm.DB { return db.Where("projects.status = ?", int(ProjectStatusInProgress)) })
	for _, s := range set {
		prjreport := list[s.ID]

		switch isOwned {
		case true:
			prjreport.OwnProjects.OngoingCount = s.Count
		case false:
			prjreport.RelatedProjects.OngoingCount = s.Count
		}

		list[s.ID] = prjreport
	}
	return err
}

func (list OrganizationReportsList) CountMonitoringPhaseProjects(db *gorm.DB, ids []uuid.UUID, isOwned bool) error {
	set, err := countPrjs(db, ids, isOwned,
		func(db *gorm.DB) *gorm.DB {
			return db.Joins("inner join monitoring_phase on projects.id = monitoring_phase.project_id").Where("monitoring_phase.project_id IS NOT NULL")
		})
	for _, s := range set {
		prjreport := list[s.ID]

		switch isOwned {
		case true:
			prjreport.OwnProjects.MonitoringPhaseCount = s.Count
		case false:
			prjreport.RelatedProjects.MonitoringPhaseCount = s.Count
		}

		list[s.ID] = prjreport
	}
	return err
}

func (list OrganizationReportsList) CountTotalProjects(db *gorm.DB, ids []uuid.UUID, isOwned bool) error {
	set, err := countPrjs(db, ids, isOwned)

	for _, s := range set {
		prjreport := list[s.ID]

		switch isOwned {
		case true:
			prjreport.OwnProjects.TotalCount = s.Count
		case false:
			prjreport.RelatedProjects.TotalCount = s.Count
		}

		list[s.ID] = prjreport
	}

	return err
}

func (list OrganizationReportsList) CountFAProjects(db *gorm.DB, ids []uuid.UUID, isOwned bool) error {
	s, err := countPrjs(db, ids, isOwned,
		func(db *gorm.DB) *gorm.DB {
			return db.Joins("inner join forfaiting_applications on forfaiting_applications.project_id = projects.id").
				Joins("inner join fa_reviews on forfaiting_applications.id = fa_reviews.forfaiting_application_id").
				Where("fa_reviews.approved = ?", true).
				Where("fa_reviews.type = ?", FAReviewTypeExecutive)
		})

	for _, el := range s {
		prjreport := list[el.ID]

		switch isOwned {
		case true:
			prjreport.OwnProjects.ApprovedForfaitingCount = el.Count
		case false:
			prjreport.RelatedProjects.ApprovedForfaitingCount = el.Count
		}

		list[el.ID] = prjreport
	}

	return err
}

func countPrjs(db *gorm.DB, ids []uuid.UUID, isOwned bool, w ...where) (set, error) {
	var s set

	db = db.Select("organizations.ID, count(projects.*)").
		Table("organizations").
		Joins("inner join projects on projects.owner = organizations.ID").
		Joins("inner join assets on assets.ID = projects.asset").
		Where("organizations.ID IN (?)", ids)

	switch isOwned {
	case true:
		db = db.Where("organizations.ID = assets.owner_id").
			Where("projects.owner = organizations.ID")
	case false:
		// XOR: ( A and not B) or ( not A and B)
		db = db.Where("(organizations.ID = assets.owner_id AND NOT projects.owner = organizations.ID) OR (NOT organizations.ID = assets.owner_id AND  projects.owner = organizations.ID)")
	}

	for _, cl := range w {
		db = cl(db)
	}

	return s, db.Group("organizations.ID").Find(&s).Error
}
