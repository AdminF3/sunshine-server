package stores

import (
	"context"
	"errors"
	"time"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

var errInvalidToken = errors.New("invalid token")

type tokenStore struct {
	db       *gorm.DB
	validate *validator.Validate
}

// NewTokenStore creates new DB.
func NewTokenStore(db *gorm.DB, v *validator.Validate) *tokenStore {
	return &tokenStore{db: db, validate: v}
}

func (s tokenStore) Create(ctx context.Context, p models.TokenPurpose, id uuid.UUID) (*models.Token, error) {
	token := models.NewToken(p, id)
	return token, s.db.Create(&token).Error
}

func (s tokenStore) Get(ctx context.Context, p models.TokenPurpose, id uuid.UUID) (*models.Token, error) {
	var token models.Token
	if err := s.db.Where("id = ? AND purpose = ?", id, p).First(&token).Error; err != nil {
		return nil, err
	}

	var err error
	if !models.ValidToken(token, p, s.validate) {
		err = errInvalidToken
	}
	return &token, err
}

func (s tokenStore) Invalidate(ctx context.Context, p models.TokenPurpose, id uuid.UUID) error {
	token, err := s.Get(ctx, p, id)
	if err != nil && !errors.Is(err, errInvalidToken) {
		return err
	}

	token.TTL = gorm.NowFunc().Sub(token.CreatedAt.Add(1 * time.Hour))
	return s.db.Save(token).Error
}
