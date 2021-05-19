package controller

import (
	"context"
	"os"
	"time"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type WorkPhase struct {
	store      stores.Store
	notifier   stores.Notifier
	uploadPath string
}

func NewWorkPhase(env *services.Env) *WorkPhase {
	return &WorkPhase{
		store:      env.WPStore,
		notifier:   env.Notifier,
		uploadPath: env.Paths.Uploads,
	}
}

func (wp *WorkPhase) AdvanceToWorkPhase(ctx context.Context, pID uuid.UUID) (*models.Document, error) {
	doc, err := wp.store.FromKind("project").Get(ctx, pID)
	if err != nil {
		return nil, err
	}

	c := doc.Data.(*models.Project).Country
	if !Can(ctx, AdvanceProjectToWorkPhase, pID, c) {
		return nil, ErrUnauthorized
	}

	w := models.WorkPhase{Project: pID}

	w.Reviews = []models.WPReview{
		{
			Approved: false,
			Type:     models.WPReviewTypeFinancial,
		},
		{
			Approved: false,
			Type:     models.WPReviewTypeTechnical,
		},
		{
			Approved: false,
			Type:     models.WPReviewTypeExecutive,
		},
		{
			Approved: false,
			Type:     models.WPReviewTypeBankAccount,
		},
		{
			Approved: false,
			Type:     models.WPReviewTypeMaintenance,
		},
	}

	wpdoc, err := wp.store.Create(ctx, &w)

	if err != nil {
		return nil, err
	}

	cv := services.FromContext(ctx)
	go wp.notifier.Broadcast(ctx, models.UserActionCreate,
		*cv.User, *wpdoc, "", "", cv.User.ID, nil)

	return wpdoc, nil
}

func (wp *WorkPhase) ReviewWP(ctx context.Context, id uuid.UUID, review models.WPReview) error {
	wph, err := wp.store.Get(ctx, id)
	if err != nil {
		return err
	}
	prjd, err := wp.store.FromKind("project").Get(ctx, wph.Data.(*models.WorkPhase).Project)
	if err != nil {
		return err
	}
	p := prjd.Data.(*models.Project)
	if !wp.canReview(ctx, *p, review.Type) {
		return ErrUnauthorized
	}

	r := models.WPReview{
		WPID:     id,
		Type:     review.Type,
		Approved: review.Approved,
		Comment:  review.Comment,
		Author:   &services.FromContext(ctx).User.ID,
	}

	if review.ID != uuid.Nil {
		r.ID = review.ID
	}

	if r.Approved && r.Type == models.WPReviewTypeExecutive {
		cv := services.FromContext(ctx)

		n := models.Notification{
			Action:     models.UserActionApproveForfaitingPayment,
			UserID:     cv.User.ID,
			UserKey:    cv.User.Name,
			TargetID:   p.ID,
			TargetKey:  p.Name,
			TargetType: "project",
			New:        "",
			Old:        "",
			Country:    p.Country,
			Comment:    "",
		}

		// add PM and if there is fa manager as recipients
		var recs []uuid.UUID
		recs = append(recs, p.Roles.PM[0])
		if p.ForfaitingApplication != nil {
			recs = append(recs, p.ForfaitingApplication.ManagerID)
		}
		notifyAll(ctx, wp.notifier, recs, n)
	}

	return wp.store.DB().Save(&r).Error
}

func (wp *WorkPhase) canReview(ctx context.Context, p models.Project, t models.WPReviewType) bool {
	switch t {
	case models.WPReviewTypeMaintenance:
		return Can(ctx, WPReviewMaintenance, p.ID, p.Country)
	case models.WPReviewTypeFinancial,
		models.WPReviewTypeTechnical,
		models.WPReviewTypeExecutive,
		models.WPReviewTypeBankAccount:
		return Can(ctx, WPReview, p.ID, p.Country)
	default:
		return false
	}
}

func (wp *WorkPhase) FetchReviews(ctx context.Context,
	ids ...uuid.UUID) ([]models.WPReview, error) {

	var result []models.WPReview
	return result, wp.store.DB().
		Where("wp_id IN (?)", ids).
		Order("created_at DESC").Find(&result).Error
}

func (wp *WorkPhase) GetWP(ctx context.Context, wpID uuid.UUID) (*models.Document, stores.Dependencies, error) {
	db := wp.store.DB()
	var res models.WorkPhase

	err := db.Preload("Reviews", func(db *gorm.DB) *gorm.DB {
		return db.Order("wp_reviews.created_at DESC")
	}).Preload("Attachments").Where("id = ? ", wpID).Find(&res).Error

	if err != nil {
		return nil, nil, err
	}

	pdoc, err := wp.store.FromKind("project").Get(ctx, res.Project)
	if err != nil {
		return nil, nil, err
	}
	prj := pdoc.Data.(*models.Project)

	ids := []uuid.UUID{pdoc.ID}
	for _, id := range prj.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}
	if !canGetProject(ctx, GetWorkPhase, prj.Country, ids...) {
		return nil, nil, ErrUnauthorized
	}

	doc := models.Wrap(&res)
	doc.Attachments = make(map[string]models.Attachment)
	for _, att := range res.Attachments {
		doc.Attachments[att.Name] = att
	}

	return doc, nil, nil
}

func (wp *WorkPhase) UploadWP(ctx context.Context, wpID uuid.UUID, form RequestForm) error {
	doc, err := wp.store.FromKind("work_phase").Get(ctx, wpID)
	if err != nil {
		return err
	}
	pdoc, err := wp.store.FromKind("project").Get(ctx, doc.Data.(*models.WorkPhase).Project)
	if err != nil {
		return err
	}
	if !Can(ctx, UploadWorkPhase, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return ErrUnauthorized
	}

	return uploadFile(ctx, wp.store, wp.notifier, form, doc, wp.uploadPath)
}

func (wp *WorkPhase) GetFileWP(ctx context.Context, wpid uuid.UUID, filename string) (*models.Attachment, *os.File, error) {
	doc, err := wp.store.FromKind("work_phase").Get(ctx, wpid)
	if err != nil {
		return nil, nil, err
	}
	pdoc, err := wp.store.FromKind("project").Get(ctx, doc.Data.(*models.WorkPhase).Project)
	if err != nil {
		return nil, nil, err
	}
	if !Can(ctx, DownloadWorkPhaseFile, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return nil, nil, ErrUnauthorized
	}

	return getFile(ctx, wp.store, wpid, filename, wp.uploadPath)
}

func (wp *WorkPhase) DeleteFileWP(ctx context.Context, id uuid.UUID, filename string) error {
	doc, err := wp.store.FromKind("work_phase").Get(ctx, id)
	if err != nil {
		return err
	}
	pdoc, err := wp.store.FromKind("project").Get(ctx, doc.Data.(*models.WorkPhase).Project)
	if err != nil {
		return err
	}
	if !Can(ctx, DeleteWorkPhaseFile, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return ErrUnauthorized
	}

	return wp.store.DeleteAttachment(ctx, doc, filename)
}

type MonitoringPhase struct {
	store      stores.Store
	notifier   stores.Notifier
	uploadPath string
}

func NewMonitoringPhase(env *services.Env) *MonitoringPhase {
	return &MonitoringPhase{
		store:      env.MPStore,
		notifier:   env.Notifier,
		uploadPath: env.Paths.Uploads,
	}
}

func (mp *MonitoringPhase) AdvanceToMonitoringPhase(ctx context.Context, pID uuid.UUID) (*models.Document, error) {
	doc, err := mp.store.FromKind("project").Get(ctx, pID)
	if err != nil {
		return nil, err
	}
	prj := doc.Data.(*models.Project)
	if !Can(ctx, AdvanceProjectToMonitoringPhase, pID, prj.Country) {
		return nil, ErrUnauthorized
	}

	m := models.MonitoringPhase{Project: pID}
	m.Reviews = make([]models.MPReview, prj.ContractTerm)

	for i := 0; i < prj.ContractTerm; i++ {
		m.Reviews[i] = models.MPReview{
			Approved: false,
			Type:     models.MPReviewTypeForfaiting,
		}
	}
	mpdoc, err := mp.store.Create(ctx, &m)

	if err != nil {
		return nil, err
	}

	// a project's commissioning date should be set to one month after work phase
	// completion which matches monitoring phase beginning.
	doc.Data.(*models.Project).CommissioningDate = time.Now().AddDate(0, 1, 0)

	_, err = mp.store.Update(ctx, doc)
	if err != nil {
		return nil, err
	}

	cv := services.FromContext(ctx)
	go mp.notifier.Broadcast(ctx, models.UserActionCreate,
		*cv.User, *mpdoc, "", "", cv.User.ID, nil)

	return mpdoc, nil

}

func (mp *MonitoringPhase) GetMP(ctx context.Context, mpID uuid.UUID) (*models.Document, stores.Dependencies, error) {
	doc, err := mp.store.FromKind("monitoring_phase").Get(ctx, mpID)
	if err != nil {
		return nil, nil, err
	}
	pdoc, err := mp.store.FromKind("project").Get(ctx, doc.Data.(*models.MonitoringPhase).Project)
	if err != nil {
		return nil, nil, err
	}
	prj := pdoc.Data.(*models.Project)

	ids := []uuid.UUID{pdoc.ID}
	for _, id := range prj.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}
	if !canGetProject(ctx, GetMonitoringPhase, prj.Country, ids...) {
		return nil, nil, ErrUnauthorized
	}

	return mp.store.Unwrap(ctx, mpID)
}

func (mp *MonitoringPhase) UploadMP(ctx context.Context, mpID uuid.UUID, form RequestForm) error {
	doc, err := mp.store.FromKind("monitoring_phase").Get(ctx, mpID)
	if err != nil {
		return err
	}
	pdoc, err := mp.store.FromKind("project").Get(ctx, doc.Data.(*models.MonitoringPhase).Project)
	if err != nil {
		return err
	}
	if !Can(ctx, UploadMonitoringPhase, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return ErrUnauthorized
	}
	doc, err = mp.store.Get(ctx, mpID)
	if err != nil {
		return err
	}

	return uploadFile(ctx, mp.store, mp.notifier, form, doc, mp.uploadPath)
}

func (mp *MonitoringPhase) GetFileMP(ctx context.Context, mpID uuid.UUID, filename string) (*models.Attachment, *os.File, error) {
	doc, err := mp.store.FromKind("monitoring_phase").Get(ctx, mpID)
	if err != nil {
		return nil, nil, err
	}
	pdoc, err := mp.store.FromKind("project").Get(ctx, doc.Data.(*models.MonitoringPhase).Project)
	if err != nil {
		return nil, nil, err
	}
	if !Can(ctx, DownloadMonitoringPhaseFile, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return nil, nil, ErrUnauthorized
	}

	return getFile(ctx, mp.store, mpID, filename, mp.uploadPath)
}

func (mp *MonitoringPhase) DeleteFileMP(ctx context.Context, id uuid.UUID, filename string) error {
	doc, err := mp.store.FromKind("monitoring_phase").Get(ctx, id)
	if err != nil {
		return err
	}
	pdoc, err := mp.store.FromKind("project").Get(ctx, doc.Data.(*models.MonitoringPhase).Project)
	if err != nil {
		return err
	}
	if !Can(ctx, DeleteMonitoringPhaseFile, pdoc.ID, pdoc.Data.(*models.Project).Country) {
		return ErrUnauthorized
	}

	return mp.store.DeleteAttachment(ctx, doc, filename)
}

func (mp *MonitoringPhase) ReviewMP(ctx context.Context, id uuid.UUID, review models.MPReview) error {
	mph, err := mp.store.Get(ctx, id)
	if err != nil {
		return err
	}
	prjd, err := mp.store.FromKind("project").Get(ctx, mph.Data.(*models.MonitoringPhase).Project)
	if err != nil {
		return err
	}
	p := prjd.Data.(*models.Project)
	if !Can(ctx, MPReview, p.ID, p.Country) {
		return ErrUnauthorized
	}

	r := models.MPReview{
		MPID:     id,
		Type:     review.Type,
		Approved: review.Approved,
		Comment:  review.Comment,
		Author:   &services.FromContext(ctx).User.ID,
	}

	if review.ID != uuid.Nil {
		r.ID = review.ID
	}

	return mp.store.DB().Save(&r).Error
}

func (mp *MonitoringPhase) FetchReviews(ctx context.Context,
	ids ...uuid.UUID) ([]models.MPReview, error) {

	var result []models.MPReview
	return result, mp.store.DB().
		Where("mp_id IN (?)", ids).
		Order("created_at DESC").Find(&result).Error
}
