package controller

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

type Portfolio struct {
	pf stores.Portfolio
	n  stores.Notifier
	u  stores.Store
}

func NewPortfolio(e *services.Env) *Portfolio {
	return &Portfolio{
		pf: e.Portfolio,
		n:  e.Notifier,
		u:  e.UserStore,
	}
}

func (p Portfolio) AddCountryAdmin(ctx context.Context, uid uuid.UUID, country models.Country) error {
	if err := country.Valid(); err != nil {
		return err
	}

	if !Can(ctx, AddCountryAdmin, uuid.Nil, country) {
		return ErrUnauthorized
	}

	err := p.pf.Put(ctx, uid, country, models.CountryAdminRole)
	if stores.IsDuplicatedRecord(err) {
		return ErrDuplicate
	}

	return err
}

func (p Portfolio) RemoveCountryAdmin(ctx context.Context, uid uuid.UUID, country models.Country) error {
	if err := country.Valid(); err != nil {
		return err
	}
	if !Can(ctx, RemoveCountryAdmin, uuid.Nil, country) {
		return ErrUnauthorized
	}
	return p.pf.Remove(ctx, uid, country, models.CountryAdminRole)
}

func (p Portfolio) AddAdminNetworkManager(ctx context.Context, uid uuid.UUID) error {
	doc, err := p.u.Get(ctx, uid)
	if err != nil {
		return err
	}
	if !Can(ctx, AddAdminNetworkMan, uuid.Nil, doc.Data.(*models.User).Country) {
		return ErrUnauthorized
	}

	doc.Data.(*models.User).AdminNwManager = true
	_, err = p.u.Update(ctx, sanitizeInputFields(doc))
	if err != nil {
		return err
	}

	cv := services.FromContext(ctx)
	n := models.Notification{
		RecipientID: uid,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		Action:      models.UserActionAssign,
		TargetType:  models.UserT,
		TargetID:    uid,
		New:         "Admin Network Manager",
		Country:     cv.User.Country,
	}
	go p.n.Notify(ctx, &n)

	return nil
}

func (p Portfolio) RemoveAdminNetworkManager(ctx context.Context, uid uuid.UUID) error {
	doc, err := p.u.Get(ctx, uid)
	if err != nil {
		return err
	}
	if !Can(ctx, RemoveAdminNetworkMan, uuid.Nil, doc.Data.(*models.User).Country) {
		return ErrUnauthorized
	}

	doc.Data.(*models.User).AdminNwManager = false
	_, err = p.u.Update(ctx, sanitizeInputFields(doc))

	return err
}

func (p Portfolio) AddPortfolioRole(ctx context.Context, uid uuid.UUID, country models.Country, role models.PortfolioRole) error {
	if err := country.Valid(); err != nil {
		return err
	}

	if !Can(ctx, AddPortfolioRole, uuid.Nil, country) {
		return ErrUnauthorized
	}

	err := p.pf.Put(ctx, uid, country, role)
	if stores.IsDuplicatedRecord(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}

	taru, err := p.u.Get(ctx, uid)
	if err != nil {
		return err
	}

	cv := services.FromContext(ctx)
	n := models.Notification{
		RecipientID: uid,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		Action:      models.UserActionAssign,
		TargetType:  models.UserT,
		TargetID:    uid,
		TargetKey:   taru.Data.(*models.User).Name,
		New:         string(role),
		Country:     country,
	}
	go p.n.Notify(ctx, &n)

	return nil
}

func (p Portfolio) RemovePortfolioRole(ctx context.Context, uid uuid.UUID, country models.Country, role models.PortfolioRole) error {
	if err := country.Valid(); err != nil {
		return err
	}
	if !Can(ctx, RemovePortfolioRole, uuid.Nil, country) {
		return ErrUnauthorized
	}

	err := p.pf.Remove(ctx, uid, country, role)
	if err != nil {
		return err
	}

	taru, err := p.u.Get(ctx, uid)
	if err != nil {
		return err
	}

	cv := services.FromContext(ctx)
	n := models.Notification{
		RecipientID: uid,
		UserID:      cv.User.ID,
		UserKey:     cv.User.Name,
		Action:      models.UserActionRemove,
		TargetType:  models.UserT,
		TargetID:    uid,
		TargetKey:   taru.Data.(*models.User).Name,
		New:         string(role),
		Country:     country,
	}
	go p.n.Notify(ctx, &n)

	return nil
}

func (p Portfolio) GetPortfolioUsersPerCountry(ctx context.Context, country *string, role models.PortfolioRole) ([]models.User, error) {
	return p.pf.GetPortfolioUsersPerCountry(ctx, country, role)
}
