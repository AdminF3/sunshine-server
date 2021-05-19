package models

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type (
	LegalForm                    int8
	StakeholdersType             int8
	ProjectStatus                int8
	Building                     int8
	Heating                      int8
	ValidationStatus             int8
	EntityType                   string
	UserAction                   string
	GDPRType                     string
	PortfolioRole                string
	ProjectCreationRequestStatus string
	MeetingType                  string
	AssetCategory                string
)

func (f LegalForm) String() string {
	switch f {
	case LegalFormFinancialInstitution:
		return "FinancialInstitution"
	case LegalFormServiceCompany:
		return "ServiceCompany"
	case LegalFormPublicOrganization:
		return "PublicOrganization"
	case LegalFormResidentsCommunity:
		return "ResidentsCommunity"
	case LegalFormNGO:
		return "NGO"
	case LegalFormNaturalPerson:
		return "NaturalPerson"
	case LegalFormSuplier:
		return "Suplier"
	case LegalFormHousingAssociation:
		return "HousingAssociation"
	default:
		return "invalid legal form"
	}
}

func (t StakeholdersType) String() string {
	switch t {

	case StakeHoldersTypeFinancialInstitution:
		return "FinancialInstitution"
	case StakeHoldersTypeServiceCompany:
		return "ServiceCompany"
	case StakeHoldersTypeCentralGovernment:
		return "CentralGovernment"
	case StakeHoldersTypeMunicipality:
		return "Municipality"
	case StakeHoldersTypeAgency:
		return "Agency"
	case StakeHoldersTypeMinistry:
		return "Ministry"
	case StakeHoldersTypeResident:
		return "Resident"
	case StakeHoldersTypeNaturalPerson:
		return "NaturalPerson"
	case StakeHoldersTypeNGO:
		return "NGO"
	case StakeHoldersTypeSupplier:
		return "Supplier"
	case StakeHoldersTypeHousingAssociation:
		return "HousingAssociation"
	case StakeHoldersTypeOther:
		return "Other"

	default:
		return "invalid stakeholders type"
	}
}

func (v ValidationStatus) String() string {
	switch v {
	case ValidationStatusRegistered:
		return "registered"
	case ValidationStatusValid:
		return "valid"
	case ValidationStatusDeclined:
		return "declined"
	case ValidationStatusPending:
		return "pending"
	default:
		return "invalid status"
	}
}

func (p ProjRoles) String() string {
	return fmt.Sprintf("ProjRoles: {PM: %v, PaCo: %v, PLSign: %v, TaMa: %v, TeMe: %v}",
		p.PM, p.PaCo, p.PLSign, p.TaMa, p.TeMe)
}

func (pfr PortfolioRole) String() string {
	switch pfr {
	case PortfolioDirectorRole:
		return "portfolio_director"
	case FundManagerRole:
		return "fund_manager"
	case CountryAdminRole:
		return "country_admin"
	case DataProtectionOfficerRole:
		return "data_protection_officer"
	case InvestorRole:
		return "investor"
	default:
		return fmt.Sprintf("invalid role: %s", string(pfr))
	}
}

func (a AssetCategory) Residential() bool {
	return strings.HasPrefix(string(a), "residential")
}

const (
	NREducationalFacilities    AssetCategory = "nonresidential_educational_facilities"
	NRCulturalFacilities       AssetCategory = "nonresidential_cultural_facilities"
	NRMedicalFacilities        AssetCategory = "nonresidential_medical_facilities"
	NRSportsFacilities         AssetCategory = "nonresidential_sports_facilities"
	NROfficeBuildings          AssetCategory = "nonresidential_office_buildings"
	NRTransportationFacilities AssetCategory = "nonresidential_transportation_facilities"
)

const (
	// Organization related types
	MTypeInternalMeeting   MeetingType = "internal_meeting"
	MTypeConference        MeetingType = "conference"
	MTypeWorkshop          MeetingType = "workshop"
	MTypeEvent             MeetingType = "event"
	MTypeTraining          MeetingType = "training"
	MTypeEUProjectMeeting  MeetingType = "eu_project_meeting"
	MTypeEUProjectActivity MeetingType = "eu_project_activity"

	// Project related types
	MTypeAcquisition                    MeetingType = "acquisition"
	MTypeAcquisitionCommitment          MeetingType = "acquisition_commitment"
	MTypeAcquisitionKickOff             MeetingType = "acquisition_kick_off"
	MTypeWorksKickOff                   MeetingType = "works_kick_off"
	MTypeWorksInitialInformation        MeetingType = "works_initial_information"
	MTypeWorksWeeklyReport              MeetingType = "works_weekly_report"
	MTypeWorksRenovationInformative     MeetingType = "works_renovation_informative"
	MTypeWorksCommunication             MeetingType = "works_communication"
	MTypeWorksConstructionManagersFinal MeetingType = "works_construction_managers_final"
	MTypeWorksFinalInformation          MeetingType = "works_final_information"

	MTypeOther MeetingType = "other"
)

const (
	OpenedStatus   ProjectCreationRequestStatus = "opened"
	AcceptedStatus ProjectCreationRequestStatus = "accepted"
	RejectedStatus ProjectCreationRequestStatus = "rejected"
)

const (
	GDPRTypeGet    GDPRType = "get"
	GDPRTypeDelete GDPRType = "delete"
)

const (
	UserActionCreate                       UserAction = "create"
	UserActionUpdate                       UserAction = "update"
	UserActionUpload                       UserAction = "upload"
	UserActionAssign                       UserAction = "assign"
	UserActionGDPR                         UserAction = "gdpr"
	UserActionRequestMembership            UserAction = "request_membership"
	UserActionLEARApply                    UserAction = "lear_apply"
	UserActionRequestProjectCreation       UserAction = "request_project_creation"
	UserActionClaimResidency               UserAction = "claim_residency"
	UserActionAcceptLEARApplication        UserAction = "accept_lear_application"
	UserActionRemove                       UserAction = "remove"
	UserActionReject                       UserAction = "reject"
	UserActionForfaitingApplication        UserAction = "forfaiting_application"
	UserActionRejectLEARApplication        UserAction = "reject_lear_application"
	UserActionApproveForfaitingApplication UserAction = "approve_forfaiting_application"
	UserActionApproveForfaitingPayment     UserAction = "approve_forfaiting_payment"
)

const (
	PortfolioDirectorRole     PortfolioRole = "portfolio_director"
	FundManagerRole           PortfolioRole = "fund_manager"
	CountryAdminRole          PortfolioRole = "country_admin"
	DataProtectionOfficerRole PortfolioRole = "data_protection_officer"
	InvestorRole              PortfolioRole = "investor"
)

const (
	UserT         EntityType = "user"
	OrganizationT EntityType = "organization"
	AssetT        EntityType = "asset"
	ProjectT      EntityType = "project"
	IndoorClimaT  EntityType = "indoor_clima"
	MeetingT      EntityType = "meeting"
)

const (
	_ ValidationStatus = iota
	ValidationStatusRegistered
	ValidationStatusValid
	ValidationStatusDeclined
	ValidationStatusPending
)

const (
	_ StakeholdersType = iota
	StakeHoldersTypeFinancialInstitution
	StakeHoldersTypeServiceCompany
	StakeHoldersTypeCentralGovernment
	StakeHoldersTypeMunicipality
	StakeHoldersTypeAgency
	StakeHoldersTypeMinistry
	StakeHoldersTypeResident
	StakeHoldersTypeNaturalPerson
	StakeHoldersTypeNGO
	StakeHoldersTypeSupplier
	StakeHoldersTypeHousingAssociation
	StakeHoldersTypeOther
)

const (
	_ LegalForm = iota
	LegalFormFinancialInstitution
	LegalFormServiceCompany
	LegalFormPublicOrganization
	LegalFormResidentsCommunity
	LegalFormNGO
	LegalFormNaturalPerson
	LegalFormSuplier
	LegalFormHousingAssociation
)

const (
	_ ProjectStatus = iota
	ProjectStatusPlanning
	ProjectStatusInProgress
	ProjectStatusFinished
	ProjectStatusAbandoned
)

const (
	_ Building = iota
	BuildingCzechProject
	BuildingType103
	BuildingType104
	BuildingType119
	BuildingType316
	BuildingType318
	BuildingType464
	BuildingType467
	BuildingType602
	BuildingOther
)

const (
	_ Heating = iota
	HeatingDistrict
	HeatingBuilding
)

// appendUnique appends elem to slice only if it's not already there.
func appendUnique(slice []uuid.UUID, elem uuid.UUID) []uuid.UUID {
	for _, v := range slice {
		if elem == v {
			return slice
		}
	}

	return append(slice, elem)
}
