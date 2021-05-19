package models

import (
	"encoding/json"
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Value

	Name            string  `json:"name" validate:"required"`
	Email           string  `json:"email" validate:"email,required"`
	Password        string  `json:"password,omitempty" validate:"required"`
	Address         string  `json:"address"`
	Avatar          string  `json:"avatar"`
	Identity        string  `json:"identity"`
	Telephone       string  `json:"telephone"`
	SuperUser       bool    `json:"superuser" gorm:"column:is_admin"`
	PlatformManager bool    `json:"platform_manager" gorm:"column:platform_manager"`
	AdminNwManager  bool    `json:"admin_nw_manager" gorm:"column:admin_network_manager"`
	Country         Country `json:"country" validate:"required"`
	IsActive        bool    `json:"is_active" gorm:"column:is_active"`

	Valid             ValidationStatus   `json:"valid" gorm:"column:status"`
	SocialProfiles    []SocialProfile    `json:"social_profiles" gorm:"foreignkey:UserID"`
	ProjectRoles      []ProjectRole      `json:"project_roles" gorm:"foreignkey:UserID"`
	OrganizationRoles []OrganizationRole `json:"organization_roles" gorm:"foreignkey:UserID"`
	CountryRoles      []CountryRole      `json:"country_roles" gorm:"foreignkey:UserID"`
}

func (User) Kind() string {
	return "user"
}

func (u User) Key() string {
	return u.Email
}

func (User) IsEntity() {}

func (u *User) BeforeCreate() error {
	return u.SetPassword(u.Password)
}

func (User) TableName() string {
	return "users"
}

// SetPassword hashes and sets given password.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.Password = string(hash)
	return err
}

func (u User) Dependencies() []config.Dependency {
	return []config.Dependency{}
}

func (u *User) MarshalJSON() ([]byte, error) {
	type usrt User
	usr := usrt(*u)
	usr.Password = ""

	return json.Marshal(usr)
}

type SocialProfile struct {
	Value

	UserID uuid.UUID `json:"user_id"`
	Type   string    `json:"type"`
	Handle string    `json:"handle"`
}

func (SocialProfile) TableName() string {
	return "social_profiles"
}

type CountryRole struct {
	ID uuid.UUID `gorm:"type:uuid; primary_key"`

	Country Country       `json:"country"`
	UserID  uuid.UUID     `json:"user_id"`
	Role    PortfolioRole `json:"role"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (CountryRole) TableName() string {
	return "country_roles"
}
