package stores

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/fatih/structs"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

type store struct {
	// db model defined by the ORM layer.
	db *gorm.DB

	// validate is user-input validator.
	validate *validator.Validate

	// new creates new entity depending of the model.
	new func() models.Entity

	// index is unique field that it is not ID.
	index string

	// search query of the model or how you perform search for that entity.
	search func(*gorm.DB, Filter) *gorm.DB

	// member modifies the db instance (mostly adding necessary JOIN and
	// WHERE clauses) so that the parent entity to preload its members. The
	// point is to have the ability to list some enitities by its member/s.
	member func(*gorm.DB, ...uuid.UUID) *gorm.DB
}

func NewUserStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "email",
		new:      func() models.Entity { return new(models.User) },
		search: func(q *gorm.DB, f Filter) *gorm.DB {
			if f.Search != "" {
				q = q.Where("users.name ILIKE ?", pattern(f.Search)).
					Or("users.email ILIKE ?", pattern(f.Search))

			}

			if f.Status != 0 {
				q = q.Where("users.status = ?", f.Status)
			}

			if f.Country != "" {
				q = q.Where("users.country = ?", f.Country)
			}

			// NOTE (edimov): The `PlatformRoles` query works with OR condition
			// only. Consider implementing operators support in the filter.
			if f.PlatformRoles != nil {
				for _, v := range f.PlatformRoles {
					r := PlatformRole(v)
					if r != PlatformManager && r != AdminNetworkManager {
						continue
					}
					q = q.Or(fmt.Sprintf("users.%s = true", r))
				}
			}

			if f.CountryRoles != nil {
				q = q.Joins("left join country_roles on users.id = country_roles.user_id::UUID").
					Where("country_roles.role in (?)", f.CountryRoles)
			}

			return q
		},
		member: func(q *gorm.DB, _ ...uuid.UUID) *gorm.DB { return q },
	}
}

func NewContractStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "project_id",
		new:      func() models.Entity { return new(contract.Contract) },
		search: func(q *gorm.DB, f Filter) *gorm.DB {
			if f.Search != "" {
				q = q.Where("contracts.project_id = ?", pattern(f.Search))
			}

			return q
		},
		member: func(q *gorm.DB, _ ...uuid.UUID) *gorm.DB { return q },
	}
}

func NewOrganizationStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "vat",
		new:      func() models.Entity { return new(models.Organization) },
		search: func(q *gorm.DB, f Filter) *gorm.DB {
			if f.Search != "" {
				q = q.Where("organizations.name ilike ?", pattern(f.Search))
			}

			if f.LegalForm != 0 {
				q = q.Where("organizations.legal_form = ?", f.LegalForm)
			}

			if f.Status != 0 {
				q = q.Where("organizations.status = ?", f.Status)
			}

			if f.Country != "" {
				q = q.Where("organizations.country = ?", f.Country)
			}

			return q
		},
		member: func(q *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			return q.Joins(`left join organization_roles
				on organizations.id = organization_roles.organization_id`).
				Where("organization_roles.user_id IN (?)", ids)
		},
	}
}

func NewAssetStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "coords",
		new:      func() models.Entity { return new(models.Asset) },
		search: func(q *gorm.DB, f Filter) *gorm.DB {
			if f.Search != "" {
				q = q.Where("assets.address ILIKE ?", pattern(f.Search))
			}

			if f.BuildingType != 0 {
				q = q.Where("assets.building_type = ?", f.BuildingType)
			}

			if f.Status != 0 {
				q = q.Where("assets.status = ?", f.Status)
			}

			if f.Country != "" {
				q = q.Where("assets.country = ?", f.Country)
			}

			if f.Owner != uuid.Nil {
				q = q.Where("assets.owner_id = ?", f.Owner)
			}
			if f.ESCO != uuid.Nil {
				q = q.Where("assets.esco_id = ?", f.ESCO)
			}

			if f.NullFields != nil {
				for _, f := range f.NullFields {
					switch f {
					case "esco":
						q = q.Where("assets.esco_id is null")
					}
				}
			}

			return q
		},
		member: func(q *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			return q.Joins(`left join organizations
				on organizations.id = assets.owner_id::uuid
				or organizations.id = assets.esco_id::uuid`).
				Where("organizations.id IN (?)", ids)
		},
	}
}

func NewProjectStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "name",
		new:      func() models.Entity { return new(models.Project) },
		search: func(q *gorm.DB, f Filter) *gorm.DB {
			if f.Search != "" {
				q = q.Where("projects.name ILIKE ?", pattern(f.Search))
			}

			if f.Status != 0 {
				q = q.Where("projects.status = ?", f.Status)
			}

			if f.Country != "" {
				q = q.Where("projects.country = ?", f.Country)
			}

			if f.Owner != uuid.Nil {
				q = q.Or("projects.owner = ?", f.Owner)
			}

			if f.AssetOwner != uuid.Nil {
				q = q.Or("projects.asset_owner_id = ?", f.AssetOwner)
			}

			if f.ESCO != uuid.Nil {
				q = q.Or("projects.asset_esco_id = ?", f.ESCO)
			}

			if f.RelatedOrganizationID != uuid.Nil {
				q = q.Or("? = ANY(projects.consortium_orgs)", f.RelatedOrganizationID)
			}

			return q
		},
		member: func(q *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			q = q.Joins(`left join project_roles
				on project_roles.project_id = projects.id`).
				Where("project_roles.user_id IN (?)", ids)

			q = q.Joins(`left join organization_roles
				on organization_roles.organization_id = projects.owner
				or organization_roles.organization_id = projects.asset_esco_id
				or organization_roles.organization_id = ANY (projects.consortium_orgs::UUID[])
				or organization_roles.organization_id = projects.asset_owner_id`).
				Or("organization_roles.user_id IN (?)", ids)

			return q
		},
	}
}

func NewIndoorClimaStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "project_id",
		new:      func() models.Entity { return new(contract.IndoorClima) },
		search:   func(db *gorm.DB, s Filter) *gorm.DB { return db },
		member:   func(db *gorm.DB, ids ...uuid.UUID) *gorm.DB { return db },
	}
}

func NewMeetingsStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "id",
		new:      func() models.Entity { return new(models.Meeting) },
		search:   func(db *gorm.DB, f Filter) *gorm.DB { return db },
		member: func(db *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			if ids[0] != uuid.Nil {
				db = db.Where("host = ?", ids).
					Or("project = ?", ids)
			}

			return db
		},
	}
}

func NewGDPRStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "id",
		new:      func() models.Entity { return new(models.GDPRRequest) },
		search:   func(db *gorm.DB, f Filter) *gorm.DB { return db },
		member: func(db *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			if ids[0] != uuid.Nil {
				db = db.Where("id = ?", ids)
			}

			return db
		},
	}
}
func NewForfaitingApplicationStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "id",
		new:      func() models.Entity { return new(models.ForfaitingApplication) },
		search:   func(q *gorm.DB, f Filter) *gorm.DB { return q },
		member: func(q *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			return q.Joins(`left join projects
				on projects.id = forfaiting_applications.project_id`).
				Where("projects.id IN (?)", ids).
				Where("projects.deleted_at IS NULL")
		},
	}
}

func NewForfaitingPaymentStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "id",
		new:      func() models.Entity { return new(models.ForfaitingPayment) },
		search:   func(q *gorm.DB, f Filter) *gorm.DB { return q },
		member: func(q *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			return q.Joins(`left join projects
				on projects.id = forfaiting_payments.project_id`).
				Where("projects.id IN (?)", ids).
				Where("projects.deleted_at IS NULL")
		},
	}
}

func NewWorkPhaseStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "id",
		new:      func() models.Entity { return new(models.WorkPhase) },
		search:   func(q *gorm.DB, f Filter) *gorm.DB { return q },
		member: func(q *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			return q.Joins(`left join projects
				on projects.id = work_phase.project`).
				Where("projects.id IN (?)", ids).
				Where("projects.deleted_at IS NULL")
		},
	}
}

func NewMonitoringPhaseStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "id",
		new:      func() models.Entity { return new(models.MonitoringPhase) },
		search:   func(q *gorm.DB, f Filter) *gorm.DB { return q },
		member: func(q *gorm.DB, ids ...uuid.UUID) *gorm.DB {
			return q.Joins(`left join projects
				on projects.id = monitoring_phase.project`).
				Where("projects.id IN (?)", ids).
				Where("projects.deleted_at IS NULL")
		},
	}
}

func NewCountryStore(db *gorm.DB, v *validator.Validate) Store {
	return store{
		db:       db,
		validate: v,
		index:    "country",
		new:      func() models.Entity { return new(models.CountryVat) },
		search:   func(db *gorm.DB, s Filter) *gorm.DB { return db },
		member:   func(db *gorm.DB, ids ...uuid.UUID) *gorm.DB { return db },
	}
}

func (s store) DB() *gorm.DB { return s.db }

func (s store) FromKind(kind string) Store {
	var constructor func(*gorm.DB, *validator.Validate) Store

	switch kind {
	case "user":
		constructor = NewUserStore
	case "asset":
		constructor = NewAssetStore
	case "organization":
		constructor = NewOrganizationStore
	case "project":
		constructor = NewProjectStore
	case "contract":
		constructor = NewContractStore
	case "indoorclima":
		constructor = NewIndoorClimaStore
	case "meeting":
		constructor = NewMeetingsStore
	case "forfaiting_application":
		constructor = NewForfaitingApplicationStore
	case "work_phase":
		constructor = NewWorkPhaseStore
	case "monitoring_phase":
		constructor = NewMonitoringPhaseStore
	case "forfaiting_payment":
		constructor = NewForfaitingPaymentStore
	default:
		panic(fmt.Sprintf("No store for %s", kind))
	}

	return constructor(s.db, s.validate)
}

func (s store) Notifications() Notifier {
	return NewNotifier(s.db, s.validate)
}

func (s store) Create(ctx context.Context, e models.Entity) (*models.Document, error) {
	if err := s.validate.Struct(e); err != nil {
		return nil, err
	}

	if err := verifyDependencies(ctx, s, e); err != nil {
		return nil, err
	}

	err := s.db.Create(e).Error
	if err != nil {
		return nil, err
	}
	doc := models.Wrap(e)
	return doc, s.populateAttachments(doc)

}

func (s store) Delete(_ context.Context, d *models.Document) error {
	return s.db.Delete(d.Data).Error
}

func (s store) Get(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	var e = s.new()
	if err := s.db.Where(kv{"id": id}).First(e).Error; err != nil {
		return nil, WithID(err, id, e.Kind())
	}
	doc := models.Wrap(e)
	return doc, s.populateAttachments(doc)
}

func (s store) GetByIndex(ctx context.Context, value string) (*models.Document, error) {
	var e = s.new()
	if err := s.db.Where(kv{s.index: value}).First(e).Error; err != nil {
		return nil, WithIndex(err, s.index, value)
	}
	doc := models.Wrap(e)
	return doc, s.populateAttachments(doc)
}

func (s store) populateAttachments(doc *models.Document) error {
	var atts []models.Attachment
	if err := s.db.Where(kv{"owner_id": doc.ID}).Find(&atts).Error; err != nil {
		return WithID(err, doc.ID, "attachment")
	}

	if doc.Attachments == nil {
		doc.Attachments = make(map[string]models.Attachment)
	}
	for _, att := range atts {
		doc.Attachments[att.Name] = att
	}

	return nil
}

// execList does reflect magic in order to construct a slice (reflect.Value of
// Slice kind, actually) of the correct entity type and execute Find query
// storing the result there. Can not just pass []models.Entity as GORM will not
// know which table to query for these records and how to unmarshal them
// afterwards.
//
// The reflection value of a slice is then converted to a regular slice of
// models.Entity, sacrificing performance (both in terms of memory for the
// second slice and CPU time to actually type assert and copy records to it)
// for the sake of readability outside of this method.
func (s store) execList(ctx context.Context, q *gorm.DB) ([]models.Entity, error) {
	var (
		// Extract the type of s.new() and make a slice of it.
		t     = reflect.TypeOf(s.new())
		slice = reflect.MakeSlice(reflect.SliceOf(t), 0, 0)

		// Create a reflect Value pointing to a slice.
		x = reflect.New(slice.Type())
	)

	// Point the reflect value to the slice allocated above.
	x.Elem().Set(slice)

	// Execute query and convert this reflect madness to a regular []models.Entity.
	//
	// It doesn't make sense to check the error above, as if the query has
	// failed x.Len() will be zero and this next block would be a noop.
	err := q.Find(x.Interface()).Error
	x = x.Elem()
	var result = make([]models.Entity, x.Len())
	for i := 0; i < x.Len(); i++ {
		result[i] = x.Index(i).Interface().(models.Entity)
	}

	return result, err
}

type qfunc func(*gorm.DB) *gorm.DB

func (s store) list(ctx context.Context, f Filter, q qfunc) ([]models.Document, Dependencies, int, error) {
	rdb := f.GORM(s.db.Model(s.new()).Select(fmt.Sprintf("DISTINCT %q.*", s.new().TableName())))
	rdb = q(rdb)

	result, err := s.execList(ctx, rdb)
	if err != nil {
		return nil, nil, 0, err
	}

	var (
		c int
		m = new(sync.Map)
		e = new(sync.Map)

		wg sync.WaitGroup
	)

	docs := make([]models.Document, len(result))
	wg.Add(len(docs) + 1)
	go func() {
		defer wg.Done()
		c, err = s.count(ctx, q)
		if err != nil {
			e.Store("count", err)
		}
	}()
	for i, v := range result {
		d := models.Wrap(v)
		docs[i] = *d
		go func(deps []config.Dependency) {
			defer wg.Done()

			unwrap(ctx, s, deps, m, e)
		}(d.Data.Dependencies())
	}
	wg.Wait()

	for i, d := range docs {
		s.populateAttachments(&d)

		docs[i] = d
	}

	return docs, convertFromSyncMap(m), c, newErrorMap(e)
}

// UnwrapDeps preload the dependencies for the given entity.
func UnwrapDeps(st Store, entities []models.Document) (Dependencies, error) {
	var (
		m = new(sync.Map)
		e = new(sync.Map)

		wg sync.WaitGroup
	)

	wg.Add(len(entities))

	for _, v := range entities {
		d := v
		go func(deps []config.Dependency) {
			defer wg.Done()

			unwrap(ctx, st, deps, m, e)
		}(d.Data.Dependencies())
	}
	wg.Wait()

	return convertFromSyncMap(m), newErrorMap(e)
}

func (s store) count(ctx context.Context, q qfunc) (int, error) {
	cdb := s.db.Model(s.new()).Table(s.new().TableName()).
		Select(fmt.Sprintf("COUNT(DISTINCT %q.*)", s.new().TableName()))
	cdb = q(cdb)

	var count int

	return count, cdb.Row().Scan(&count)
}

func (s store) Count(ctx context.Context, f Filter) (int, error) {
	return s.count(ctx, func(q *gorm.DB) *gorm.DB {
		return s.search(q, f)
	})
}

func (s store) List(ctx context.Context, f Filter) ([]models.Document, Dependencies, int, error) {
	return s.list(ctx, f, func(q *gorm.DB) *gorm.DB {
		return s.search(q, f)
	})
}

func (s store) ListByMember(ctx context.Context, f Filter, ids ...uuid.UUID) ([]models.Document, Dependencies, int, error) {
	if len(ids) == 0 {
		return []models.Document{}, Dependencies{}, 0, nil
	}

	return s.list(ctx, f, func(q *gorm.DB) *gorm.DB {
		return s.member(s.search(q, f), ids...)
	})
}

func (s store) GetAttachment(ctx context.Context, doc *models.Document, filename string) (*models.Attachment, error) {
	var att models.Attachment
	err := s.db.Where(kv{"name": filename, "owner_id": doc.ID}).First(&att).Error
	if err != nil {
		WithIndex(err, filename, doc.ID.String())
	}
	return &att, err
}

func (s store) PutAttachment(ctx context.Context, _ *models.Document, att *models.Attachment) error {
	if err := s.validate.Struct(att); err != nil {
		return err
	}

	return s.db.Save(att).Error
}

func (s store) DeleteAttachment(ctx context.Context, doc *models.Document, filename string) error {
	att, err := s.GetAttachment(ctx, doc, filename)
	if err != nil {
		return err
	}

	return s.db.Delete(att).Error
}

func (s store) Unwrap(ctx context.Context, id uuid.UUID) (*models.Document, Dependencies, error) {
	return Unwrap(ctx, s, id)
}

func (s store) Update(ctx context.Context, d *models.Document) (*models.Document, error) {
	structs.New(d.Data).Field("Value").Field("ID").Set(d.ID)
	if err := s.validate.Struct(d.Data); err != nil {
		return nil, err
	}

	if err := verifyDependencies(ctx, s, d.Data); err != nil {
		return nil, err
	}

	err := s.db.Save(d.Data).Error
	if err != nil {
		return nil, err
	}
	doc := models.Wrap(d.Data)
	return doc, s.populateAttachments(doc)
}

// AtomicDelete tries to delete all given values in transaction and
// rolls it back on any error. Calling this on any store
// implementation other than psqlStore is a noop.
func AtomicDelete(s Store, values ...interface{}) error {
	ps, ok := s.(store)
	if !ok {
		return nil
	}

	tx := ps.db.Begin()
	defer tx.Commit()
	for _, value := range values {
		if err := tx.Delete(value).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

// kv is just an alias to shorten gorm.DB.Where calls.
type kv = map[string]interface{}

func pattern(v string) string {
	return "%" + v + "%"
}

func (s store) Portfolio() Portfolio {
	return NewPortfolioStore(s.db)
}
