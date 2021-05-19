package models

import (
	"encoding/json"
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/fatih/structs"
	"github.com/google/uuid"
)

// Entity is an interface that has to be implemented by any type in order to be
// written in the database.
type Entity interface {

	// Key returns unique value an entity can be distinguished by.
	Key() string

	// Kind returns the name of entity's kind.
	Kind() string

	// Dependencies returns an array of all business sub-objects of the entity
	Dependencies() []config.Dependency

	// TableName is used by GORM to know which table to use.
	TableName() string
}

// Document is a complete representation of a CouchDB document and a GORM model.
type Document struct {
	ID          uuid.UUID             `json:"_id" validate:"required" gorm:"type:uuid; primary_key"`
	Deleted     bool                  `json:"_deleted,omitempty" gorm:"-"`
	Kind        string                `json:"type" validate:"required" gorm:"-"`
	Data        Entity                `json:"data" validate:"required" gorm:"embedded"`
	Timestamp   time.Time             `json:"timestamp" validate:"required" gorm:"-"`
	Attachments map[string]Attachment `json:"_attachments,omitempty" gorm:"-"`
}

// NewDocument creates a Document from a Typer.
//
// It generates ID and records timestamp of creation.
func NewDocument(e Entity) *Document {
	return &Document{
		ID:        uuid.New(),
		Kind:      e.Kind(),
		Data:      e,
		Timestamp: time.Now().UTC(),
	}
}

// Wrap Entity in a Documents. PostgreSQL needs this in order to
// mimic working with document-based database.
func Wrap(e Entity) *Document {
	json.Marshal(e)
	s := structs.New(e)
	return &Document{
		ID:        s.Field("ID").Value().(uuid.UUID),
		Kind:      e.Kind(),
		Data:      e,
		Timestamp: s.Field("UpdatedAt").Value().(time.Time),
	}

}

// Dependency stores an ID and Kind of an entity.
type Dependency = config.Dependency
