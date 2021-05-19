package stores

import (
	"context"
	"fmt"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type pstore struct {
	db *gorm.DB
}

func NewPortfolioStore(db *gorm.DB) Portfolio {
	return pstore{db: db}
}

func (p pstore) GetPDCountries(ctx context.Context, id uuid.UUID) []string {
	var res []string

	p.db.Table("country_roles").
		Select("country").
		Where("user_id = ? AND role = ?", id, models.PortfolioDirectorRole).
		Pluck("country", &res)

	return res
}

func (p pstore) GetPortfolioRole(ctx context.Context, country models.Country, role models.PortfolioRole) (uuid.UUID, error) {
	var usr struct{ UserID uuid.UUID }

	err := p.db.Table("country_roles").
		Select("user_id").
		Where("country = ? AND role = ?", country, role).
		Order("random()").
		First(&usr).Error

	if err != nil || usr.UserID == uuid.Nil {
		err = p.db.Table("country_roles").
			Where("role = ?", role).
			Select("user_id").
			First(&usr).Error
		if err != nil {
			return uuid.Nil, fmt.Errorf("no such user with %v role in %q: %v", role, country, err)
		}
	}

	return usr.UserID, nil
}

func (p pstore) GetPortfolioRolesPerCountry(ctx context.Context, country models.Country, role models.PortfolioRole) ([]uuid.UUID, error) {
	var userid []uuid.UUID
	return userid, p.db.Table("country_roles").
		Select("user_id").
		Where("role = ? AND country = ?", role, country).
		Pluck("user_id", &userid).Error
}

func (p pstore) GetPortfolioUsersPerCountry(ctx context.Context, country *string, role models.PortfolioRole) ([]models.User, error) {
	var users []models.User
	q := p.db.Table("users").
		Joins("left join country_roles on users.id = country_roles.user_id::UUID").
		Where("country_roles.role = ?", role)
	if country != nil {
		q = q.Where("country_roles.country = ?", country)
	}
	return users, q.Group("users.id").Find(&users).Error
}

func (p pstore) Put(ctx context.Context, user uuid.UUID, country models.Country, role models.PortfolioRole) error {
	pd := models.CountryRole{
		Country: country,
		UserID:  user,
		Role:    role,
	}

	return p.db.Create(&pd).Error
}

func (p pstore) Remove(ctx context.Context, user uuid.UUID, country models.Country, role models.PortfolioRole) error {
	return p.db.Table("country_roles").Where("user_id = ? AND country = ? AND role =?", user, country, role).Delete(nil).Error
}
