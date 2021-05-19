package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

type User struct {
	st stores.Store
	ts stores.TokenStore
	m  services.Mailer
	n  stores.Notifier
	pf stores.Portfolio

	uploadPath string
}

func NewUser(env *services.Env) *User {
	return &User{
		st:         env.UserStore,
		ts:         env.TokenStore,
		m:          env.Mailer,
		n:          env.Notifier,
		pf:         env.Portfolio,
		uploadPath: env.Paths.Uploads,
	}
}

func (u *User) Create(ctx context.Context, rc io.ReadCloser) (*models.Document, error) {
	var user models.User

	if err := json.NewDecoder(rc).Decode(&user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	user.Valid = models.ValidationStatusRegistered
	user.SuperUser = false
	user.IsActive = false
	user.PlatformManager = false
	user.AdminNwManager = false

	if _, err := u.st.GetByIndex(ctx, user.Email); err == nil {
		return nil, fmt.Errorf("%w: email", ErrDuplicate)
	}

	doc, err := u.st.Create(ctx, sanitizeEntityFields(&user))
	if err != nil {
		return nil, err
	}

	token, err := u.ts.Create(ctx, models.CreateToken, doc.ID)
	if err != nil {
		return nil, err
	}

	go services.NewUserEmail(u.m, user, token.ID)

	return doc, nil
}

func (u *User) Get(ctx context.Context, userID uuid.UUID) (*models.Document, stores.Dependencies, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, nil, ErrUnauthorized
	}

	doc, err := u.st.Get(ctx, userID)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	if !Can(ctx, GetUser, userID, doc.Data.(*models.User).Country) {
		return nil, nil, ErrUnauthorized
	}

	return u.st.Unwrap(ctx, userID)
}

func (u *User) Update(ctx context.Context, userID uuid.UUID, r io.Reader) (*models.Document, stores.Dependencies, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, nil, ErrUnauthorized
	}

	doc, err := u.st.Get(ctx, userID)
	if err != nil {
		return nil, nil, ErrNotFound
	}
	if !Can(ctx, UpdateUser, userID, doc.Data.(*models.User).Country) {
		return nil, nil, ErrUnauthorized
	}

	olddoc, _ := u.st.Get(ctx, userID)

	oldValid := doc.Data.(*models.User).Valid

	new := *doc.Data.(*models.User)
	if err = json.NewDecoder(r).Decode(&new); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	old := doc.Data.(*models.User)
	new.Password = old.Password
	new.IsActive = old.IsActive
	new.Value = old.Value
	updated := models.NewDocument(&new)

	if !services.FromContext(ctx).User.SuperUser {
		new.SuperUser = false
	}

	if stores.ShouldInvalidate(olddoc, updated, services.FromContext(ctx).User.SuperUser, oldValid) {
		new.Valid = models.ValidationStatusPending
		usr := updated.Data.(*models.User)

		admins, err := u.pf.GetPortfolioRolesPerCountry(ctx, usr.Country, models.CountryAdminRole)
		if err == nil {
			not := models.Notification{

				Action:     models.UserActionUpdate,
				UserID:     cv.User.ID,
				UserKey:    cv.User.Name,
				TargetID:   usr.ID,
				TargetType: models.UserT,
				TargetKey:  usr.Name,
				Old:        "valid",
				New:        "invalid",
				Country:    usr.Country,
			}
			notifyAll(ctx, u.n, admins, not)
		} else {
			sentry.Report(err)
		}

	}

	doc.Data = &new

	_, err = u.st.Update(ctx, sanitizeInputFields(doc))
	if err != nil {
		return nil, nil, err
	}

	result, deps, err := u.st.Unwrap(ctx, userID)

	return result, deps, err
}

func (u *User) List(ctx context.Context, filter stores.Filter) ([]models.Document, stores.Dependencies, int, error) {
	cv := services.FromContext(ctx)

	if !cv.Authorized() || !Can(ctx, ListUsers, uuid.Nil, cv.User.Country) {
		return nil, nil, 0, ErrUnauthorized
	}

	docs, deps, n, err := u.st.List(ctx, filter)

	return docs, deps, n, err
}

func (u *User) AdminList(ctx context.Context, filter stores.Filter) ([]models.User, int, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() || !Can(ctx, ListAdminUsers, uuid.Nil, cv.User.Country) {
		return nil, 0, ErrUnauthorized
	}
	// find all user ids in portfolio roles with investor role
	var invs []string
	u.st.DB().Where("role = ?", "investor").Find(&models.CountryRole{}).Pluck("user_id", &invs)

	var n int
	var adms []models.User
	//finds all the admin roles
	//the limit(-1) is to remove the limit so Count() can work
	err := u.st.DB().Where("is_admin = ?", true).
		Or("platform_manager = ?", true).
		Or("admin_network_manager = ?", true).
		Or("id IN(?)", invs).
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&adms).
		Offset(0).
		Limit(-1).
		Count(&n).
		Error

	return adms, n, err
}

func (u *User) UploadFile(ctx context.Context, uid uuid.UUID, form RequestForm) error {
	cv := services.FromContext(ctx)
	if !Can(ctx, UploadUser, uid, cv.User.Country) {
		return ErrUnauthorized
	}

	udoc, err := u.st.Get(ctx, uid)
	if err != nil {
		return err
	}

	return uploadFile(ctx, u.st, u.n, form, udoc, u.uploadPath)
}

func (u *User) GetFile(ctx context.Context, uid uuid.UUID, filename string) (*models.Attachment, *os.File, error) {
	doc, err := u.st.Get(ctx, uid)
	if err != nil {
		return nil, nil, err
	}

	// special case for becoming a lear - that organization's lear
	// should be able to see your application
	att, file, err := getFile(ctx, u.st, uid, filename, u.uploadPath)
	if err != nil {
		return nil, nil, err
	}
	if att.UploadType == "lear apply" {
		cv := services.FromContext(ctx)
		cmnt := &RequestFormComment{}
		err := json.Unmarshal([]byte(att.Comment), cmnt)
		if err != nil {
			return nil, nil, err
		}
		if !Can(ctx, AcceptLEARApplication, cmnt.OrganizationID, cv.User.Country) {
			return nil, nil, ErrUnauthorized
		}
	}

	// Everybody should access other users' avatars.
	avatar := filepath.Base(doc.Data.(*models.User).Avatar)
	if filename != avatar && att.UploadType != "lear apply" {
		if !Can(ctx, DownloadUserFile, uid, doc.Data.(*models.User).Country) {
			return nil, nil, ErrUnauthorized
		}
	}

	return att, file, nil
}

func (u *User) DeleteFile(ctx context.Context, uid uuid.UUID, filename string) error {
	cv := services.FromContext(ctx)
	if !Can(ctx, DeleteUserFile, uid, cv.User.Country) {
		return ErrUnauthorized
	}

	doc, err := u.st.Get(ctx, uid)
	if err != nil {
		return err
	}

	user := doc.Data.(*models.User)
	avatar := filepath.Base(user.Avatar)
	idcard := filepath.Base(user.Identity)

	var shouldUpdate bool

	if filename == avatar {
		user.Avatar = ""
		shouldUpdate = true
	}

	if filename == idcard {
		user.Identity = ""
		user.Valid = models.ValidationStatusPending
		shouldUpdate = true
	}

	if shouldUpdate {
		if _, err := u.st.Update(ctx, doc); err != nil {
			return err
		}
	}

	return u.st.DeleteAttachment(ctx, doc, filename)
}

func (u *User) ResendActivationEmail(ctx context.Context, email string) error {
	doc, err := u.st.GetByIndex(ctx, email)
	if err != nil {
		return err
	}

	token, err := u.ts.Create(ctx, models.CreateToken, doc.ID)
	if err != nil {
		return err
	}
	user := doc.Data.(*models.User)

	go services.NewUserEmail(u.m, *user, token.ID)

	return nil
}

func (u *User) Validate(ctx context.Context, uid uuid.UUID, status models.ValidationStatus, comment *string) error {
	doc, err := u.st.Get(ctx, uid)
	if err != nil {
		return err
	}

	usr := doc.Data.(*models.User)
	if !Can(ctx, ValidateUser, usr.ID, usr.Country) {
		return ErrUnauthorized
	}

	doc.Data.(*models.User).Valid = status

	_, err = u.st.Update(ctx, doc)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadInput, err)
	}
	cv := services.FromContext(ctx)

	cmnt := ""
	if comment != nil {
		cmnt = *comment
	}
	n := models.Notification{
		RecipientID: usr.ID,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		TargetID:    usr.ID,
		TargetKey:   usr.Name,
		TargetType:  models.UserT,
		Action:      models.UserActionUpdate,
		Old:         doc.Data.(*models.User).Valid.String(),
		New:         status.String(),
		Country:     usr.Country,
		Comment:     cmnt,
	}
	go u.n.Notify(ctx, &n)

	return nil
}

func (u *User) FetchUsers(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	var result []models.User
	return result, u.st.DB().
		Where("id IN (?)", ids).
		Order("created_at DESC").Find(&result).Error
}
