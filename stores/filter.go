package stores

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"

	"stageai.tech/sunshine/sunshine/models"
)

// PlatformRole role is role in the platform.
type PlatformRole string

const (
	// PlatformManager is user with `platform_manager=true`.
	PlatformManager PlatformRole = "platform_manager"
	// AdminNetworkManager is user with `admin_network_manager=true`.
	AdminNetworkManager PlatformRole = "admin_network_manager"
)

// Filter holds option values for filtering and pagination.
type Filter struct {
	Offset              int
	Limit               int
	Ascending           bool
	Search              string
	Status              int
	BuildingType        int
	LegalForm           int
	Country             models.Country
	PlatformManger      bool
	AdminNetworkManager bool

	// PlatformRoles are platform roles part of the users
	// entity.
	PlatformRoles []string

	// CountryRoles is applicable filter for listing users.
	CountryRoles []string

	// Owner is applicable only for assets and projects.
	Owner uuid.UUID

	// ESCO is applicable only for assets and projects.
	ESCO uuid.UUID

	// AssetOwner is applicable only for projects.
	AssetOwner uuid.UUID

	// RelatedOrganizationID is organization UUID related to
	// projects. It is used for consortium organization filter. It
	// is applicable only for projects.
	RelatedOrganizationID uuid.UUID

	// NullFields holds slice of fields that will be included
	// in Filter results only if NULL.
	NullFields []string
}

func (f Filter) GORM(db *gorm.DB) *gorm.DB {
	var o = "created_at"
	if !f.Ascending {
		o += " DESC"
	}
	db = db.Order(o)

	if f.Offset > 0 {
		db = db.Offset(f.Offset)
	}

	if f.Limit > 0 {
		db = db.Limit(f.Limit)
	}

	return db
}
