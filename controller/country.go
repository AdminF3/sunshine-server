package controller

import (
	"context"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

type Country struct {
	st         stores.Store
	notifier   stores.Notifier
	uploadPath string
}

func NewCountry(env *services.Env) *Country {
	return &Country{
		st:         env.CountryStore,
		notifier:   env.Notifier,
		uploadPath: env.Paths.Uploads,
	}
}

func (c *Country) SetVat(ctx context.Context, country models.Country, vat int) (*models.CountryVat, error) {
	var ctrv models.CountryVat
	if !Can(ctx, SetVat, uuid.Nil, country) {
		return nil, ErrUnauthorized
	}
	return &ctrv, c.st.DB().Model(&ctrv).Where("country = ?", country).Update("vat", vat).Error
}

func (c *Country) GetCountry(ctx context.Context, country models.Country) (*models.CountryVat, error) {
	var ctrv models.CountryVat
	if !Can(ctx, GetCountry, uuid.Nil, country) {
		return nil, ErrUnauthorized
	}
	return &ctrv, c.st.DB().Where("country = ?", country).First(&ctrv).Error
}
