package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// Notification is a general description for any of the Actions we notify for.
type Notification struct {
	Value

	// Action to be notified about.
	Action UserAction `json:"action"`

	// Recipient should receive this notification.
	RecipientID uuid.UUID `json:"recipient_id" gorm:"column:recipient"`

	// User is the ID of the one who made the action.
	UserID uuid.UUID `json:"user"`

	// UserKey is the Key() value for the given user record (email).
	UserKey string `json:"user_key"`

	// Target is the ID of a non-user entity that made the action.
	TargetID uuid.UUID `json:"target"`

	// TargetKey is the Key() value of the target entity record.
	TargetKey string `json:"target_key"`

	// TargetType is the type of entity that made the action.
	TargetType EntityType `json:"target_type"`

	// Old value if changed by the action.
	Old string `json:"old"`

	// New value if changed by the action.
	New string `json:"new"`

	Seen bool `json:"seen"`

	// Country of the target entity.
	Country Country `json:"country"`

	// Comment is only used for approving document notifications.
	Comment string `json:"comment"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (et *EntityType) Scan(value interface{}) error {
	var v, ok = value.([]byte)
	if !ok {
		return fmt.Errorf("invalid entity type: %v", v)
	}

	*et = EntityType(v)
	return nil
}

func (et EntityType) Value() (driver.Value, error) {
	if len(et) == 0 {
		return nil, nil
	}
	return string(et), nil
}

func (ua *UserAction) Scan(value interface{}) error {
	var v, ok = value.([]byte)
	if !ok {
		return fmt.Errorf("invalid user action type: %v", v)
	}

	*ua = UserAction(v)
	return nil
}

func (ua UserAction) Value() (driver.Value, error) {
	if len(ua) == 0 {
		return nil, nil
	}
	return string(ua), nil
}
