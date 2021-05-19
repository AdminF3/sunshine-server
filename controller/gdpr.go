package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type GDPR struct {
	db     *gorm.DB
	nt     stores.Notifier
	pf     stores.Portfolio
	st     stores.Store
	uppath string // upload path
}

func NewGDPR(env *services.Env) *GDPR {
	return &GDPR{
		db:     env.DB,
		nt:     env.Notifier,
		pf:     env.Portfolio,
		uppath: env.Paths.Uploads,
		st:     env.GDPRStore,
	}
}

func (g GDPR) SendRequest(ctx context.Context, req *models.GDPRRequest, u []Upload) error {
	gReqDoc, err := g.st.Create(ctx, req)
	if err != nil {
		return err
	}

	if err := uploadGQLFiles(g.st, u, gReqDoc.ID, g.uppath); err != nil {
		return fmt.Errorf("fail to upload file: %w", err)
	}

	gReqDoc, err = g.st.GetByIndex(ctx, gReqDoc.ID.String())
	if err != nil {
		return err
	}

	req = gReqDoc.Data.(*models.GDPRRequest)

	reqJSON, err := json.Marshal(struct {
		R models.GDPRRequest           `json:"GDPRRequest"`
		A map[string]models.Attachment `json:"Attachments"`
	}{
		R: *req,
		A: gReqDoc.Attachments,
	})

	if err != nil {
		return err
	}

	// find the target user with his email
	var usr models.User
	err = g.st.DB().Where("email = ?", req.Email).First(&usr).Error
	if err != nil || usr.ID == uuid.Nil || !usr.Country.IsConsortium() {
		// find the admin network manager
		var anm models.User
		err = g.st.DB().First(&anm).Where("admin_network_manager = ?", true).Error
		if err != nil {
			return err
		}

		// country is not in consortium or
		// user is not found, send notification to ANM
		n := models.Notification{
			RecipientID: anm.ID,
			Action:      models.UserActionGDPR,
			UserID:      uuid.Nil,
			UserKey:     "",
			New:         string(reqJSON),
			Country:     models.CountryLatvia,
		}
		go g.nt.Notify(ctx, &n)
		return nil
	}

	// country is in the consortium, send notification to the dpo
	// find the dpo of the country
	dpo, err := g.pf.GetPortfolioRole(ctx, usr.Country, models.DataProtectionOfficerRole)
	if err != nil {
		return err
	}
	n := models.Notification{
		RecipientID: dpo,
		Action:      models.UserActionGDPR,
		UserID:      usr.ID,
		UserKey:     usr.Key(),
		New:         string(reqJSON),
		Country:     usr.Country,
	}

	go g.nt.Notify(ctx, &n)

	return nil
}

func (g GDPR) List(ctx context.Context, first, offset int) ([]models.Document, int, error) {
	cv := services.FromContext(ctx)
	if !Can(ctx, ListGDPRRequests, uuid.Nil, cv.User.Country) {
		return nil, 0, ErrUnauthorized
	}
	docs, _, total, err := g.st.List(ctx, stores.Filter{Limit: first, Offset: offset})
	return docs, total, err

}

func (g GDPR) Get(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	cv := services.FromContext(ctx)

	if !cv.Authorized() {
		return nil, ErrUnauthorized
	}

	if !Can(ctx, GetGDPRRequest, uuid.Nil, cv.User.Country) {
		return nil, ErrUnauthorized
	}
	return g.st.Get(ctx, id)
}
