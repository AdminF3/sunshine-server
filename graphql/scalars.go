package graphql

import (
	"fmt"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	switch id := v.(type) {
	case string:
		return uuid.Parse(id)
	case []byte:
		return uuid.ParseBytes(id)
	default:
		return uuid.Nil, fmt.Errorf("%[1]T(%[1]v) is not uuid", v)
	}
}

func MarshalUUID(id uuid.UUID) graphql.Marshaler {
	return graphql.MarshalString(id.String())
}

func UnmarshalEntityType(v interface{}) (models.EntityType, error) {
	et, _ := v.(string)
	switch et {
	case "USER":
		return models.UserT, nil
	case "ORGANIZATION":
		return models.OrganizationT, nil
	case "ASSET":
		return models.AssetT, nil
	case "PROJECT":
		return models.ProjectT, nil
	case "INDOOR_CLIMA":
		return models.IndoorClimaT, nil
	case "MEETING":
		return models.MeetingT, nil
	default:
		return "", fmt.Errorf("%[1]T(%[1]v)  is not valid entity type", v)
	}
}

func MarshalEntityType(et models.EntityType) graphql.Marshaler {
	return graphql.MarshalString(string(et))
}

func UnmarshalUserAction(v interface{}) (models.UserAction, error) {
	ua, _ := v.(string)
	switch ua {
	case "CREATE":
		return models.UserActionCreate, nil
	case "UPDATE":
		return models.UserActionUpdate, nil
	case "UPLOAD":
		return models.UserActionUpload, nil
	case "ASSIGN":
		return models.UserActionAssign, nil
	case "GDPR":
		return models.UserActionGDPR, nil
	case "REQUEST_MEMBERSHIP":
		return models.UserActionRequestMembership, nil
	case "LEAR_APPLY":
		return models.UserActionLEARApply, nil
	case "REQUEST_PROJECT_CREATION":
		return models.UserActionRequestProjectCreation, nil
	case "CLAIM_RESIDENCY":
		return models.UserActionClaimResidency, nil
	case "ACCEPT_LEAR_APPLICATION":
		return models.UserActionAcceptLEARApplication, nil
	case "REMOVE":
		return models.UserActionRemove, nil
	case "REJECT":
		return models.UserActionReject, nil
	case "FORFAITING_APPLICATION":
		return models.UserActionForfaitingApplication, nil
	case "REJECT_LEAR_APPLICATION":
		return models.UserActionRejectLEARApplication, nil
	case "APPROVE_FORFAITING_APPLICATION":
		return models.UserActionApproveForfaitingApplication, nil
	case "APPROVE_FORFAITING_PAYMENT":
		return models.UserActionApproveForfaitingPayment, nil
	default:
		return "", fmt.Errorf("%[1]T(%[1]v) is not user action", v)
	}
}

func MarshalUserAction(ua models.UserAction) graphql.Marshaler {
	return graphql.MarshalString(string(ua))
}

func UnmarshalQuality(v interface{}) (contract.Quality, error) {
	quality, _ := v.(string)
	switch quality {
	case "GOOD":
		return contract.QualityGood, nil
	case "POOR":
		return contract.QualityPoor, nil
	case "NO":
		return contract.QualityNo, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not quality", v)
	}
}

func UnmarshalFinance(v interface{}) (models.Finance, error) {
	f, _ := v.(string)
	switch f {
	case "EQUITY":
		return models.FinanceEquity, nil
	case "BANK_FUNDING":
		return models.FinanceBankFunding, nil
	case "OTHER":
		return models.FinanceOther, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not string", v)
	}
}

func UnmarshalFAReviewType(v interface{}) (models.FAReviewType, error) {
	t, _ := v.(string)
	switch t {
	case "FINANCIAL":
		return models.FAReviewTypeFinancial, nil
	case "TECHNICAL":
		return models.FAReviewTypeTechnical, nil
	case "GUIDELINES":
		return models.FAReviewTypeGuidelines, nil
	case "EXECUTIVE":
		return models.FAReviewTypeExecutive, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not string", v)
	}
}

func MarshalCurrency(c models.Currency) graphql.Marshaler {
	return graphql.MarshalString(string(c))
}

func UnmarshalCurrency(v interface{}) (models.Currency, error) {
	t, _ := v.(string)
	switch t {
	case "EUR":
		return models.CurrencyEUR, nil
	case "ALL":
		return models.CurrencyALL, nil
	case "AMD":
		return models.CurrencyAMD, nil
	case "BYN":
		return models.CurrencyBYN, nil
	case "BAM":
		return models.CurrencyBAM, nil
	case "BGN":
		return models.CurrencyBGN, nil
	case "HRK":
		return models.CurrencyHRK, nil
	case "CZK":
		return models.CurrencyCZK, nil
	case "DKK":
		return models.CurrencyDKK, nil
	case "GEL":
		return models.CurrencyGEL, nil
	case "HUF":
		return models.CurrencyHUF, nil
	case "ISK":
		return models.CurrencyISK, nil
	case "CHF":
		return models.CurrencyCHF, nil
	case "MDL":
		return models.CurrencyMDL, nil
	case "MKD":
		return models.CurrencyMKD, nil
	case "NOK":
		return models.CurrencyNOK, nil
	case "PLN":
		return models.CurrencyPLN, nil
	case "RON":
		return models.CurrencyRON, nil
	case "RUB":
		return models.CurrencyRUB, nil
	case "RSD":
		return models.CurrencyRSD, nil
	case "SEK":
		return models.CurrencySEK, nil
	case "TRY":
		return models.CurrencyTRY, nil
	case "UAH":
		return models.CurrencyUAH, nil
	case "GBP":
		return models.CurrencyGBP, nil
	default:
		return "", fmt.Errorf("%[1]T(%[1]v) is not string", v)
	}
}

func UnmarshalWPReviewType(v interface{}) (models.WPReviewType, error) {
	t, _ := v.(string)
	switch t {
	case "FINANCIAL":
		return models.WPReviewTypeFinancial, nil
	case "TECHNICAL":
		return models.WPReviewTypeTechnical, nil
	case "BANK_ACCOUNT":
		return models.WPReviewTypeBankAccount, nil
	case "EXECUTIVE":
		return models.WPReviewTypeExecutive, nil
	case "MAINTENANCE":
		return models.WPReviewTypeMaintenance, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not string", v)
	}
}

func UnmarshalMPReviewType(v interface{}) (models.MPReviewType, error) {
	t, _ := v.(string)
	switch t {
	case "FORFAITING":
		return models.MPReviewTypeForfaiting, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not string", v)
	}
}

func UnmarshalColumnKind(v interface{}) (contract.Kind, error) {
	kind, _ := v.(string)

	switch kind {
	case "String":
		return contract.String, nil
	case "Name":
		return contract.Name, nil
	case "Decimal":
		return contract.Decimal, nil
	case "Money":
		return contract.Money, nil
	case "Temperature":
		return contract.Temperature, nil
	case "Count":
		return contract.Count, nil
	case "Volume":
		return contract.Volume, nil
	case "Energy":
		return contract.Energy, nil
	case "Area":
		return contract.Area, nil
	case "Percent":
		return contract.Percent, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not column", v)
	}
}

// MarshalQuality returns marshaled string for quality enum
func MarshalQuality(q contract.Quality) graphql.Marshaler {
	return graphql.MarshalString(qualityMap[q])
}

func MarshalFinance(f models.Finance) graphql.Marshaler {
	return graphql.MarshalString(financeMap[f])
}

func MarshalFAReviewType(f models.FAReviewType) graphql.Marshaler {
	return graphql.MarshalString(faReviewTypeMap[f])
}

func MarshalWPReviewType(f models.WPReviewType) graphql.Marshaler {
	return graphql.MarshalString(wpReviewTypeMap[f])
}

func MarshalMPReviewType(f models.MPReviewType) graphql.Marshaler {
	return graphql.MarshalString(mpReviewTypeMap[f])
}

func MarshalColumnKind(k contract.Kind) graphql.Marshaler {
	return graphql.MarshalString(kindMap[k])
}

func MarshalValidationStatus(v models.ValidationStatus) graphql.Marshaler {
	return graphql.MarshalString(validationStatusMap[v])
}

func MarshalBuildingType(v models.Building) graphql.Marshaler {
	return graphql.MarshalString(buildingTypeMap[v])
}

func MarshalHeatingType(v models.Heating) graphql.Marshaler {
	return graphql.MarshalString(heatingTypeMap[v])
}

func MarshalProjectStatus(v models.ProjectStatus) graphql.Marshaler {
	return graphql.MarshalString(projectStatusMap[v])
}

func MarshalMeetingType(mt models.MeetingType) graphql.Marshaler {
	return graphql.MarshalString(meetingTopicMap[mt])
}

func UnmarshalMeetingType(v interface{}) (models.MeetingType, error) {
	mt, _ := v.(string)
	switch mt {
	case "TRAINING":
		return models.MTypeTraining, nil
	case "EU_PROJECT_ACTIVITY":
		return models.MTypeEUProjectActivity, nil
	case "ACQUISITION":
		return models.MTypeAcquisition, nil
	case "ACQUISITION_COMMITMENT":
		return models.MTypeAcquisitionCommitment, nil
	case "ACQUISITION_KICK_OFF":
		return models.MTypeAcquisitionKickOff, nil
	case "WORKS_KICK_OFF":
		return models.MTypeWorksKickOff, nil
	case "WORKS_INITIAL_INFORMATION":
		return models.MTypeWorksInitialInformation, nil
	case "WORKS_WEEKLY_REPORT":
		return models.MTypeWorksWeeklyReport, nil
	case "WORKS_RENOVATION_INFORMATIVE":
		return models.MTypeWorksRenovationInformative, nil
	case "WORKS_COMMUNICATION":
		return models.MTypeWorksCommunication, nil
	case "WORKS_CONSTRUCTION_MANAGERS_FINAL":
		return models.MTypeWorksConstructionManagersFinal, nil
	case "WORKS_FINAL_INFORMATION":
		return models.MTypeWorksFinalInformation, nil
	case "INTERNAL_MEETING":
		return models.MTypeInternalMeeting, nil
	case "CONFERENCE":
		return models.MTypeConference, nil
	case "WORKSHOP":
		return models.MTypeWorkshop, nil
	case "EVENT":
		return models.MTypeEvent, nil
	case "EU_PROJECT_MEETING":
		return models.MTypeEUProjectMeeting, nil
	case "OTHER":
		return models.MTypeOther, nil
	default:
		return "", fmt.Errorf("%[1]T(%[1]v) is not meeting type", v)
	}
}

func UnmarshalValidationStatus(v interface{}) (models.ValidationStatus, error) {
	vs, _ := v.(string)
	switch vs {
	case "REGISTERED":
		return models.ValidationStatusRegistered, nil
	case "VALID":
		return models.ValidationStatusValid, nil
	case "DECLINED":
		return models.ValidationStatusDeclined, nil
	case "PENDING":
		return models.ValidationStatusPending, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not validation status", v)
	}
}

func UnmarshalBuildingType(v interface{}) (models.Building, error) {
	t, _ := v.(string)
	switch t {
	case "CZECH_PROJECT":
		return models.BuildingCzechProject, nil
	case "TYPE_103":
		return models.BuildingType103, nil
	case "TYPE_104":
		return models.BuildingType104, nil
	case "TYPE_119":
		return models.BuildingType119, nil
	case "TYPE_316":
		return models.BuildingType316, nil
	case "TYPE_318":
		return models.BuildingType318, nil
	case "TYPE_464":
		return models.BuildingType464, nil
	case "TYPE_467":
		return models.BuildingType467, nil
	case "TYPE_602":
		return models.BuildingType602, nil
	case "OTHER":
		return models.BuildingOther, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not building type", v)
	}
}

func UnmarshalHeatingType(v interface{}) (models.Heating, error) {
	t, _ := v.(string)
	switch t {
	case "DISTRICT":
		return models.HeatingDistrict, nil
	case "BUILDING":
		return models.HeatingBuilding, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not heating type", v)
	}
}

func UnmarshalProjectStatus(v interface{}) (models.ProjectStatus, error) {
	t, _ := v.(string)
	switch t {
	case "PLANNING":
		return models.ProjectStatusPlanning, nil
	case "IN_PROGRESS":
		return models.ProjectStatusInProgress, nil
	case "FINISHED":
		return models.ProjectStatusFinished, nil
	case "ABANDONED":
		return models.ProjectStatusAbandoned, nil
	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not project status", v)
	}
}

func UnmarshalLegalForm(v interface{}) (models.LegalForm, error) {
	form, _ := v.(string)
	switch form {
	case "FINANCIAL_INSTITUTION":
		return models.LegalFormFinancialInstitution, nil
	case "SERVICE_COMPANY":
		return models.LegalFormServiceCompany, nil
	case "PUBLIC_ORGANIZATION":
		return models.LegalFormPublicOrganization, nil
	case "RESIDENTS_COMMUNITY":
		return models.LegalFormResidentsCommunity, nil
	case "NGO":
		return models.LegalFormNGO, nil
	case "NATURAL_PERSON":
		return models.LegalFormNaturalPerson, nil
	case "SUPLIER":
		return models.LegalFormSuplier, nil
	case "HOUSING_ASSOCIATION":
		return models.LegalFormHousingAssociation, nil

	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not legal form", v)
	}
}

// MarshalLegalForm returns the marshaled string representation of leagalform enum
func MarshalLegalForm(f models.LegalForm) graphql.Marshaler {
	return graphql.MarshalString(legalFormMap[f])
}

// UnmarshalStakeholdersType properly to the necessary StakeholdersType.
// Returns error otherwise.
func UnmarshalStakeholdersType(v interface{}) (models.StakeholdersType, error) {
	form, _ := v.(string)
	switch form {
	case "FINANCIAL_INSTITUTION":
		return models.StakeHoldersTypeFinancialInstitution, nil
	case "SERVICE_COMPANY":
		return models.StakeHoldersTypeServiceCompany, nil
	case "CENTRAL_GOVERNMENT":
		return models.StakeHoldersTypeCentralGovernment, nil
	case "MUNICIPALITY":
		return models.StakeHoldersTypeMunicipality, nil
	case "AGENCY":
		return models.StakeHoldersTypeAgency, nil
	case "MINISTRY":
		return models.StakeHoldersTypeMinistry, nil
	case "RESIDENT":
		return models.StakeHoldersTypeResident, nil
	case "NATURAL_PERSON":
		return models.StakeHoldersTypeNaturalPerson, nil
	case "NGO":
		return models.StakeHoldersTypeNGO, nil
	case "SUPPLIER":
		return models.StakeHoldersTypeSupplier, nil
	case "HOUSING_ASSOCIATION":
		return models.StakeHoldersTypeHousingAssociation, nil
	case "OTHER":
		return models.StakeHoldersTypeOther, nil

	default:
		return 0, fmt.Errorf("%[1]T(%[1]v) is not stakeholders type", v)
	}
}

// MarshalStakeholdersType returns the marshaled string representation of StakeholdersType enum
func MarshalStakeholdersType(t models.StakeholdersType) graphql.Marshaler {
	return graphql.MarshalString(meetingStakeholdersMap[t])
}

func MarshalMilestone(m models.Milestone) graphql.Marshaler {
	return graphql.MarshalString(milestoneMap[m])
}

func UnmarshalMilestone(v interface{}) (models.Milestone, error) {
	milestone, _ := v.(string)

	switch milestone {
	case "ZERO":
		return models.MilestoneZero, nil
	case "ACQUISITION_MEETING":
		return models.MilestoneAcquisitionMeeting, nil
	case "FEASIBILITY_STUDY":
		return models.MilestoneFeasibilityStudy, nil
	case "COMMITMENT_STUDY":
		return models.MilestoneCommitmentStudy, nil
	case "PROJECT_DESIGN":
		return models.MilestoneProjectDesign, nil
	case "PROJECT_PREPARATION":
		return models.MilestoneProjectPreparation, nil
	case "KICK_OFF_MEETING":
		return models.MilestoneKickOffMeeting, nil
	case "WORK_PHASE":
		return models.MilestoneWorkPhase, nil
	case "MONITORING_PHASE":
		return models.MilestoneMonitoringPhase, nil
	case "COMMISSIONING":
		return models.MilestoneCommissioning, nil
	case "FORFAITING_PAYMENT":
		return models.MilestoneForfaitingPayment, nil
	default:
		return "", fmt.Errorf("%[1]T(%[1]v) is not milestone", v)
	}
}

func MarshalProjectRole(pr models.ProjectRole) graphql.Marshaler {
	return graphql.MarshalString(pr.Position)
}

func UnmarshalProjectRole(v interface{}) (models.ProjectRole, error) {
	pr, _ := v.(string)
	switch pr {
	case "PM":
		return models.ProjectRole{Position: "pm"}, nil
	case "PLSIGN":
		return models.ProjectRole{Position: "plsign"}, nil
	case "PACO":
		return models.ProjectRole{Position: "paco"}, nil
	case "TAMA":
		return models.ProjectRole{Position: "tama"}, nil
	case "TEME":
		return models.ProjectRole{Position: "teme"}, nil
	default:
		return models.ProjectRole{}, fmt.Errorf("%[1]T(%[1]v) is not project role", v)
	}
}

func MarshalOrganizationRole(or models.OrganizationRole) graphql.Marshaler {
	return graphql.MarshalString(or.Position)
}

func UnmarshalOrganizationRole(v interface{}) (models.OrganizationRole, error) {
	or, _ := v.(string)
	switch or {
	case "LEAR":
		return models.OrganizationRole{Position: "lear"}, nil
	case "LSIGNS":
		return models.OrganizationRole{Position: "lsign"}, nil
	case "LEAAS":
		return models.OrganizationRole{Position: "leaa"}, nil
	case "MEMBERS":
		return models.OrganizationRole{Position: "member"}, nil
	default:
		return models.OrganizationRole{}, fmt.Errorf("%[1]T(%[1]v) is not organization role", v)
	}
}

func MarshalGDPRType(a models.GDPRType) graphql.Marshaler {
	return graphql.MarshalString(string(a))
}

func UnmarshalGDPRType(v interface{}) (models.GDPRType, error) {
	a, _ := v.(string)
	switch a {
	case "GET":
		return models.GDPRTypeGet, nil
	case "DELETE":
		return models.GDPRTypeDelete, nil
	default:
		return "", fmt.Errorf("%[1]T(%[1]v) is not gdpr action", v)
	}
}

func MarshalPortfolioRole(pr models.PortfolioRole) graphql.Marshaler {
	return graphql.MarshalString(string(pr))
}

func UnmarshalPortfolioRole(v interface{}) (models.PortfolioRole, error) {
	pr, _ := v.(string)
	switch pr {
	case "PD":
		return models.PortfolioDirectorRole, nil
	case "DPO":
		return models.DataProtectionOfficerRole, nil
	case "CA":
		return models.CountryAdminRole, nil
	case "FM":
		return models.FundManagerRole, nil
	case "INVESTOR":
		return models.InvestorRole, nil
	default:
		return "", fmt.Errorf("%[1]T(%[1]v) is not portfolio role", v)
	}
}

var (
	msgOK      = &Message{Message: "Success!"}
	msgErr     = &Message{Message: "Error!"}
	qualityMap = map[contract.Quality]string{
		1: "GOOD",
		2: "POOR",
		3: "NO",
	}
	kindMap = map[contract.Kind]string{
		0: "String",
		1: "Name",
		2: "Decimal",
		3: "Money",
		4: "Temperature",
		5: "Count",
		6: "Volume",
		7: "Energy",
		8: "Area",
		9: "Percent",
	}
	validationStatusMap = map[models.ValidationStatus]string{
		models.ValidationStatusRegistered: "REGISTERED",
		models.ValidationStatusValid:      "VALID",
		models.ValidationStatusDeclined:   "DECLINED",
		models.ValidationStatusPending:    "PENDING",
	}
	legalFormMap = map[models.LegalForm]string{
		models.LegalFormFinancialInstitution: "FINANCIAL_INSTITUTION",
		models.LegalFormServiceCompany:       "SERVICE_COMPANY",
		models.LegalFormPublicOrganization:   "PUBLIC_ORGANIZATION",
		models.LegalFormResidentsCommunity:   "RESIDENTS_COMMUNITY",
		models.LegalFormNGO:                  "NGO",
		models.LegalFormNaturalPerson:        "NATURAL_PERSON",
		models.LegalFormSuplier:              "SUPLIER",
		models.LegalFormHousingAssociation:   "HOUSING_ASSOCIATION",
	}

	meetingStakeholdersMap = map[models.StakeholdersType]string{
		models.StakeHoldersTypeFinancialInstitution: "FINANCIAL_INSTITUTION",
		models.StakeHoldersTypeServiceCompany:       "SERVICE_COMPANY",
		models.StakeHoldersTypeCentralGovernment:    "CENTRAL_GOVERNMENT",
		models.StakeHoldersTypeMunicipality:         "MUNICIPALITY",
		models.StakeHoldersTypeAgency:               "AGENCY",
		models.StakeHoldersTypeMinistry:             "MINISTRY",
		models.StakeHoldersTypeResident:             "RESIDENT",
		models.StakeHoldersTypeNaturalPerson:        "NATURAL_PERSON",
		models.StakeHoldersTypeNGO:                  "NGO",
		models.StakeHoldersTypeSupplier:             "SUPPLIER",
		models.StakeHoldersTypeHousingAssociation:   "HOUSING_ASSOCIATION",
		models.StakeHoldersTypeOther:                "OTHER",
	}

	milestoneMap = map[models.Milestone]string{
		models.MilestoneZero:               "ZERO",
		models.MilestoneAcquisitionMeeting: "ACQUISITION_MEETING",
		models.MilestoneFeasibilityStudy:   "FEASIBILITY_STUDY",
		models.MilestoneCommitmentStudy:    "COMMITMENT_STUDY",
		models.MilestoneProjectDesign:      "PROJECT_DESIGN",
		models.MilestoneProjectPreparation: "PROJECT_PREPARATION",
		models.MilestoneKickOffMeeting:     "KICK_OFF_MEETING",
		models.MilestoneWorkPhase:          "WORK_PHASE",
		models.MilestoneMonitoringPhase:    "MONITORING_PHASE",
		models.MilestoneCommissioning:      "COMMISSIONING",
		models.MilestoneForfaitingPayment:  "FORFAITING_PAYMENT",
	}

	meetingTopicMap = map[models.MeetingType]string{
		models.MTypeInternalMeeting:   "INTERNAL_MEETING",
		models.MTypeConference:        "CONFERENCE",
		models.MTypeWorkshop:          "WORKSHOP",
		models.MTypeEvent:             "EVENT",
		models.MTypeTraining:          "TRAINING",
		models.MTypeEUProjectMeeting:  "EU_PROJECT_MEETING",
		models.MTypeEUProjectActivity: "EU_PROJECT_ACTIVITY",

		models.MTypeAcquisition:                    "ACQUISITION",
		models.MTypeAcquisitionCommitment:          "ACQUISITION_COMMITMENT",
		models.MTypeAcquisitionKickOff:             "ACQUISITION_KICK_OFF",
		models.MTypeWorksKickOff:                   "WORKS_KICK_OFF",
		models.MTypeWorksInitialInformation:        "WORKS_INITIAL_INFORMATION",
		models.MTypeWorksWeeklyReport:              "WORKS_WEEKLY_REPORT",
		models.MTypeWorksRenovationInformative:     "WORKS_RENOVATION_INFORMATIVE",
		models.MTypeWorksCommunication:             "WORKS_COMMUNICATION",
		models.MTypeWorksConstructionManagersFinal: "WORKS_CONSTRUCTION_MANAGERS_FINAL",
		models.MTypeWorksFinalInformation:          "WORKS_FINAL_INFORMATION",

		models.MTypeOther: "OTHER",
	}

	buildingTypeMap = map[models.Building]string{
		models.BuildingCzechProject: "CZECH_PROJECT",
		models.BuildingType103:      "TYPE_103",
		models.BuildingType104:      "TYPE_104",
		models.BuildingType119:      "TYPE_119",
		models.BuildingType316:      "TYPE_316",
		models.BuildingType318:      "TYPE_318",
		models.BuildingType464:      "TYPE_464",
		models.BuildingType467:      "TYPE_467",
		models.BuildingType602:      "TYPE_602",
		models.BuildingOther:        "OTHER",
	}

	heatingTypeMap = map[models.Heating]string{
		models.HeatingBuilding: "BUILDING",
		models.HeatingDistrict: "DISTRICT",
	}

	projectStatusMap = map[models.ProjectStatus]string{
		models.ProjectStatusPlanning:   "PLANNING",
		models.ProjectStatusInProgress: "IN_PROGRESS",
		models.ProjectStatusFinished:   "FINISHED",
		models.ProjectStatusAbandoned:  "ABANDONED",
	}

	financeMap = map[models.Finance]string{
		1: "EQUITY",
		2: "BANK_FUNDING",
		3: "OTHER",
	}

	faReviewTypeMap = map[models.FAReviewType]string{
		1: "FINANCIAL",
		2: "TECHNICAL",
		3: "GUIDELINES",
		4: "EXECUTIVE",
	}

	wpReviewTypeMap = map[models.WPReviewType]string{
		1: "FINANCIAL",
		2: "TECHNICAL",
		3: "BANK_ACCOUNT",
		4: "EXECUTIVE",
		5: "MAINTENANCE",
	}

	mpReviewTypeMap = map[models.MPReviewType]string{
		1: "FORFAITING",
	}
)
