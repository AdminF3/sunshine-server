package controller

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/google/uuid"
)

// Action is defined by who is allowed to perform it.
type Action uint32

const (
	self = 1 << iota
	logged

	// Organization roles
	lear
	lsigns
	leaas
	members

	// Project roles
	pm
	paco
	plsign
	tama
	teme

	// Country roles
	pd
	fm
	dpo
	ca

	valid

	// Admin roles
	pfm // platform manager
	anm // admin network manager
	investor
	superuser
)

const (
	// user actions
	ListUsers             Action = logged
	ListAdminUsers        Action = superuser | pfm | anm
	GetUser               Action = superuser | pfm | anm | self | ca
	UpdateUser            Action = superuser | pfm | anm | self
	UploadUser            Action = superuser | pfm | anm | self
	ListUserAssets        Action = superuser | pfm | anm | self | pd
	ListUserOrganizations Action = superuser | pfm | anm | self | pd
	ListUserProjects      Action = superuser | pfm | anm | self | pd
	DownloadUserFile      Action = superuser | pfm | anm | self
	DeleteUserFile        Action = superuser | pfm | anm | self
	ValidateUser          Action = superuser | pfm | anm | ca

	// organization actions
	CreateOrganization            Action = logged
	GetOrganization               Action = logged
	UpdateOrganization            Action = superuser | pfm | anm | lear | leaas | ca
	UploadOrganization            Action = superuser | pfm | anm | lear | leaas | lsigns | ca
	DeleteOrgFile                 Action = superuser | pfm | anm | lear | leaas | lsigns | ca
	GetOrgMeetings                Action = superuser | pfm | anm | lear | leaas | lsigns | ca
	GetPrjMeetings                Action = superuser | pfm | anm | lear | leaas | lsigns | ca
	DownloadOrgFile               Action = superuser | pfm | anm | lear | lsigns | leaas | members | pd | ca
	AddOrganizationRole           Action = superuser | pfm | anm | lear | leaas | ca
	RemoveOrganizationRole        Action = superuser | pfm | anm | lear | leaas | ca
	ValidateOrganization          Action = superuser | pfm | anm | ca
	RequestOrganizationMembership Action = logged
	ClaimAssetResidency           Action = logged
	AcceptLEARApplication         Action = superuser | lear | ca

	// asset actions
	GetAsset          Action = logged
	UpdateAsset       Action = superuser | pfm | anm | pd | lear | leaas | lsigns | members | ca
	UploadAsset       Action = superuser | pfm | anm | lear | leaas | lsigns | members | ca
	DeleteAssetFile   Action = superuser | pfm | anm | lear | leaas | lsigns | members | ca
	DownloadAssetFile Action = superuser | pfm | anm | lear | lsigns | leaas | members | pd | ca
	ValidateAsset     Action = superuser | pfm | anm | ca

	// project actions
	CreateProject       Action = logged
	ListProjects        Action = logged
	GetProject          Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | lsigns | leaas | members | investor | fm | ca
	UpdateProject       Action = superuser | pfm | anm | pm | pd | fm | ca
	UploadProject       Action = superuser | pfm | anm | pm | tama | fm | ca
	DeleteProjectFile   Action = superuser | pfm | anm | pm | ca | fm
	DownloadProjectFile Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | investor | fm | ca
	AddProjectRole      Action = superuser | pfm | anm | pm | pd | ca
	RemoveProjectRole   Action = superuser | pfm | anm | pm | pd | ca
	ChangeFundManager   Action = superuser | pfm | anm | pd | ca
	AssignPM            Action = superuser | pfm | anm | pm | plsign | ca
	CommentProject      Action = superuser | pfm | anm | pm | paco | plsign | tama | fm | ca

	// contract actions
	DownloadProjectContract     Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | fm | ca
	DownloadProjectAgreement    Action = superuser | pfm | anm | pm | paco | plsign | teme | pd | lear | fm | ca
	UpdateProjectAgreement      Action = superuser | pfm | anm | pm | paco | plsign | fm | ca
	GetProjectAgreement         Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | fm | ca
	GetProjectContractFields    Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | fm | ca
	UpdateProjectContractFields Action = superuser | pfm | anm | pm | paco | plsign | fm | ca
	GetProjectContractTable     Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | fm | ca
	UpdateProjectContractTable  Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	UpdateProjectMaintenance    Action = superuser | pfm | anm | pm | paco | plsign | fm | ca
	GetProjectMaintenance       Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | fm | ca

	// indoor clima actions
	GetProjectIndoorClima    Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | ca
	UpdateProjectIndoorClima Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | ca

	// milestones
	AdvanceProjectToWorkPhase       Action = superuser | pfm | anm | pd | pm | ca
	AdvanceProjectToMilestone       Action = superuser | pfm | anm | pm | ca
	GetWorkPhase                    Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	UploadWorkPhase                 Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	DownloadWorkPhaseFile           Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	DeleteWorkPhaseFile             Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	AdvanceProjectToMonitoringPhase Action = superuser | pfm | anm | pd | pm | ca
	GetMonitoringPhase              Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	UploadMonitoringPhase           Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	DownloadMonitoringPhaseFile     Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	DeleteMonitoringPhaseFile       Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | fm | ca
	CreateTask                      Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | lsigns | leaas | members | fm | ca
	GetTask                         Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | lsigns | leaas | members | fm | ca
	DeleteTask                      Action = superuser | pfm | anm | pm | paco | plsign | fm | ca
	CommentTask                     Action = superuser | pfm | anm | pm | pd | fm | ca
	UpdateTask                      Action = superuser | pfm | anm | pm | paco | plsign | tama | teme | pd | lear | lsigns | leaas | members | fm | ca
	WPReview                        Action = superuser | pfm | anm | pd | pm | fm | ca
	MPReview                        Action = superuser | pfm | anm | pd | pm | fm | ca
	WPReviewMaintenance             Action = superuser | pfm | anm | pd | pm | tama | fm | ca

	// fa
	CreateFA           Action = superuser | pfm | anm | pm | ca
	ReviewFA           Action = superuser | pfm | anm | pd | fm | pm | ca
	ListFAByCountry    Action = superuser | pfm | anm | pd | fm | dpo | ca | investor
	GetFAByProject     Action = ListFAByCountry | pm | investor | fm | ca
	GetFA              Action = ListFAByCountry | pm | investor | fm | ca
	UpdateFA           Action = superuser | pfm | anm | pm | fm | ca
	UploadFAAttachment Action = superuser | pfm | anm | pm | tama | fm | ca
	GetFAAttachment    Action = UploadFAAttachment
	DeleteFAAttachment Action = UploadFAAttachment
	CreateFP           Action = CreateFA
	GetFP              Action = GetFA
	UpdateFP           Action = UpdateFA

	// gdpr
	GetGDPRRequest   Action = superuser | pfm | anm | dpo
	ListGDPRRequests Action = superuser | pfm | anm | dpo

	// portfolio
	AddPortfolioRole      Action = superuser | pfm | anm
	RemovePortfolioRole   Action = superuser | pfm | anm
	AddAdminNetworkMan    Action = superuser | pfm | anm
	RemoveAdminNetworkMan Action = superuser | pfm | anm
	AddCountryAdmin       Action = superuser | pfm | anm
	RemoveCountryAdmin    Action = superuser | pfm | anm

	// global
	addEurobor Action = superuser | anm | pfm
	SetVat     Action = superuser | anm | pfm
	GetCountry Action = superuser | anm | pfm
)

func roleAction(u models.User, target uuid.UUID, country models.Country) Action {
	var a Action = logged
	if u.ID == target {
		a |= self
	}
	if u.SuperUser {
		a |= superuser
	}
	if u.PlatformManager {
		a |= pfm
	}
	if u.AdminNwManager {
		a |= anm
	}
	if u.Valid == models.ValidationStatusValid {
		a |= valid
	}
	for _, role := range u.ProjectRoles {
		if role.ProjectID == target {
			a |= Action(roleBit(role.Position))
		}
	}

	for _, role := range u.OrganizationRoles {
		if role.OrganizationID == target {
			a |= Action(roleBit(role.Position))
		}
	}

	for _, role := range u.CountryRoles {
		if role.Country == country && role.UserID == u.ID {
			a |= Action(roleBit(role.Role.String()))
		}
	}

	return a
}

// Can reports whether actor holding a ctx is authorized to perform action on target.
func Can(ctx context.Context, action Action, target uuid.UUID, country models.Country) bool {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return false
	}

	u := cv.User
	actor := roleAction(*u, target, country)

	return action&actor != 0
}

func roleBit(position string) int32 {
	switch position {
	case "lear":
		return lear
	case "lsign":
		return lsigns
	case "leaa":
		return leaas
	case "member":
		return members
	case "pm":
		return pm
	case "paco":
		return paco
	case "plsign":
		return plsign
	case "tama":
		return tama
	case "teme":
		return teme
	case "portfolio_director":
		return pd
	case "fund_manager":
		return fm
	case "data_protection_officer":
		return dpo
	case "country_admin":
		return ca
	case "investor":
		return investor
	default:
		return 0
	}
}
