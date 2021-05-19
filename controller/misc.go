package controller

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"regexp"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"
)

type RequestForm struct {
	FileHeader *multipart.FileHeader
	File       multipart.File
	UploadType string
	Kind       string
	Comment    string
}

// RequestFormComment is a structured comment.
type RequestFormComment struct {
	OrganizationID   uuid.UUID  `json:"organizationID"`
	OrganizationName string     `json:"organizationName"`
	AssetID          uuid.UUID  `json:"assetID"`
	AssetAddress     string     `json:"assetAddress"`
	AssetOwnerID     uuid.UUID  `json:"assetOwnerID"`
	AssetEscoID      *uuid.UUID `json:"assetEscoID"`
	Country          string     `json:"country"`
}

// RequestCommentOrganization retrieves organization document
// deduced from request's comment.
func (rf RequestForm) RequestCommentOrganization() (*models.Document, error) {
	c := &RequestFormComment{}
	err := json.Unmarshal([]byte(rf.Comment), c)
	if err != nil {
		return nil, err
	}

	v := models.Value{ID: c.OrganizationID}
	o := models.Organization{
		Value: v,
		Name:  c.OrganizationName,
	}

	return &models.Document{
		ID:   c.OrganizationID,
		Kind: o.Kind(),
		Data: &o,
	}, nil
}

// RequestCommentAsset retrieves asset document
// deduced from request's comment.
func (rf RequestForm) RequestCommentAsset() (*models.Document, error) {
	c := &RequestFormComment{}
	err := json.Unmarshal([]byte(rf.Comment), c)
	if err != nil {
		return nil, err
	}

	v := models.Value{ID: c.AssetID}
	a := models.Asset{
		Value:   v,
		Address: c.AssetAddress,
		Owner:   c.AssetOwnerID,
		ESCO:    c.AssetEscoID,
		Country: models.Country(c.Country),
	}

	return &models.Document{
		ID:   c.AssetID,
		Kind: a.Kind(),
		Data: &a,
	}, nil
}

func notifyAll(ctx context.Context, st stores.Notifier, ids []uuid.UUID, not models.Notification) {
	for _, a := range ids {
		not.RecipientID = a
		go st.Notify(ctx, &not)
	}
}

type alterRole struct {
	Position string    `json:"position"`
	User     uuid.UUID `json:"user"`
}

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("record not found")
	ErrBadInput     = errors.New("bad input")
	ErrFatal        = errors.New("internal system failure")
	ErrInvalidTable = errors.New("invalid table")
	ErrDuplicate    = errors.New("duplicate entry")
)

var CheckFilenameRe = regexp.MustCompile("[a-zA-Z0-9._-]+")

func setUniqueOrgRoles(roles []models.OrganizationRole) []models.OrganizationRole {
	unique := make(map[models.OrganizationRole]struct{}, len(roles))
	for _, m := range roles {
		unique[m] = struct{}{}
	}

	updated := make([]models.OrganizationRole, 0, len(unique))
	for k := range unique {
		updated = append(updated, k)
	}

	return updated
}

func sanitizeInputFields(doc *models.Document) *models.Document {
	doc.Data = sanitizeEntityFields(doc.Data)
	return doc
}

func sanitizeEntityFields(e models.Entity) models.Entity {
	p := bluemonday.StrictPolicy()
	switch t := e.(type) {
	case *models.User:
		p.Sanitize(t.Name)
		p.Sanitize(t.Email)
		p.Sanitize(t.Address)
	case *models.Organization:
		p.Sanitize(t.Name)
		p.Sanitize(t.VAT)
		p.Sanitize(t.Address)
		p.Sanitize(t.Website)
		p.Sanitize(t.Email)
		p.Sanitize(t.Telephone)
	case *models.Asset:
		p.Sanitize(t.Address)
		p.Sanitize(t.Cadastre)
	case *models.Project:
		p.Sanitize(t.Name)
	}
	return e
}

func getANWManager(s stores.Store) uuid.UUID {
	var u models.User
	err := s.DB().Where("admin_network_manager = TRUE").First(&u).Error
	if err != nil {
		panic("There is no admin network manager in the platform")
	}
	return u.ID
}
