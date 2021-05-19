package models

import (
	"time"

	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
)

const timeDay = 24 * time.Hour

// makes it easier to mock it.
var now = time.Now

// Token is an immutable value giving limited time credentials to given entity
// for a given resource.
//
// Supported resources:
//
//	session		Session token is to be given on successful login
//	create		Create token is to be given upon registering new user
//	resetpwd	Reset password token is to be given when user initially
//			reports about forgotten password.
//      createprj       Create project token permits a guest organization to
//                      create a project for host organization's asset.
type Token struct {
	ID      uuid.UUID     `gorm:"primary_key"`
	Purpose TokenPurpose  `validate:"required" gorm:"column:purpose"`
	UserID  uuid.UUID     `validate:"required" gorm:"column:user_id"`
	User    User          `validate:"-"`
	TTL     time.Duration `validate:"required" gorm:"column:ttl"`

	CreatedAt time.Time
}

func (Token) TableName() string {
	return "tokens"
}

// NewToken creates new token value for given resource to entity with given id.
func NewToken(purpose TokenPurpose, id uuid.UUID) *Token {
	var t = Token{
		Purpose: purpose,
		UserID:  id,
	}

	switch purpose {
	case SessionToken:
		t.TTL = 30 * timeDay
	case CreateToken:
		t.TTL = 48 * time.Hour
	case ResetPwdToken:
		t.TTL = 24 * time.Hour
	case CreateProjectToken:
		t.TTL = 7 * timeDay
	default:
		t.TTL = 0
	}

	return &t
}

// ValidToken returns true when given document is a valid token.
//
// Token is invalid if it has ever been edited or has expired.
func ValidToken(t Token, purpose TokenPurpose, validate *validator.Validate) bool {
	// has it expired?
	if t.CreatedAt.Add(t.TTL).Before(now()) {
		return false
	}

	// cannot reuse the same token for different resources.
	if t.Purpose != purpose {
		return false
	}

	return validate.Struct(t) == nil
}
