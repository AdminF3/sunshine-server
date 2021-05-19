package stores

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
)

var validate = validator.New()

var _ Store = new(store)

type StoreTest struct {
	store  Store
	entity models.Entity

	// invalidate takes a valid entity and creates a new invalid from it
	invalidate func(models.Entity) models.Entity

	// postCreate executes after successful first create and fail the
	// test on non-nil returned error.
	postCreate func(models.Entity) error

	// update modifies given instance before updating
	update func(*models.Document) models.Entity

	// duplicate valid record which passes unique checks
	duplicate func(models.Entity) models.Entity

	// searchBy returns the search parameter specify by the API docs
	searchBy func(models.Entity) string

	// memberUUID update the entity and return its desire member's uuid:
	// for assets it is the owner
	// for organizations it is any user in the Roles
	// for projects it is any user in the Roles
	// if Entity does not have user - it is safe to return empty string
	memberUUID func(*testing.T, models.Entity) uuid.UUID

	// beforeSave is just a temporary work-around until !284 gets used by
	// the frontend and we could remove the legacy way of managing roles.
	beforeSave func(models.Entity)
}

func (st StoreTest) Name() string {
	return fmt.Sprintf("%s%s", strings.Title(st.entity.Kind()), "StoreTest")
}

func (st StoreTest) Run(t *testing.T) {
	var (
		updated *models.Document
		created *models.Document
		got     *models.Document
	)

	t.Run("Create", func(t *testing.T) {
		created = st.Create(t)
	})

	t.Run("GetAndUnwrap", func(t *testing.T) {
		if created == nil {
			t.Skipf("%s/Create did not pass", st.Name())
		}

		st.Get(t, created)
		st.Unwrap(t, created)
	})
	t.Run("GetByIndex", func(t *testing.T) {
		got = st.GetByIndex(t)
	})

	t.Run("Update", func(t *testing.T) {
		if got == nil {
			t.Skipf("%s/GetByIndex did not pass", st.Name())
		}

		updated = st.Update(t, got)
	})

	t.Run("List", func(t *testing.T) {
		if updated == nil {
			t.Skipf("%s/Update did not pass", st.Name())
		}
		st.List(t, updated)
	})

	t.Run("Search", func(t *testing.T) {
		if st.searchBy(st.entity) == "" {
			t.Skipf("%s is not searchable", st.entity.Kind())
		}
		if updated == nil {
			t.Skipf("%s/Update did not pass", st.Name())
		}
		st.Search(t, updated)
	})

	t.Run("Delete", func(t *testing.T) {
		st.Delete(t)
	})

	t.Run("ListByMember", func(t *testing.T) {
		if updated == nil {
			t.Skipf("%s/Update did not pass", st.Name())
		}
		if userID := st.memberUUID(t, updated.Data); userID != uuid.Nil {
			st.ListByMember(t, userID)
		}
	})
}

func (st StoreTest) cmpKindKey(t *testing.T, exp, got models.Entity) {
	if got.Kind() != exp.Kind() {
		t.Errorf("Data is not %s", strings.ToTitle(exp.Kind()))
	}

	if got.Key() != exp.Key() {
		t.Errorf("Expected key to be %q; got %q", exp.Key(), got.Key())
	}
}

func (st StoreTest) Create(t *testing.T) *models.Document {
	if st.beforeSave != nil {
		st.beforeSave(st.entity)
	}
	d, err := st.store.Create(ctx, st.entity)
	if err != nil {
		t.Fatalf("store.Create(ctx, %q) failed with %q", st.entity, err)
	}

	st.cmpKindKey(t, st.entity, d.Data)

	if err := st.postCreate(d.Data); err != nil {
		t.Error(err.Error())
	}

	_, err = st.store.Create(ctx, st.entity)
	if err == nil {
		t.Fatalf("store.Create(ctx, %q) should've failed as duplicate", st.entity)
	}

	invalid := st.invalidate(st.entity)
	if _, err := st.store.Create(ctx, invalid); err == nil {
		t.Fatalf("Expected error from store.Create(ctx, %s); got nil", invalid)
	}

	return d
}

func (st StoreTest) Get(t *testing.T, doc *models.Document) {
	got, err := st.store.Get(ctx, doc.ID)
	if err != nil {
		t.Fatalf("store.Get(ctx, %q) failed with %q", doc.ID, err)
	}

	st.cmpKindKey(t, st.entity, got.Data)
}

func (st StoreTest) Unwrap(t *testing.T, doc *models.Document) {
	got, deps, err := st.store.Unwrap(ctx, doc.ID)
	if err != nil {
		t.Fatalf("store.Unwrap(ctx, %q) failed with %q", doc.ID, err)
	}

	st.cmpKindKey(t, st.entity, got.Data)

	for _, d := range st.entity.Dependencies() {
		dep, ok := deps[d.ID]
		if !ok {
			continue
		}

		if dep.Kind != d.Kind {
			t.Errorf("Dependency %s is %q not %q", d.ID, dep.Kind, d.Kind)
		}
	}
}

func (st StoreTest) GetByIndex(t *testing.T) *models.Document {
	got, err := st.store.GetByIndex(ctx, st.entity.Key())
	if err != nil {
		t.Fatalf("store.GetByIndex(ctx, %q) failed with %q", st.entity.Key(), err)
	}

	st.cmpKindKey(t, st.entity, got.Data)

	return got
}

func (st StoreTest) Update(t *testing.T, doc *models.Document) *models.Document {
	copy := *doc
	changed := st.update(doc)

	copy.Data = changed
	if st.beforeSave != nil {
		st.beforeSave(st.entity)
	}
	updated, err := st.store.Update(ctx, &copy)
	if err != nil {
		t.Fatalf("store.Update(ctx, %v) failed with %q", doc, err)
	}

	st.cmpKindKey(t, changed, updated.Data)
	if reflect.DeepEqual(updated.Data, doc.Data) {
		t.Errorf("Expected update to change data, got the same instead")
	}

	copy = *updated
	copy.Data = st.invalidate(st.entity)
	if _, err := st.store.Update(ctx, &copy); err == nil {
		t.Fatalf("Expected error from store.Update(ctx, %v); got nil", updated)
	}

	return updated
}

func (st StoreTest) List(t *testing.T, updated *models.Document) {
	another := st.duplicate(updated.Data)

	if _, err := st.store.Create(ctx, another); err != nil {
		t.Errorf("Create failed: %s", err)
	}

	list, _, n, err := st.store.List(ctx, Filter{})
	if err != nil {
		t.Fatalf("store.List(ctx) failed with %q", err)
	}

	if len(list) > n {
		t.Errorf("Reported %d records, got %d instead", n, len(list))
	}

	if len(list) < 2 {
		t.Fatalf("Expected at least 2 documents, got %d", len(list))
	}

	if reflect.DeepEqual(list[0].Data, list[1].Data) {
		t.Fatalf("Got identical documents: \n\t%s, \n\t%s", list[0].Data, list[1].Data)
	}

	for _, doc := range list {
		if doc.Data.Kind() != st.entity.Kind() {
			t.Errorf("Expected user from List(); got %q", doc.Data.Kind())
		}
	}

	first, _, nf, err := st.store.List(ctx, Filter{Limit: 1})
	if err != nil {
		t.Fatalf("store.List(ctx, Filter{Limit: 1}) failed with %q", err)
	}
	if len(first) != 1 {
		t.Errorf("Expected one result with Filter{Limit: 1}")
	}

	if nf != n {
		t.Errorf("Got different total rows count %d than before %d", nf, n)
	}

	second, _, ns, err := st.store.List(ctx, Filter{Offset: 1})
	if err != nil {
		t.Fatalf("store.List(ctx, Filter{Offset: 1}) failed with %q", err)
	}
	if len(second)+1 != len(list) {
		t.Errorf("Expected one result less with Filter{Offset: 1}")
	}
	if ns != n {
		t.Errorf("Got different total rows count %d than before %d", ns, n)
	}

	if reflect.DeepEqual(first[0].Data, second[0].Data) {
		t.Fatalf("Got identical documents with filter: \n\t%s, \n\t%s",
			first[0].Data, second[1].Data)
	}
}

func (st StoreTest) ListByMember(t *testing.T, id uuid.UUID) {
	docs, _, n, err := st.store.ListByMember(ctx, Filter{Limit: 5}, id)

	if err != nil {
		t.Fatalf("store.ListByMember - failed with %q", err)
	}

	if len(docs) != 2 && len(docs) != int(n) {
		t.Errorf("Expected %v record for %q; got %d and x-doc-count: %d", 2, id, len(docs), n)
	}
}

func (st StoreTest) Search(t *testing.T, updated *models.Document) {
	another := st.duplicate(updated.Data)

	search, _, _, err := st.store.List(ctx, Filter{
		Search: st.searchBy(another),
	})

	if err != nil {
		t.Fatalf("store.List(ctx, Filter{Search: %q}) failed with %q", another.Kind(), err)
	}

	if len(search) != 1 {
		t.Fatalf("Expected one result but found: %d of kind %q", len(search), another.Kind())
	}

	// This part tests the search once more, with the search word changed to be
	// only uppercase and expects to have the same results as with the normal one
	searchUpper, _, _, errUpper := st.store.List(ctx, Filter{
		Search: strings.ToUpper(st.searchBy(another)),
	})
	if errUpper != nil {
		t.Fatalf("store.List(ctx, Filter{Search: %q}) failed with %q", another.Kind(), errUpper)
	}

	if len(searchUpper) != 1 {
		t.Fatalf("Expected one result but found: %d of kind %q", len(searchUpper), another.Kind())
	}
}

func (st StoreTest) Delete(t *testing.T) {
	list, _, n, err := st.store.List(ctx, Filter{})
	if err != nil || n == 0 || len(list) < 1 {
		t.Fatalf("List gave %d results, (x-count is %d) and err: %s", len(list), n, err)
	}

	err = st.store.Delete(ctx, &list[0])
	if err != nil {
		t.Fatalf("store.Delete(ctx, %v) failed with: %s", list[0], err)
	}

	got, err := st.store.Get(ctx, list[0].ID)
	if err == nil {
		t.Errorf("Record shouldn't be getable after delete; got: %#v", got)
	}
}
