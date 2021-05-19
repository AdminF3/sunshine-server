package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type ForfaitingAgreement struct {
	st         stores.Store
	fpst       stores.Store
	notifier   stores.Notifier
	pf         stores.Portfolio
	uploadPath string
}

func NewForfaitingAgreement(env *services.Env) *ForfaitingAgreement {
	return &ForfaitingAgreement{
		st:         env.FAStore,
		fpst:       env.FPStore,
		notifier:   env.Notifier,
		uploadPath: env.Paths.Uploads,
		pf:         env.Portfolio,
	}
}

func (f *ForfaitingAgreement) Create(ctx context.Context, fa models.ForfaitingApplication) (*models.ForfaitingApplication, error) {
	prjd, err := f.st.FromKind("project").Get(ctx, fa.Project)
	if err != nil {
		return nil, err
	}
	prj := prjd.Data.(*models.Project)

	if !Can(ctx, CreateFA, fa.Project, prj.Country) {
		return nil, ErrUnauthorized
	}

	if prj.Milestone.Index() < models.MilestoneProjectDesign.Index() {
		return nil, fmt.Errorf("%w: milestone is lower than 'forfaiting_payout'", ErrBadInput)
	}

	fa.Reviews = []models.FAReview{
		{Approved: false, Type: models.FAReviewTypeExecutive},
		{Approved: false, Type: models.FAReviewTypeFinancial},
		{Approved: false, Type: models.FAReviewTypeGuidelines},
		{Approved: false, Type: models.FAReviewTypeTechnical},
	}

	fms, err := f.pf.GetPortfolioRolesPerCountry(ctx, prj.Country, models.FundManagerRole)
	if err == nil {
		cv := services.FromContext(ctx)
		n := models.Notification{
			Action:     models.UserActionForfaitingApplication,
			UserID:     cv.User.ID,
			UserKey:    cv.User.Name,
			TargetID:   prj.ID,
			TargetType: models.ProjectT,
			TargetKey:  prj.Name,
			Country:    prj.Country,
		}
		if !prj.Country.IsConsortium() {
			fms = []uuid.UUID{getANWManager(f.st)}
		}
		notifyAll(ctx, f.notifier, fms, n)
	}

	// start transaction block
	tx := f.st.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	if err := tx.Create(&fa).Error; err != nil {
		if stores.IsDuplicatedRecord(err) {
			err = ErrDuplicate
		}
		tx.Rollback()
		return nil, err
	}

	// add tama role for manager
	err = addContactRole(tx, fa.ManagerID, fa.Project)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}
	// end transaction block

	return &fa, err
}

func (f *ForfaitingAgreement) Get(ctx context.Context, id uuid.UUID) (*models.ForfaitingApplication, error) {
	fad, err := f.st.FromKind("forfaiting_application").Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fa := fad.Data.(*models.ForfaitingApplication)

	if !f.can(ctx, GetFA, fa.Project) {
		return nil, ErrUnauthorized
	}

	return fa, nil
}

func (f *ForfaitingAgreement) Review(ctx context.Context,
	id uuid.UUID, review models.FAReview) error {
	fad, err := f.st.Get(ctx, id)
	if err != nil {
		return err
	}
	fa := fad.Data.(*models.ForfaitingApplication)

	if !f.can(ctx, ReviewFA, fa.Project) {
		return ErrUnauthorized
	}

	r := models.FAReview{
		ForfaitingApplicationID: id,
		Type:                    review.Type,
		Approved:                review.Approved,
		Comment:                 review.Comment,
		Author:                  &services.FromContext(ctx).User.ID,
	}

	if review.ID != uuid.Nil {
		r.ID = review.ID
	}

	if r.Approved && r.Type == models.FAReviewTypeExecutive {
		cv := services.FromContext(ctx)
		prjd, err := f.st.FromKind("project").Get(ctx, fa.Project)
		if err != nil {
			return err
		}
		prj := prjd.Data.(*models.Project)

		n := models.Notification{
			Action:     models.UserActionApproveForfaitingApplication,
			UserID:     cv.User.ID,
			UserKey:    cv.User.Name,
			TargetID:   prj.ID,
			TargetKey:  prj.Name,
			TargetType: "project",
			New:        "",
			Old:        "",
			Country:    prj.Country,
			Comment:    "",
		}

		// add PM and if there is fund manager as recipients
		var recs []uuid.UUID
		if len(prj.Roles.PM) > 0 {
			recs = append(recs, prj.Roles.PM[0])
		}
		if prj.FundManager != nil {
			recs = append(recs, *prj.FundManager)
		}
		notifyAll(ctx, f.notifier, recs, n)
	}

	return f.st.DB().Save(&r).Error
}

func (f *ForfaitingAgreement) can(ctx context.Context, a Action, prjID uuid.UUID) bool {
	prjd, err := f.st.FromKind("project").Get(ctx, prjID)
	if err != nil {
		return false
	}
	prj := prjd.Data.(*models.Project)

	return Can(ctx, a, prj.ID, prj.Country)
}

func (f *ForfaitingAgreement) FetchReviews(ctx context.Context,
	ids ...uuid.UUID) ([]models.FAReview, error) {

	var result []models.FAReview
	return result, f.st.DB().
		Where("forfaiting_application_id IN (?)", ids).
		Order("created_at DESC").Find(&result).Error
}

func (f *ForfaitingAgreement) FetchAttachments(ctx context.Context, attType string, ids ...uuid.UUID) ([]models.Attachment, error) {
	var result []models.Attachment

	return result, f.st.DB().
		Where("owner_id IN (?)", ids).
		Where("upload_type = ?", attType).
		Order("created_at DESC").Find(&result).Error
}

func (f *ForfaitingAgreement) ListByCountries(ctx context.Context,
	countries []models.Country) ([]models.ForfaitingApplication, error) {

	for _, c := range countries {
		if !Can(ctx, ListFAByCountry, uuid.Nil, c) {
			return nil, ErrUnauthorized
		}
	}

	var result []models.ForfaitingApplication
	return result, f.st.DB().
		Joins(`INNER JOIN projects
			ON forfaiting_applications.project_id = projects.id`).
		Where("projects.country IN (?)", countries).
		Order("forfaiting_applications.created_at DESC").
		Find(&result).Error
}

func (f *ForfaitingAgreement) Update(ctx context.Context, faID uuid.UUID, fa models.ForfaitingApplication) (*models.ForfaitingApplication, error) {
	fad, err := f.st.FromKind("forfaiting_application").Get(ctx, faID)
	if err != nil {
		return nil, err
	}
	foundFA := fad.Data.(*models.ForfaitingApplication)

	if !f.can(ctx, UpdateFA, foundFA.Project) {
		return nil, ErrUnauthorized
	}

	if fa.ManagerID != uuid.Nil {
		mngr, err := f.st.FromKind("user").Get(ctx, fa.ManagerID)
		// we don't want the whole update to fail if an invalid user is
		// given as manager to be updated so we will just ignore the field
		// in such a case
		if err == nil {
			foundFA.ManagerID = fa.ManagerID
			foundFA.Manager = *mngr.Data.(*models.User)
		}
	}

	if fa.BankAccount.IBAN != "" {
		foundFA.BankAccount.IBAN = fa.BankAccount.IBAN
	}

	if fa.BankAccount.BeneficiaryName != "" {
		foundFA.BankAccount.BeneficiaryName = fa.BankAccount.BeneficiaryName
	}

	if fa.BankAccount.BankNameAddress != "" {
		foundFA.BankAccount.BankNameAddress = fa.BankAccount.BankNameAddress
	}

	if fa.Finance != models.Finance(0) {
		foundFA.Finance = fa.Finance
	}

	// start transaction block
	tx := f.st.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	foundFA.PrivateBond = fa.PrivateBond

	if err := tx.Save(&foundFA).Error; err != nil {
		if stores.IsDuplicatedRecord(err) {
			err = ErrDuplicate
		}
		tx.Rollback()
		return nil, err
	}

	// add tama role for manager
	err = addContactRole(tx, foundFA.ManagerID, foundFA.Project)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}
	// end transaction block

	return foundFA, nil
}

func (f *ForfaitingAgreement) GetByProject(ctx context.Context, pid uuid.UUID) (*models.ForfaitingApplication, error) {
	if !f.can(ctx, GetFAByProject, pid) {
		return nil, ErrUnauthorized
	}

	var result models.ForfaitingApplication

	return &result, f.st.DB().
		Where("project_id = ?", pid).
		First(&result).Error
}

func (f *ForfaitingAgreement) UploadFile(ctx context.Context, faid uuid.UUID, form RequestForm) error {
	fadoc, err := f.st.Get(ctx, faid)
	if err != nil {
		return err
	}

	fa := fadoc.Data.(*models.ForfaitingApplication)

	cv := services.FromContext(ctx)
	if !Can(ctx, UploadFAAttachment, fa.Project, cv.User.Country) {
		return ErrUnauthorized
	}

	return uploadFile(ctx, f.st, f.notifier, form, fadoc, f.uploadPath)
}

func (f *ForfaitingAgreement) GetFile(ctx context.Context, faid uuid.UUID, filename string) (*models.Attachment, *os.File, error) {
	fadoc, err := f.st.Get(ctx, faid)
	if err != nil {
		return nil, nil, err
	}
	pdoc, err := f.st.FromKind("project").Get(ctx, fadoc.Data.(*models.ForfaitingApplication).Project)
	if err != nil {
		return nil, nil, err
	}
	if !Can(ctx, GetFAAttachment, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return nil, nil, ErrUnauthorized
	}

	return getFile(ctx, f.st, faid, filename, f.uploadPath)
}

func (f *ForfaitingAgreement) DeleteFile(ctx context.Context, faid uuid.UUID, filename string) error {
	fadoc, err := f.st.Get(ctx, faid)
	if err != nil {
		return err
	}
	pdoc, err := f.st.FromKind("project").Get(ctx, fadoc.Data.(*models.ForfaitingApplication).Project)
	if err != nil {
		return err
	}
	if !Can(ctx, DeleteFAAttachment, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return ErrUnauthorized
	}

	return f.st.DeleteAttachment(ctx, fadoc, filename)
}

func (f *ForfaitingAgreement) CreateFP(ctx context.Context, transferValue int, currency models.Currency, pid uuid.UUID, td *time.Time) (*models.ForfaitingPayment, error) {
	if !f.can(ctx, CreateFP, pid) {
		return nil, ErrUnauthorized
	}
	fp := models.ForfaitingPayment{
		TransferValue: transferValue,
		Currency:      currency,
		Project:       pid}
	if td != nil {
		fp.TransferDate = *td
	}
	doc, err := f.fpst.Create(ctx, &fp)
	if err != nil {
		return nil, err
	}
	return doc.Data.(*models.ForfaitingPayment), err
}

func (f *ForfaitingAgreement) GetFP(ctx context.Context, id, pid uuid.UUID) (*models.ForfaitingPayment, error) {
	if !f.can(ctx, GetFP, pid) {
		return nil, ErrUnauthorized
	}

	fpd, err := f.fpst.FromKind("forfaiting_payment").Get(ctx, id)
	if err != nil {
		return nil, err
	}
	fp := fpd.Data.(*models.ForfaitingPayment)

	return fp, nil
}

func (f *ForfaitingAgreement) UpdateFP(ctx context.Context, faid, pid uuid.UUID, transferValue *int, currency *models.Currency, td *time.Time) (*models.ForfaitingPayment, error) {
	if !f.can(ctx, UpdateFP, pid) {
		return nil, ErrUnauthorized
	}

	fp, err := f.GetFP(ctx, faid, pid)
	if err != nil {
		return nil, err
	}

	if transferValue != nil {
		fp.TransferValue = *transferValue
	}

	if currency != nil {
		fp.Currency = *currency
	}

	if td != nil {
		fp.TransferDate = *td
	}

	return fp, f.st.DB().Save(fp).Error
}

func addContactRole(db *gorm.DB, managerID, projectID uuid.UUID) error {
	// check the manager for tama role in the project
	var ur models.ProjectRole
	err := db.Where("user_id = ?  ", managerID).
		Where("project_id = ?", projectID).
		Where("position = ?", "tama").
		First(&ur).Error

	if stores.IsRecordNotFound(err) {
		// if the manager doesnt have tama role in project
		// add tama role to fa manager
		ur = models.ProjectRole{
			UserID:    managerID,
			ProjectID: projectID,
			Position:  "tama",
		}
		// using tx to create a transaction block
		err = db.Create(&ur).Error
		if err != nil {
			return err
		}
	}

	return err
}
