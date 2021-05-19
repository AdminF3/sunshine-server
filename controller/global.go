package controller

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Global struct {
	DB *gorm.DB
}

func NewGlobal(env *services.Env) *Global {
	return &Global{
		DB: env.DB,
	}
}

func (c *Global) AddEUROBOR(ctx context.Context, value float64) error {
	if !Can(ctx, addEurobor, uuid.Nil, models.CountryLatvia) {
		return ErrUnauthorized
	}

	v := struct {
		Value float64
	}{
		Value: value,
	}

	return c.DB.Table("eurobor").Save(&v).Error
}
