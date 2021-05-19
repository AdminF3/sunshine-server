package stores

import (
	"context"
	"fmt"
	"sync"

	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Store takes care for reading and writing documents to a database. Records
// are types implementing models.Entity which get wrapped in a models.Document
// value.
//
// Store values are to be created for each entity kind and passing entity of
// the wrong kind is considered as undefined behaviour. Each implementation
// might handle it internally, fail or even panic.
type Store interface {
	// Create new record of given Entity. The result is wrapped in a Document.
	Create(context.Context, models.Entity) (*models.Document, error)

	// Get document by given unique ID.
	Get(ctx context.Context, id uuid.UUID) (*models.Document, error)

	// GetByIndex fetches document from the database by given index value.
	GetByIndex(ctx context.Context, value string) (*models.Document, error)

	// Unwrap behaves like GetByIndex but preloads all dependencies.
	Unwrap(ctx context.Context, id uuid.UUID) (*models.Document, Dependencies, error)

	// Update document, matched by ID.
	Update(context.Context, *models.Document) (*models.Document, error)

	// Delete given document. Store implementations might do soft-deletion.
	Delete(context.Context, *models.Document) error

	// List all documents matching Filter. This method preloads dependencies.
	List(context.Context, Filter) ([]models.Document, Dependencies, int, error)

	// Count returns the number of all documents matching Filter.
	Count(context.Context, Filter) (int, error)

	// ListByMember lists documents where Filter apply and any of ids is
	// member of. This method preloads dependencies.
	ListByMember(ctx context.Context, filter Filter, ids ...uuid.UUID) ([]models.Document, Dependencies, int, error)

	// GetAttachment retreives attachment of doc by given filename.
	GetAttachment(ctx context.Context, doc *models.Document, filename string) (*models.Attachment, error)

	// PutAttachment uploads attachment of doc.
	PutAttachment(context.Context, *models.Document, *models.Attachment) error

	// DeleteAttachment deletes an attachment from a doc.
	DeleteAttachment(ctx context.Context, doc *models.Document, filename string) error

	// FromKind creates a new Store based on given kind as would've been
	// returned from calling models.Entity.Kind().
	//
	// It panics when called with non-existent kind.
	FromKind(string) Store

	// Notifications creates a new Notifier from any given store.
	Notifications() Notifier

	// Portfolio returns instance of store to interact with
	// portfolio DB.
	Portfolio() Portfolio

	// DB returns underlying *gorm.DB.
	DB() *gorm.DB
}

func verifyDependencies(ctx context.Context, s Store, e models.Entity) error {
	for _, dep := range e.Dependencies() {
		doc, err := s.FromKind(dep.Kind).Get(ctx, dep.ID)
		if err != nil {
			return fmt.Errorf("dependency error for %v: %w", e.Kind(), err)
		}

		if doc.Kind != dep.Kind {
			return fmt.Errorf("got dependency for kind %q, exp %q", doc.Kind, dep.Kind)
		}
	}
	return nil
}

// convertFromSyncMap converts sync.Map to map[string]*models.Document. If
// there are keys and/or values that don't match the respective types this
// simply ignores them and moves forward.
func convertFromSyncMap(m *sync.Map) Dependencies {
	var r = make(Dependencies)

	m.Range(func(key interface{}, value interface{}) bool {
		var (
			k, kok = key.(uuid.UUID)
			d, dok = value.(*models.Document)
		)

		if kok && dok {
			r[k] = d
		}
		return true
	})

	return r
}

// Dependencies is map of documents of at least one models.Document.
type Dependencies map[uuid.UUID]*models.Document

// ShouldInvalidate reports whether document's data is outdated.
func ShouldInvalidate(old, new *models.Document, isAdmin bool, oldValid models.ValidationStatus) bool {
	if isAdmin {
		// TODO: Maybe fire to Sentry or notification to someone here to avoid shenanigans
		return false
	}
	switch t := old.Data.(type) {
	case *models.User:
		user := new.Data.(*models.User)
		if t.Name != user.Name ||
			t.Identity != user.Identity ||
			t.Address != user.Address {
			return true
		}
		new.Data.(*models.User).Valid = oldValid
	case *models.Organization:
		org := new.Data.(*models.Organization)
		if t.Name != org.Name ||
			t.Address != org.Address ||
			t.LegalForm != org.LegalForm ||
			t.VAT != org.VAT ||
			t.RegistrationNumber != org.RegistrationNumber {
			return true
		}
		new.Data.(*models.Organization).Valid = oldValid
	case *models.Asset:
		asset := new.Data.(*models.Asset)
		if t.Owner != asset.Owner ||
			t.Address != asset.Address ||
			t.Coordinates != asset.Coordinates ||
			t.Cadastre != asset.Cadastre ||
			t.Category != asset.Category {
			return true
		}
		new.Data.(*models.Asset).Valid = oldValid
	case *models.Project:
		prj := new.Data.(models.Project)
		if t.Name != prj.Name ||
			t.AirTemperature != prj.AirTemperature ||
			t.WaterTemperature != prj.WaterTemperature ||
			t.GuaranteedSavings != prj.GuaranteedSavings ||
			t.ContractTerm != prj.ContractTerm ||
			t.FirstYear != prj.FirstYear {
			return true
		}
	}
	return false
}

func Unwrap(ctx context.Context, s Store, id uuid.UUID) (*models.Document, Dependencies, error) {
	var (
		err error
		m   = new(sync.Map)
		e   = new(sync.Map)
	)

	doc, err := s.Get(ctx, id)
	if err != nil {
		return doc, nil, err
	}

	m.Store(id, doc)

	unwrap(ctx, s, doc.Data.Dependencies(), m, e)
	included := convertFromSyncMap(m)
	delete(included, id)

	return doc, included, newErrorMap(e)
}

func unwrap(ctx context.Context, s Store, deps []config.Dependency, m, e *sync.Map) {
	if len(deps) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(deps))

	for _, dep := range deps {
		go func(dep config.Dependency) {
			defer wg.Done()

			if dep.ID == uuid.Nil {
				return
			}

			if _, ok := m.LoadOrStore(dep.ID, nil); ok {
				// this is already (to be) fetched.
				return
			}
			store := s.FromKind(dep.Kind)
			doc, err := store.Get(ctx, dep.ID)
			if err != nil {
				e.Store(dep.ID, err)
				return
			}

			m.Store(dep.ID, doc)
			unwrap(ctx, store, doc.Data.Dependencies(), m, e)
		}(dep)
	}

	wg.Wait()
}
