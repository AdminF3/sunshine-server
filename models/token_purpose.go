package models

import (
	"database/sql/driver"
	"fmt"
)

type TokenPurpose string

const (
	SessionToken       TokenPurpose = "session"
	CreateToken        TokenPurpose = "create"
	ResetPwdToken      TokenPurpose = "resetpwd"
	CreateProjectToken TokenPurpose = "createprj"
)

// Scan implements the database/sql.Scanner interface.
func (p *TokenPurpose) Scan(value interface{}) error {
	var v, ok = value.([]byte)
	if !ok {
		return fmt.Errorf("invalid token purpose: %v", value)
	}

	*p = TokenPurpose(v)
	return nil
}

// Value implements the database/sql/driver.Valuer interface.
func (p TokenPurpose) Value() (driver.Value, error) {
	return string(p), nil
}

func (p TokenPurpose) String() string {
	return string(p)
}
