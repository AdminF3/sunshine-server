package controller

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

type Project struct {
	st         stores.Store
	token      stores.TokenStore
	notifier   stores.Notifier
	pf         stores.Portfolio
	uploadPath string
	validator  *validator.Validate
}

func NewProject(env *services.Env) *Project {
	return &Project{
		st:         env.ProjectStore,
		token:      env.TokenStore,
		notifier:   env.Notifier,
		pf:         env.Portfolio,
		uploadPath: env.Paths.Uploads,
		validator:  env.Validator,
	}
}

func (p *Project) Create(ctx context.Context, r io.Reader) (*models.Document, stores.Dependencies, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, nil, ErrUnauthorized
	}

	var proj models.Project
	if err := json.NewDecoder(r).Decode(&proj); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	odoc, err := p.st.FromKind("organization").Get(ctx, proj.Owner)
	if err != nil {
		return nil, nil, err
	}

	org := odoc.Data.(*models.Organization)
	adoc, err := p.st.FromKind("asset").Get(ctx, proj.Asset)
	if err != nil {
		return nil, nil, err
	}

	asset := adoc.Data.(*models.Asset)
	if err := p.checkProjectAuth(ctx, org, asset); err != nil {
		return nil, nil, err
	}

	proj.Country = asset.Country
	proj.Milestone = models.MilestoneZero
	proj.Roles = models.ProjRoles{}
	proj.AssetSnapshot = asset.Snapshot()
	proj.ProjectRoles = append(proj.ProjectRoles, models.ProjectRole{UserID: cv.User.ID, ProjectID: proj.ID, Position: "pm"})
	proj.PortfolioDirector, err = p.pf.GetPortfolioRole(ctx, asset.Country, models.PortfolioDirectorRole)
	if err != nil {
		return nil, nil, err
	}

	doc, err := p.st.Create(ctx, &proj)
	if err != nil {
		return nil, nil, err
	}

	if _, err := p.st.FromKind("contract").Create(ctx, contract.New(doc.ID)); err != nil {
		return nil, nil, err
	}
	doc = sanitizeInputFields(doc)

	return p.st.Unwrap(ctx, doc.ID)
}

func (p *Project) Get(ctx context.Context, id uuid.UUID) (*models.Document, stores.Dependencies, error) {
	doc, err := p.st.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	prj := doc.Data.(*models.Project)
	ids := []uuid.UUID{id, prj.Owner}
	for _, id := range prj.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}

	if !canGetProject(ctx, GetProject, prj.Country, ids...) {
		return nil, nil, ErrUnauthorized
	}

	return p.st.Unwrap(ctx, id)
}

// canGetProject
func canGetProject(ctx context.Context, action Action, c models.Country, ids ...uuid.UUID) bool {
	for _, id := range ids {
		if Can(ctx, action, id, c) {
			return true
		}
	}
	return false
}

func (p *Project) Update(ctx context.Context, id uuid.UUID, r io.Reader) (*models.Document, stores.Dependencies, error) {
	doc, err := p.st.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	proj := *doc.Data.(*models.Project)

	if !Can(ctx, UpdateProject, id, proj.Country) {
		return nil, nil, ErrUnauthorized
	}
	// save old values
	oldAT := proj.AirTemperature
	oldWT := proj.WaterTemperature
	oldGS := proj.GuaranteedSavings
	oldCT := proj.ContractTerm
	oldFm := proj.FundManager
	if err = json.NewDecoder(r).Decode(&proj); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	updated := models.NewDocument(proj)
	// shouldInvalidate is used on projects only to send notifications, it does not invalidate
	// The status is not used in this case so I just pass one as it is mandatory - projects don't have valid field anyway
	if stores.ShouldInvalidate(doc, updated, services.FromContext(ctx).User.SuperUser, models.ValidationStatusValid) {
		var recipients []uuid.UUID
		recipients = append(recipients, proj.Roles.PM...)
		recipients = append(recipients, proj.Roles.PaCo...)
		recipients = append(recipients, proj.Roles.PLSign...)

		cv := services.FromContext(ctx)

		for _, r := range recipients {
			n := models.Notification{
				RecipientID: r,
				Action:      models.UserActionUpdate,
				UserID:      cv.User.ID,
				UserKey:     cv.User.Name,
				TargetID:    proj.ID,
				TargetType:  models.ProjectT,
				Old:         "valid",
				New:         "invalid",
				Country:     proj.Country,
			}
			go p.notifier.Notify(ctx, &n)
		}
	}

	// check for fund manager change try
	if (proj.FundManager != oldFm) && !Can(ctx, ChangeFundManager, id, proj.Country) {
		return nil, nil, ErrUnauthorized
	}

	lid := services.FromContext(ctx).User.ID
	ustore := p.st.FromKind("user")
	userdoc, uerr := ustore.Get(ctx, lid)
	if uerr != nil {
		// Basically this should never happened.
		sentry.Report(uerr, "Logged with unexciting user")
		return nil, nil, uerr
	}

	if userdoc.Data.(*models.User).SuperUser {
		pd, err := setpd(ctx, ustore, proj)
		if err != nil {
			return nil, nil, ErrFatal
		}

		proj.PortfolioDirector = pd
	} else {
		// not logged as admin so PD cannot be updated.
		proj.PortfolioDirector = doc.Data.(*models.Project).PortfolioDirector
	}

	var c models.Country
	if proj.Country != "" {
		c = proj.Country
	} else {
		c = doc.Data.(*models.Project).Country
	}

	proj.Value = doc.Data.(*models.Project).Value
	doc.Data = &proj

	if _, err = p.st.Update(ctx, sanitizeInputFields(doc)); err != nil {
		return nil, nil, err
	}
	// check for new values
	if oldAT != proj.AirTemperature ||
		oldWT != proj.WaterTemperature ||
		oldGS != proj.GuaranteedSavings ||
		oldCT != proj.ContractTerm {

		if err := calcContract(p.st.DB(), proj.ID); err != nil {
			return nil, nil, err
		}
	}
	p.pf.Put(ctx, proj.PortfolioDirector, c, models.PortfolioDirectorRole)

	return p.st.Unwrap(ctx, id)
}
func (p *Project) List(ctx context.Context, f stores.Filter, id uuid.UUID) ([]models.Document, stores.Dependencies, int, error) {
	if id != uuid.Nil {
		return p.st.ListByMember(ctx, f, id)
	}

	return p.st.List(ctx, f)
}

func (p *Project) Reports(ctx context.Context, filter stores.Filter) ([]models.Document, stores.Dependencies, int, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, nil, 0, ErrUnauthorized
	}

	db, can := projectReportAuth(p.st.DB(), cv.User)
	if !can {
		return nil, nil, 0, ErrUnauthorized
	}

	var count int
	db.Table("projects").Count(&count)

	if filter.Offset != 0 {
		db = db.Offset(filter.Offset)
	}

	if filter.Limit != 0 {
		db = db.Limit(filter.Limit)
	}

	var projects []models.Project

	if err := db.Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, nil, 0, err
	}

	docs := make([]models.Document, 0)
	for _, p := range projects {
		tmp := p
		docs = append(docs, *models.NewDocument(&tmp))
	}

	deps, errm := stores.UnwrapDeps(p.st, docs)

	return docs, deps, count, errm
}

func projectReportAuth(db *gorm.DB, user *models.User) (*gorm.DB, bool) {
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

	orgs := []uuid.UUID{}

	for _, role := range user.OrganizationRoles {
		// if user have significant role - project must be
		// with owner == orgID. Otherwise it is fine the orgID
		// to be in consortiumOrgs
		orgs = append(orgs, role.OrganizationID)
	}

	if len(orgs) > 0 {
		db = db.Joins("INNER JOIN organizations ON projects.owner = organizations.id").
			Or("owner IN (?) ", orgs).
			Or("? = ANY(projects.consortium_orgs)", orgs)

		// Project.asset –>
		//    asset.Esco ->
		//        org -> user with any role in org

		db = db.Joins("INNER JOIN assets on projects.asset = assets.id").
			Or("assets.esco_id in (?)", orgs)

		*can = true
	}

	return db, *can
}

func (p *Project) UploadFile(ctx context.Context, pid uuid.UUID, form RequestForm) error {
	doc, err := p.st.Get(ctx, pid)
	if err != nil {
		return err
	}

	proj := *doc.Data.(*models.Project)
	if !Can(ctx, UploadProject, pid, proj.Country) {
		return ErrUnauthorized
	}

	return uploadFile(ctx, p.st, p.notifier, form, doc, p.uploadPath)
}

func (p *Project) GetFile(ctx context.Context, pid uuid.UUID, filename string) (*models.Attachment, *os.File, error) {
	doc, err := p.st.Get(ctx, pid)
	if err != nil {
		return nil, nil, err
	}

	prj := doc.Data.(*models.Project)

	if !Can(ctx, DownloadProjectFile, pid, prj.Country) {
		return nil, nil, ErrUnauthorized
	}

	return getFile(ctx, p.st, pid, filename, p.uploadPath)
}

func (p *Project) DeleteFile(ctx context.Context, pid uuid.UUID, filename string) error {
	doc, err := p.st.Get(ctx, pid)
	if err != nil {
		return err
	}

	proj := *doc.Data.(*models.Project)
	if !Can(ctx, DeleteProjectFile, pid, proj.Country) {
		return ErrUnauthorized
	}

	return p.st.DeleteAttachment(ctx, doc, filename)
}

func (p *Project) AddRole(ctx context.Context, id uuid.UUID, body io.Reader) (*models.Document, stores.Dependencies, error) {
	doc, err := p.st.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	proj := doc.Data.(*models.Project)

	if !Can(ctx, AddProjectRole, id, proj.Country) {
		return nil, nil, ErrUnauthorized
	}

	var ar alterRole
	if err := json.NewDecoder(body).Decode(&ar); err != nil {
		return nil, nil, err
	}

	roles := proj.ProjectRoles
	switch ar.Position {
	case "plsign", "paco", "tama", "teme":
		roles = append(roles, models.ProjectRole{
			UserID:    ar.User,
			ProjectID: proj.ID,
			Position:  ar.Position,
		})
	case "pm":
		return nil, nil, errors.New("tried to assing pm")
	default:
		return nil, nil, fmt.Errorf("%w: non existing position", ErrBadInput)
	}

	proj.ProjectRoles = setUniqueProjRoles(roles)
	_, err = p.st.Update(ctx, sanitizeInputFields(doc))
	if err != nil {
		return nil, nil, err
	}

	udoc, err := p.st.FromKind("user").Get(ctx, ar.User)
	if err != nil {
		sentry.Report(err)
	} else {
		logged := services.FromContext(ctx).User.ID
		go p.notifier.Broadcast(ctx, models.UserActionAssign,
			*udoc.Data.(*models.User), *doc, "", ar.Position, logged, nil)
	}
	return p.st.Unwrap(ctx, id)
}

func (p *Project) RemoveRole(ctx context.Context, id uuid.UUID, body io.Reader) (*models.Document, stores.Dependencies, error) {
	doc, err := p.st.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	prj := doc.Data.(*models.Project)

	if !Can(ctx, RemoveProjectRole, id, prj.Country) {
		return nil, nil, ErrUnauthorized
	}

	var ar alterRole
	if err := json.NewDecoder(body).Decode(&ar); err != nil {
		return nil, nil, err
	}

	for _, role := range prj.ProjectRoles {
		if role.Position == ar.Position && role.UserID == ar.User {
			stores.AtomicDelete(p.st, role)
		}
	}

	udoc, err := p.st.FromKind("user").Get(ctx, ar.User)
	if err != nil {
		sentry.Report(err)
	} else {
		logged := services.FromContext(ctx).User.ID
		go p.notifier.Broadcast(ctx, models.UserActionRemove,
			*udoc.Data.(*models.User), *doc, "", ar.Position, logged, nil)
	}
	return p.st.Unwrap(ctx, id)
}

func (p *Project) AssignProjectRoles(ctx context.Context, pid uuid.UUID, members []uuid.UUID, action Action) error {
	doc, err := p.st.Get(ctx, pid)
	if err != nil {
		return ErrNotFound
	}
	prj := doc.Data.(*models.Project)

	var target uuid.UUID

	roles := prj.ProjectRoles

	switch action {
	case AssignPM:
		for _, uid := range members {
			roles = append(roles, models.ProjectRole{
				UserID:    uid,
				ProjectID: pid,
				Position:  "pm",
			})
		}

		target = prj.ID
	default:
		err := fmt.Errorf("%w: wrong action while assigning project role: %v", ErrFatal, action)
		sentry.Report(err, "Assign project role with wrong action")
		return ErrFatal
	}

	if !Can(ctx, action, target, prj.Country) {
		return ErrUnauthorized
	}

	prj.ProjectRoles = roles
	doc.Data = prj
	_, err = p.st.Update(ctx, sanitizeInputFields(doc))
	return err
}

func (p *Project) RequestProjectCreation(ctx context.Context, asset, org uuid.UUID) error {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return ErrUnauthorized
	}

	adoc, e := p.st.FromKind("asset").Get(ctx, asset)
	if e != nil {
		return e
	}
	aent := adoc.Data.(*models.Asset)

	var orgRole struct {
		Position       string
		UserID         uuid.UUID
		OrganizationID uuid.UUID
	}

	err := p.st.DB().Table("organization_roles").
		Where("organization_id = ? ", aent.Owner).
		Where("position = ?", "lear").
		Find(&orgRole).Error

	if err != nil {
		return fmt.Errorf("fail getting org lear: %w", err)
	}

	req := models.ProjectCreationRequest{
		Asset:        asset,
		Organization: org,
		User:         cv.User.ID,
		Status:       models.OpenedStatus,
	}

	err = p.st.DB().Create(&req).Error
	if err != nil {
		if stores.IsDuplicatedRecord(err) {
			err = fmt.Errorf("%w, %v", ErrDuplicate, err)
		}
		return err
	}

	n := models.Notification{
		RecipientID: orgRole.UserID,
		Action:      models.UserActionRequestProjectCreation,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		TargetID:    asset,
		TargetKey:   aent.Address,
		TargetType:  models.AssetT,
		New:         "requested",
		Country:     aent.Country,
	}
	go p.notifier.Notify(ctx, &n)

	return err
}

func (p *Project) ProcessProjectRequest(ctx context.Context, user, asset uuid.UUID, isApprove bool) error {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return ErrUnauthorized
	}

	ad, err := p.st.FromKind("asset").Get(ctx, asset)
	if err != nil {
		return fmt.Errorf("fetch asset: %w", err)
	}
	aent := ad.Data.(*models.Asset)

	org := aent.Owner
	var auth bool

	for _, or := range cv.User.OrganizationRoles {
		if or.OrganizationID == org && or.Position == "lear" {
			auth = true
			break
		}
	}

	if !auth {
		return ErrUnauthorized
	}

	var req models.ProjectCreationRequest
	err = p.st.DB().
		Where("asset_id = ? AND user_id = ?", asset, user).
		First(&req).Error
	if err != nil {
		if stores.IsRecordNotFound(err) {
			err = ErrNotFound
		}
		return err
	}

	req.Status = models.RejectedStatus

	if isApprove {
		tk, err := p.token.Create(ctx, models.CreateProjectToken, user)
		if err != nil {
			return err
		}

		req.Status = models.AcceptedStatus
		req.Token = &tk.ID
	}

	if err = p.st.DB().Save(&req).Error; err != nil {
		return err
	}

	n := models.Notification{
		RecipientID: user,
		Action:      models.UserActionRequestProjectCreation,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		TargetID:    aent.ID,
		TargetKey:   aent.Address,
		TargetType:  models.AssetT,
		New:         string(req.Status),
		Country:     aent.Country,
	}
	go p.notifier.Notify(ctx, &n)

	return nil
}

func setUniqueProjRoles(roles []models.ProjectRole) []models.ProjectRole {
	unique := make(map[models.ProjectRole]struct{}, len(roles))
	for _, m := range roles {
		unique[m] = struct{}{}
	}

	updated := make([]models.ProjectRole, 0, len(unique))
	for k := range unique {
		updated = append(updated, k)
	}

	return updated
}

// setpd checks if the PD in proj model is valid or existing user and if it is not - it returns one random PD.
//
// If the uuid of proj.PortfolioDirector is legit (he is already a PD for a
// country) it is just returned. If it is legit not a PD for any country - it
// is set by the Portfolio as a valid PD for proj.Country and also returned.
func setpd(ctx context.Context, us stores.Store, proj models.Project) (uuid.UUID, error) {
	_, err := us.Get(ctx, proj.PortfolioDirector)
	if err != nil {
		return us.Portfolio().GetPortfolioRole(ctx, proj.Country, models.PortfolioDirectorRole)
	} else {
		c := us.Portfolio().GetPDCountries(ctx, proj.PortfolioDirector)
		if len(c) != 0 {
			return proj.PortfolioDirector, nil
		}
		err := us.Portfolio().Put(ctx, proj.PortfolioDirector, proj.Country, models.PortfolioDirectorRole)
		if err != nil {
			return uuid.Nil, err
		}
		return proj.PortfolioDirector, nil
	}
}

func (p *Project) checkProjectAuth(ctx context.Context, org *models.Organization, asset *models.Asset) error {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return ErrUnauthorized
	}

	if org.ID == asset.Owner {
		// Default case where the owner of the asset is making
		// a project.
		if !Can(ctx, superuser|lear|lsigns|leaas|members, org.ID, org.Country) {
			return ErrUnauthorized
		}
	} else {
		// A guest organization is trying to make a
		// project. It must have ProjectCreationRequest which
		// must be approved.

		tids, err := getActiveTokens(ctx, p.st.DB(), p.token, cv.User.ID)
		if err != nil || len(tids) == 0 {
			return ErrUnauthorized
		}

		for _, id := range tids {
			err := p.token.Invalidate(ctx, models.CreateProjectToken, id)
			if err != nil {
				sentry.Report(err)
			}
		}
	}

	if org.Valid != models.ValidationStatusValid {
		return fmt.Errorf("%w: organization is not validated", ErrBadInput)
	}
	if asset.Valid != models.ValidationStatusValid {
		return fmt.Errorf("%w: asset is not validated", ErrBadInput)
	}

	return nil
}

func getActiveTokens(ctx context.Context,
	db *gorm.DB,
	tst stores.TokenStore,
	uid uuid.UUID) ([]uuid.UUID, error) {

	var ts []models.Token
	if err := db.Where("user_id = ?", uid).
		Where("purpose = ?", models.CreateProjectToken).
		Find(&ts).Error; err != nil {
		return nil, err
	}

	if len(ts) == 0 {
		// no tokens at all
		return nil, ErrUnauthorized
	}

	result := make([]uuid.UUID, 0)

	for _, token := range ts {
		token, err := tst.Get(ctx, models.CreateProjectToken, token.ID)
		if err != nil || token.ID == uuid.Nil {
			// invalid token, skip
			continue
		}

		result = append(result, token.ID)
	}
	return result, nil
}

func (p *Project) ListByIDs(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID]models.Project, error) {
	var projects []models.Project
	err := p.st.DB().
		Where("id in (?)", ids).
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]models.Project)
	for _, v := range projects {
		result[v.ID] = v
	}
	return result, nil
}

func (p *Project) CommentProject(ctx context.Context, pid uuid.UUID, content string, topic *string) (*models.Project, error) {
	doc, err := p.st.FromKind("project").Get(ctx, pid)
	if err != nil {
		return nil, err
	}

	if !Can(ctx, CommentProject, doc.ID, doc.Data.(*models.Project).Country) {
		return nil, ErrUnauthorized
	}

	comment := models.ProjectComment{
		UserID:    services.FromContext(ctx).User.ID,
		ProjectID: pid,
		Content:   content,
	}

	if topic != nil {
		comment.Topic = *topic
	}

	if err := p.st.DB().Create(&comment).Error; err != nil {
		return nil, err
	}

	pdoc, err := p.st.Get(ctx, pid)
	return pdoc.Data.(*models.Project), err
}

func (p *Project) AdvanceToMilestone(ctx context.Context, id uuid.UUID, m models.Milestone) error {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return ErrUnauthorized
	}

	if !Can(ctx, AdvanceProjectToMilestone, id, cv.User.Country) {
		return ErrUnauthorized
	}

	doc, err := p.st.Get(ctx, id)
	if err != nil {
		return err
	}

	doc.Data.(*models.Project).Milestone = m

	_, err = p.st.Update(ctx, doc)
	return err
}

func (p *Project) ExportMeetings(ctx context.Context, prjID uuid.UUID) (string, error) {
	doc, err := p.st.Get(ctx, prjID)
	if err != nil {
		return "", err
	}
	proj := doc.Data.(*models.Project)
	if !Can(ctx, GetPrjMeetings, proj.ID, proj.Country) {
		return "", ErrUnauthorized
	}

	meetings, _, _, err := p.st.FromKind("meeting").ListByMember(ctx, stores.Filter{}, prjID)
	if err != nil {
		return "", err
	}

	fpath := fmt.Sprintf("/tmp/meetings_%s.csv", prjID.String())
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
