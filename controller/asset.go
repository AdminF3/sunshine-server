package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Asset struct {
	store      stores.Store
	orgStore   stores.Store
	notifier   stores.Notifier
	pf         stores.Portfolio
	uploadPath string
}

func NewAsset(env *services.Env) *Asset {
	return &Asset{
		store:      env.AssetStore,
		orgStore:   env.OrganizationStore,
		notifier:   env.Notifier,
		uploadPath: env.Paths.Uploads,
		pf:         env.Portfolio,
	}
}

func (a *Asset) Create(ctx context.Context, r io.Reader) (*models.Document, stores.Dependencies, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, nil, ErrUnauthorized
	}

	var asset models.Asset
	if err := json.NewDecoder(r).Decode(&asset); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	org, err := a.store.FromKind("organization").Get(ctx, asset.Owner)
	if err != nil {
		return nil, nil, err
	}

	var authorized = false
	lid := cv.User.ID
	for _, role := range org.Data.(*models.Organization).OrganizationRoles {
		if lid == role.UserID {
			authorized = true
			break
		}
	}

	if !authorized {
		return nil, nil, ErrUnauthorized
	}

	asset.Valid = models.ValidationStatusRegistered
	doc, err := a.store.Create(ctx, &asset)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	ast := doc.Data.(*models.Asset)
	if ast.Owner != uuid.Nil {
		go a.notifier.Broadcast(ctx, models.UserActionCreate, *cv.User, *doc, "", ast.Valid.String(), cv.User.ID, nil)
	}

	return a.store.Unwrap(ctx, doc.ID)
}

func (a *Asset) can(ctx context.Context, action Action, asset models.Asset) bool {
	return Can(ctx, action, asset.Owner, asset.Country) ||
		(asset.ESCO != nil && Can(ctx, action, *asset.ESCO, asset.Country))
}

func (a *Asset) Get(ctx context.Context, id uuid.UUID) (*models.Document, stores.Dependencies, error) {
	doc, deps, err := a.store.Unwrap(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if !a.can(ctx, GetAsset, *doc.Data.(*models.Asset)) {
		return nil, nil, ErrUnauthorized
	}

	return doc, deps, err
}

func (a *Asset) Update(ctx context.Context, id uuid.UUID, r io.Reader) (*models.Document, stores.Dependencies, error) {
	doc, err := a.store.Get(ctx, id)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	old, _ := a.store.Get(ctx, id)

	asset := doc.Data.(*models.Asset)
	docVal := asset.Value
	oldValid := asset.Valid
	oldFields := asset.ExcerptString()

	if !a.can(ctx, UpdateAsset, *asset) {
		return nil, nil, ErrUnauthorized
	}

	if err := json.NewDecoder(r).Decode(&asset); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	updated := models.NewDocument(asset)
	if stores.ShouldInvalidate(old, updated, services.FromContext(ctx).User.SuperUser, oldValid) {
		updated.Data.(*models.Asset).Valid = models.ValidationStatusPending

		newFields := updated.Data.(*models.Asset).ExcerptString()
		cv := services.FromContext(ctx)

		admins, err := a.pf.GetPortfolioRolesPerCountry(ctx, asset.Country, models.CountryAdminRole)
		if err == nil {
			n := models.Notification{
				Action:     models.UserActionUpdate,
				UserID:     cv.User.ID,
				UserKey:    cv.User.Name,
				TargetID:   asset.ID,
				TargetType: models.AssetT,
				TargetKey:  asset.Address,
				Old:        oldFields,
				New:        newFields,
				Country:    asset.Country,
			}
			if !asset.Country.IsConsortium() {
				admins = []uuid.UUID{getANWManager(a.store)}
			}
			notifyAll(ctx, a.notifier, admins, n)
		} else {
			sentry.Report(err)
		}
	}

	doc.Data = updated.Data
	doc.Data.(*models.Asset).Value = docVal
	_, err = a.store.Update(ctx, sanitizeInputFields(doc))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	return a.store.Unwrap(ctx, id)
}

func (a *Asset) List(ctx context.Context, id uuid.UUID, filter stores.Filter) ([]models.Document, stores.Dependencies, int, error) {
	var (
		docs []models.Document
		deps stores.Dependencies
		n    int
		err  error
	)

	var orgsIDs []uuid.UUID

	if id != uuid.Nil {
		// get all organizations that a user is member of.
		orgs, _, _, oerr := a.store.FromKind("organization").ListByMember(
			ctx, stores.Filter{}, id)
		if oerr != nil {
			return nil, nil, 0, fmt.Errorf("%w: %v", ErrBadInput, err)
		}

		for _, v := range orgs {
			orgsIDs = append(orgsIDs, v.ID)
		}

		docs, deps, n, err = a.store.ListByMember(ctx, filter, orgsIDs...)
	} else {
		docs, deps, n, err = a.store.List(ctx, filter)
	}

	if err != nil && len(docs) == 0 {
		return nil, nil, 0, fmt.Errorf("%w: %v", ErrFatal, err)
	}

	return docs, deps, n, err
}

func (a *Asset) Reports(ctx context.Context, filter stores.Filter) ([]models.Document, int, error) {
	user := services.FromContext(ctx).User

	db, can := assetReportAuth(a.store.DB(), user)
	if !can {
		return nil, 0, ErrUnauthorized
	}

	var c int

	// count
	db.Table("assets").Count(&c)

	// list
	if filter.Offset != 0 {
		db = db.Offset(filter.Offset)
	}

	if filter.Limit != 0 {
		db = db.Limit(filter.Limit)
	}

	var assets []models.Asset

	err := db.Order("created_at DESC").Find(&assets).Error

	// TODO Remove those nonsense when models.Document is removed!
	docs := make([]models.Document, 0)
	for _, a := range assets {
		docs = append(docs, *models.NewDocument(&a))
	}

	return docs, c, err
}

// TODO this is same as organization.go#auth, refactor it.
// add test cases for this
func assetReportAuth(db *gorm.DB, user *models.User) (*gorm.DB, bool) {
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
		if role.Position == "lear" ||
			role.Position == "leaa" ||
			role.Position == "lsign" {
			orgs = append(orgs, role.OrganizationID)
		}
	}

	if len(orgs) > 0 {
		db = db.Joins("INNER JOIN organizations ON assets.owner_id = organizations.id").
			Or("owner_id IN (?) ", orgs)
		*can = true
	}

	return db, *can
}

func (a *Asset) UploadFile(ctx context.Context, aid uuid.UUID, form RequestForm) error {
	doc, err := a.store.Get(ctx, aid)
	if err != nil {
		return err
	}

	if !a.can(ctx, UploadAsset, *doc.Data.(*models.Asset)) {
		return ErrUnauthorized
	}

	return uploadFile(ctx, a.store, a.notifier, form, doc, a.uploadPath)
}

func (a *Asset) GetFile(ctx context.Context, aid uuid.UUID, filename string) (*models.Attachment, *os.File, error) {
	doc, err := a.store.Get(ctx, aid)
	if err != nil {
		return nil, nil, err
	}

	if !a.can(ctx, DownloadAssetFile, *doc.Data.(*models.Asset)) {
		return nil, nil, ErrUnauthorized
	}

	return getFile(ctx, a.store, aid, filename, a.uploadPath)
}

func (a *Asset) DeleteFile(ctx context.Context, id uuid.UUID, filename string) error {
	doc, err := a.store.Get(ctx, id)
	if err != nil {
		return err
	}

	if !a.can(ctx, DeleteAssetFile, *doc.Data.(*models.Asset)) {
		return ErrUnauthorized
	}

	return a.store.DeleteAttachment(ctx, doc, filename)
}

func (a *Asset) Validate(ctx context.Context, asset uuid.UUID, status models.ValidationStatus, comment *string) error {
	doc, err := a.store.Get(ctx, asset)
	if err != nil {
		return err
	}

	as := doc.Data.(*models.Asset)
	if !a.can(ctx, ValidateAsset, *as) {
		return ErrUnauthorized
	}

	as.Valid = status

	_, err = a.store.Update(ctx, sanitizeInputFields(doc))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	if err == nil {
		user := services.FromContext(ctx).User
		go a.notifier.Broadcast(ctx, models.UserActionUpdate, *user, *doc, "", as.Valid.String(), user.ID, comment)
	}
	return nil
}

func (a *Asset) ResidentsCount(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID]int, error) {
	var set []struct {
		ID    uuid.UUID
		Count int
	}

	err := a.store.DB().
		Select("assets.ID, COUNT(organization_roles.*)").
		Table("organization_roles").
		Joins("inner join organizations on organization_roles.organization_id = organizations.id").
		Joins("inner join assets on assets.esco_id = organizations.id").
		Where("assets.id in (?)", ids).
		Group("assets.ID").
		Find(&set).Error

	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int)
	for _, v := range set {
		result[v.ID] = v.Count
	}

	return result, nil
}

func (a *Asset) ClaimAssetResidency(ctx context.Context, assetID uuid.UUID) error {
	doc, err := a.store.Get(ctx, assetID)
	if err != nil {
		return err
	}

	asset := doc.Data.(*models.Asset)
	if !Can(ctx, ClaimAssetResidency, uuid.Nil, asset.Country) {
		return ErrUnauthorized
	}

	// Get the organization which is owner of the asset
	doc, err = a.orgStore.Get(ctx, asset.Owner)
	if err != nil {
		return err
	}
	org := doc.Data.(*models.Organization)

	cv := services.FromContext(ctx)
	n := models.Notification{
		RecipientID: org.Roles.LEAR,
		Action:      models.UserActionClaimResidency,
		TargetID:    asset.ID,
		TargetType:  models.AssetT,
		TargetKey:   asset.Address,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		Country:     asset.Country,
	}
	go a.notifier.Notify(ctx, &n)
	return nil
}

func (a *Asset) OwnerNames(ctx context.Context, assets ...uuid.UUID) (map[uuid.UUID]string, error) {
	var set []struct {
		ID   uuid.UUID
		Name string
	}

	if err := a.store.DB().
		Select("assets.ID, organizations.Name").
		Table("assets").
		Joins("inner join organizations on assets.owner_id = organizations.id").
		Where("assets.id in (?)", assets).
		Group("assets.ID, organizations.Name").
		Find(&set).Error; err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]string)
	for _, v := range set {
		result[v.ID] = v.Name
	}

	return result, nil
}
