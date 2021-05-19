package controller

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Organization struct {
	store      stores.Store
	notifier   stores.Notifier
	pf         stores.Portfolio
	uploadPath string
}

func NewOrganization(env *services.Env) *Organization {
	return &Organization{
		store:      env.OrganizationStore,
		notifier:   env.Notifier,
		pf:         env.Portfolio,
		uploadPath: env.Paths.Uploads,
	}
}

func (o *Organization) Create(ctx context.Context, r io.Reader) (*models.Document, error) {
	var org models.Organization

	if err := json.NewDecoder(r).Decode(&org); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	if org.LegalForm != models.LegalFormResidentsCommunity {
		// only residents community can be with null values for these fields

		if org.Email == "" ||
			org.Website == "" ||
			time.Time.IsZero(org.Registered) ||
			org.VAT == "" {
			return nil, fmt.Errorf("%w: %v", ErrBadInput, "Missing mandatory fields")
		}

	}

	if !Can(ctx, CreateOrganization, org.ID, org.Country) {
		return nil, ErrUnauthorized
	}

	org.Valid = models.ValidationStatusRegistered
	org.OrganizationRoles = append(org.OrganizationRoles, models.OrganizationRole{
		Position: "lear",
		UserID:   org.Roles.LEAR,
	})

	doc, err := o.store.Create(ctx, &org)
	if err != nil {
		return nil, err
	}
	cv := services.FromContext(ctx)
	if cv.User != nil {
		admins, err := o.pf.GetPortfolioRolesPerCountry(ctx, cv.User.Country, models.CountryAdminRole)
		if err == nil {
			n := models.Notification{
				Action:     models.UserActionCreate,
				UserID:     cv.User.ID,
				UserKey:    cv.User.Name,
				TargetID:   org.ID,
				TargetType: models.OrganizationT,
				TargetKey:  org.Name,
				Country:    org.Country,
			}
			if !org.Country.IsConsortium() {
				admins = []uuid.UUID{getANWManager(o.store)}
			}
			notifyAll(ctx, o.notifier, admins, n)
		} else {
			sentry.Report(err)
		}
	}
	return doc, nil
}

func (o *Organization) Get(ctx context.Context, id uuid.UUID) (*models.Document, stores.Dependencies, error) {
	cv := services.FromContext(ctx)
	if !Can(ctx, GetOrganization, id, cv.User.Country) {
		return nil, nil, ErrUnauthorized
	}

	return o.store.Unwrap(ctx, id)
}

func (o *Organization) Update(ctx context.Context, id uuid.UUID, r io.Reader) (*models.Document, stores.Dependencies, error) {
	doc, err := o.store.Get(ctx, id)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old, _ := o.store.Get(ctx, id)

	org := doc.Data.(*models.Organization)
	oldValid := org.Valid
	oldFields := org.ExcerptString()
	docVal := org.Value

	if !Can(ctx, UpdateOrganization, id, org.Country) {
		return nil, nil, ErrUnauthorized
	}

	if err := json.NewDecoder(r).Decode(&org); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	if org.LegalForm != models.LegalFormResidentsCommunity {
		//only residents community can be with null values for these fields

		if org.Email == "" ||
			org.Website == "" ||
			time.Time.IsZero(org.Registered) ||
			org.VAT == "" {
			return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, "Missing mandatory fields")
		}

	}

	updated := models.NewDocument(org)
	if stores.ShouldInvalidate(old, updated, services.FromContext(ctx).User.SuperUser, oldValid) {
		updated.Data.(*models.Organization).Valid = models.ValidationStatusPending

		newFields := updated.Data.(*models.Organization).ExcerptString()
		cv := services.FromContext(ctx)

		admins, err := o.pf.GetPortfolioRolesPerCountry(ctx, org.Country, models.CountryAdminRole)
		if err == nil {
			n := models.Notification{
				Action:     models.UserActionUpdate,
				UserID:     cv.User.ID,
				UserKey:    cv.User.Name,
				TargetID:   org.ID,
				TargetType: models.OrganizationT,
				TargetKey:  org.Name,
				Old:        oldFields,
				New:        newFields,
				Country:    org.Country,
			}
			notifyAll(ctx, o.notifier, admins, n)
		} else {
			sentry.Report(err)
		}
	}

	doc.Data = updated.Data
	doc.Data.(*models.Organization).Value = docVal

	_, err = o.store.Update(ctx, doc)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	return o.store.Unwrap(ctx, id)
}

func (o *Organization) List(ctx context.Context, f stores.Filter, id uuid.UUID) ([]models.Document, stores.Dependencies, int, error) {
	if id != uuid.Nil {
		return o.store.ListByMember(ctx, f, id)
	}

	return o.store.List(ctx, f)
}

func (o *Organization) UploadFile(ctx context.Context, oid uuid.UUID, form RequestForm) error {
	cv := services.FromContext(ctx)
	if !Can(ctx, UploadOrganization, oid, cv.User.Country) {
		return ErrUnauthorized
	}

	doc, err := o.store.Get(ctx, oid)
	if err != nil {
		return err
	}

	return uploadFile(ctx, o.store, o.notifier, form, doc, o.uploadPath)
}

func (o *Organization) GetFile(ctx context.Context, oid uuid.UUID, filename string) (*models.Attachment, *os.File, error) {
	doc, err := o.store.Get(ctx, oid)
	if err != nil {
		return nil, nil, err
	}

	org := doc.Data.(*models.Organization)

	// Everybody should access organization's logos.
	if filename != filepath.Base(org.Logo) && !Can(ctx, DownloadOrgFile, oid, org.Country) {
		return nil, nil, ErrUnauthorized
	}

	return getFile(ctx, o.store, oid, filename, o.uploadPath)
}

func (o *Organization) DeleteFile(ctx context.Context, oid uuid.UUID, filename string) error {
	cv := services.FromContext(ctx)
	if !Can(ctx, DeleteOrgFile, oid, cv.User.Country) {
		return ErrUnauthorized
	}

	doc, err := o.store.Get(ctx, oid)
	if err != nil {
		return err
	}

	return o.store.DeleteAttachment(ctx, doc, filename)
}

func (o *Organization) AddRole(ctx context.Context, id uuid.UUID, body io.Reader) (*models.Document, stores.Dependencies, error) {
	doc, err := o.store.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	org := doc.Data.(*models.Organization)

	if !Can(ctx, AddOrganizationRole, id, org.Country) {
		return nil, nil, ErrUnauthorized
	}

	var ar alterRole
	if err := json.NewDecoder(body).Decode(&ar); err != nil {
		return nil, nil, err
	}

	roles := org.OrganizationRoles
	switch ar.Position {
	case "lear":
		for i, role := range roles {
			if role.Position == "lear" {
				roles[i].UserID = ar.User
			}
		}
	case "leaa", "lsign", "member":
		roles = append(roles, models.OrganizationRole{
			UserID:         ar.User,
			Position:       ar.Position,
			OrganizationID: org.ID})
	default:
		return nil, nil, fmt.Errorf("%w: non existing position", ErrBadInput)
	}

	org.OrganizationRoles = setUniqueOrgRoles(roles)
	_, err = o.store.Update(ctx, sanitizeInputFields(doc))
	if err != nil {
		return nil, nil, err
	}

	udoc, err := o.store.FromKind("user").Get(ctx, ar.User)
	if err != nil {
		sentry.Report(err)
	} else {
		logged := services.FromContext(ctx).User.ID
		go o.notifier.Broadcast(ctx, models.UserActionAssign,
			*udoc.Data.(*models.User), *doc, "", ar.Position, logged, nil)
	}

	return o.store.Unwrap(ctx, id)
}

func (o *Organization) RemoveRole(ctx context.Context, id uuid.UUID, body io.Reader) (*models.Document, stores.Dependencies, error) {
	doc, err := o.store.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	org := doc.Data.(*models.Organization)

	if !Can(ctx, RemoveOrganizationRole, id, org.Country) {
		return nil, nil, ErrUnauthorized
	}

	var ar alterRole
	if err := json.NewDecoder(body).Decode(&ar); err != nil {
		return nil, nil, err
	}
	if ar.Position == "lear" || !validPosition(ar.Position) {
		return nil, nil, fmt.Errorf("%w: tried to delete LEAR", ErrBadInput)
	}
	roles := org.OrganizationRoles

	for _, role := range roles {
		if role.Position == ar.Position && role.UserID == ar.User {
			stores.AtomicDelete(o.store, role)
		}
	}

	udoc, err := o.store.FromKind("user").Get(ctx, ar.User)
	if err != nil {
		sentry.Report(err)
	} else {
		logged := services.FromContext(ctx).User.ID
		go o.notifier.Broadcast(ctx, models.UserActionRemove,
			*udoc.Data.(*models.User), *doc, "", ar.Position, logged, nil)
	}

	return o.store.Unwrap(ctx, id)
}

func validPosition(pos string) bool {
	positions := map[string]struct{}{"leaa": struct{}{}, "lsign": struct{}{}, "member": struct{}{}}
	_, ok := positions[pos]
	return ok
}

func (o *Organization) Validate(ctx context.Context, org uuid.UUID, status models.ValidationStatus, comment *string) error {
	doc, err := o.store.Get(ctx, org)
	if err != nil {
		return err
	}

	or := doc.Data.(*models.Organization)
	old := or.Valid
	if !Can(ctx, ValidateOrganization, or.ID, or.Country) {
		return ErrUnauthorized
	}
	or.Valid = status

	// Only do this for community organizations.
	if status == models.ValidationStatusValid && regexp.MustCompile(`^RC_+`).MatchString(or.VAT) {
		addCommunityOrg(ctx, o.store, or)
	}

	_, err = o.store.Update(ctx, doc)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	if err == nil {
		u := services.FromContext(ctx).User
		go o.notifier.Broadcast(ctx, models.UserActionUpdate, *u, *doc, old.String(), or.Valid.String(), u.ID, comment)
	}
	return nil
}

func addCommunityOrg(ctx context.Context, st stores.Store, org *models.Organization) {
	// Community organizations have VAT in the form RC_assetID.
	arr := strings.Split(org.VAT, "_")
	if _, err := uuid.Parse(arr[1]); err != nil {
		return
	}

	var asset models.Asset
	st.DB().Where("id = ?", arr[1]).First(&asset)
	if asset.ID == uuid.Nil {
		return
	}

	asset.ESCO = &org.ID
	if err := st.DB().Save(&asset).Error; err != nil {
		sentry.Report(err, "ValidateOrg.addCommunityORG2asset fails")
	}
}

func (o *Organization) RequestOrganizationMembership(ctx context.Context, oid uuid.UUID) error {
	doc, err := o.store.Get(ctx, oid)
	if err != nil {
		return err
	}
	org := doc.Data.(*models.Organization)
	if !Can(ctx, RequestOrganizationMembership, org.ID, org.Country) {
		return ErrUnauthorized
	}
	cv := services.FromContext(ctx)
	n := models.Notification{
		RecipientID: org.Roles.LEAR,
		Action:      models.UserActionRequestMembership,
		TargetID:    org.ID,
		TargetType:  models.OrganizationT,
		TargetKey:   org.Name,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		Country:     org.Country,
	}
	go o.notifier.Notify(ctx, &n)
	return nil
}

func (o *Organization) GetReport(ctx context.Context, first, offset int) ([]models.OrganizationReport, int, error) {
	// all orgs
	organizations, ids, total, err := getOrgs(ctx, o.store, services.FromContext(ctx).User, first, offset)
	if err != nil {
		return nil, 0, err
	}

	if err := organizations.PopulateLear(o.store.DB(), ids); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateLear fails")
		return nil, 0, err
	}

	// Own projects
	// populate total count of projects
	if err := organizations.CountTotalProjects(o.store.DB(), ids, true); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateOwnedTotalProjects fails")
		return nil, 0, err
	}
	// populate total count of in progress projects
	if err := organizations.CountOngoingProjects(o.store.DB(), ids, true); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateOwnedOngoingProjects fails")
		return nil, 0, err
	}
	// total for forfaiting agreement
	if err := organizations.CountFAProjects(o.store.DB(), ids, true); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateOwnedFAProjects fails")
		return nil, 0, err
	}

	if err := organizations.CountMonitoringPhaseProjects(o.store.DB(), ids, true); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateRelatedMPProjects fails")
		return nil, 0, err
	}

	// Related Projects
	// populate total count of projects
	if err := organizations.CountTotalProjects(o.store.DB(), ids, false); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateRelatedTotalProjects fails")
		return nil, 0, err
	}
	// populate total count of in progress projects
	if err := organizations.CountOngoingProjects(o.store.DB(), ids, false); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateRelatedOngoingProjects fails")
		return nil, 0, err
	}
	// total for forfaiting agreement
	if err := organizations.CountFAProjects(o.store.DB(), ids, false); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateRelatedFAProjects fails")
		return nil, 0, err
	}

	if err := organizations.CountMonitoringPhaseProjects(o.store.DB(), ids, false); err != nil {
		sentry.Report(err, "GetOrganizationReports.populateRelatedMPProjects fails")
		return nil, 0, err
	}

	return organizations.Transform(), total, nil
}

func getOrgs(ctx context.Context, st stores.Store, user *models.User, first, offset int) (models.OrganizationReportsList, []uuid.UUID, int, error) {
	db, can := auth(st.DB(), user)
	if !can {
		return nil, nil, 0, ErrUnauthorized
	}

	var (
		total int
		orgs  []models.Organization
	)

	// count
	db.Table("organizations").Count(&total)

	// list
	if offset != 0 {
		db = db.Offset(offset)

	}

	if first != 0 {
		db = db.Limit(first)
	}
	err := db.
		Order("created_at DESC").
		Find(&orgs).Error

	orgIDs := make([]uuid.UUID, len(orgs))
	result := make(models.OrganizationReportsList)

	for i, o := range orgs {
		report := models.OrganizationReport{
			UsersCount:         o.Roles.Len(),
			Name:               o.Name,
			VAT:                o.VAT,
			Address:            o.Address,
			Telephone:          o.Telephone,
			Website:            o.Website,
			LegalForm:          o.LegalForm,
			Registered:         o.Registered,
			Valid:              o.Valid,
			Roles:              o.Roles,
			Email:              o.Email,
			Country:            o.Country,
			RegistrationNumber: o.RegistrationNumber,
		}

		result[o.ID] = report
		orgIDs[i] = o.ID
	}

	return result, orgIDs, total, err
}

// auth checks whether a person have necessary auth level to make the
// call.  It will modify db according the user's roles. If the user do
// not have authorization level - it won't modify anything and will
// return 'false' for can.
func auth(db *gorm.DB, user *models.User) (*gorm.DB, bool) {
	can := new(bool)

	if user.SuperUser ||
		user.AdminNwManager ||
		user.PlatformManager {
		// superman kind of guy who can do anything
		return db, true
	}

	var countries = make(map[models.Country]bool)

	for _, role := range user.CountryRoles {
		if role.Role == models.FundManagerRole ||
			role.Role == models.CountryAdminRole {
			if !countries[role.Country] {
				countries[role.Country] = true
			}
		}
	}

	if len(countries) != 0 {
		// have some country role
		*can = true

		all := make([]models.Country, len(countries))
		i := 0
		for c := range countries {
			all[i] = c
			i++
		}

		db = db.Or("country in (?)", all)
	}

	for _, role := range user.OrganizationRoles {
		if role.Position == "lear" ||
			role.Position == "leaa" ||
			role.Position == "lsign" {
			*can = true
			db = db.Or("id = ? ", role.OrganizationID)
		}
	}

	return db, *can
}

func (o *Organization) ListByIDs(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID]models.Organization, error) {
	var orgs []models.Organization
	err := o.store.DB().
		Where("id in (?)", ids).
		Find(&orgs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]models.Organization)
	for _, v := range orgs {
		result[v.ID] = v
	}
	return result, nil
}

func (o *Organization) AcceptLEARApplication(ctx context.Context, uid, oid uuid.UUID, comment string, filename string, approved bool) error {
	doc, err := o.store.Get(ctx, oid)
	if err != nil {
		return err
	}
	org := doc.Data.(*models.Organization)
	if !Can(ctx, AcceptLEARApplication, org.ID, org.Country) {
		return ErrUnauthorized
	}

	utar, err := o.store.FromKind("user").Get(ctx, uid)
	if err != nil {
		return err
	}
	usr := utar.Data.(*models.User)

	admins, err := o.pf.GetPortfolioRolesPerCountry(ctx, org.Country, models.CountryAdminRole)
	if err != nil {
		return err
	}
	n := models.Notification{
		Action:     models.UserActionAcceptLEARApplication,
		TargetID:   oid,
		TargetType: models.OrganizationT,
		TargetKey:  org.Name,
		UserID:     usr.ID,
		UserKey:    usr.Email,
		New:        comment,
		Old:        filename,
		Country:    org.Country,
	}
	if !org.Country.IsConsortium() {
		admins = []uuid.UUID{getANWManager(o.store)}
	}
	if !approved {
		admins = []uuid.UUID{uid}
		n.Action = models.UserActionRejectLEARApplication
	}
	notifyAll(ctx, o.notifier, admins, n)
	return nil
}

func (o *Organization) ExportMeetings(ctx context.Context, oid uuid.UUID) (string, error) {
	doc, err := o.store.Get(ctx, oid)
	if err != nil {
		return "", err
	}
	org := doc.Data.(*models.Organization)
	if !Can(ctx, GetOrgMeetings, org.ID, org.Country) {
		return "", ErrUnauthorized
	}

	meetings, _, _, err := o.store.FromKind("meeting").ListByMember(ctx, stores.Filter{}, oid)
	if err != nil {
		return "", err
	}

	fpath := fmt.Sprintf("/tmp/meetings_%s.csv", oid.String())
	file, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	records := meetings2String(meetings)

	w := csv.NewWriter(file)
	w.WriteAll(records)

	return fpath, w.Error()
}

func meetings2String(docs []models.Document) [][]string {
	result := [][]string{
		{
			"name",
			"host",
			"location",
			"date",
			"objective",
			"stakeholder",
			"stage",
			"notes",
			"actions taken",
			"next contact",
			"project",
			"topic",
			"guests",
			"internal project",
		},
	}

	for _, m := range docs {
		result = append(result, m.Data.(*models.Meeting).Parse2String())
	}
	return result
}
