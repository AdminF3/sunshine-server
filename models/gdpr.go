package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/config"
)

type GDPRRequest struct {
	Value
	UserID           uuid.UUID `json:"user_id"`
	RequesterName    string    `json:"requester_name"`
	RequesterPhone   string    `json:"requester_phone"`
	RequesterEmail   string    `json:"requester_email"`
	RequesterAddress string    `json:"requester_address"`
	Name             string    `json:"name"`
	Phone            string    `json:"phone"`
	Email            string    `json:"email"`
	Address          string    `json:"address"`
	Action           GDPRType  `json:"action"`
	Reason           string    `json:"reason"`
	Information      string    `json:"information"`
}

func (GDPRRequest) TableName() string {
	return "gdpr_requests"
}

func (GDPRRequest) IsEntity() {}

func (t *GDPRType) Scan(value interface{}) error {
	var v, ok = value.([]byte)
	if !ok {
		return fmt.Errorf("invalid GDPR type: %v", v)
	}

	*t = GDPRType(v)
	return nil
}

func (t GDPRType) Value() (driver.Value, error) {
	if len(t) == 0 {
		return nil, nil
	}
	return string(t), nil
}

func (g GDPRRequest) Dependencies() []config.Dependency {
	return []config.Dependency{}
}

func (GDPRRequest) Kind() string {
	return "gdpr_request"
}

func (g GDPRRequest) Key() string {
	return g.ID.String()
}
