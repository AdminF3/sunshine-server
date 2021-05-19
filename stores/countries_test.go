package stores

import (
	"reflect"
	"sort"
	"testing"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/jinzhu/gorm"
)

type countries []models.Country

func (c countries) Len() int           { return len(c) }
func (c countries) Less(i, j int) bool { return c[i] < c[j] }
func (c countries) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func TestCountriesAreInSync(t *testing.T) {
	db := models.NewTestGORM(t)

	rows, err := db.Raw("SELECT unnest(enum_range(NULL::country))").Rows()
	if err != nil {
		t.Fatalf("Get countries from DB: %v", err)
	}

	inCode := countries(models.Countries())
	inDB := make(countries, 0, inCode.Len())
	defer rows.Close()
	for rows.Next() {
		var c models.Country
		if err := rows.Scan(&c); err != nil {
			t.Fatalf("country scan: %v", err)
		}
		inDB = append(inDB, c)
	}

	sort.Sort(inCode)
	sort.Sort(inDB)
	if !reflect.DeepEqual(inCode, inDB) {
		t.Error("Countries in code and DB are out of sync!")
		t.Errorf("In code:\t%v", inCode)
		t.Errorf("In DB:\t%v", inDB)
	}
}

func TestCountryCasesIsEnforced(t *testing.T) {
	db := models.NewTestGORM(t)

	stores := map[string]Store{
		"asset":        NewAssetStore(db, validate),
		"organization": NewOrganizationStore(db, validate),
		"project":      NewProjectStore(db, validate),
		"user":         NewUserStore(db, validate),
	}

	list := func(s Store, c models.Country) error {
		_, _, _, err := s.List(ctx, Filter{Country: c})
		return err
	}

	for k, s := range stores {
		t.Run(k, func(t *testing.T) {
			if err := list(s, models.CountryBulgaria); err != nil {
				t.Errorf("List by valid country: %v", err)
			}

			if err := list(s, "bulgaria"); err != nil {
				t.Errorf("List by country in lower case: %v", err)
			}

			if err := list(s, "BULGARIA"); err != nil {
				t.Errorf("List by country in upper case: %v", err)
			}

			if err := list(s, "BuLgArIa"); err != nil {
				t.Errorf("List by country with bad casing: %v", err)
			}

			// Querying by invalid enum value will rollback the transaction.
			db.Exec("BEGIN")
			defer func(db *gorm.DB) { db.Exec("ROLLBACK") }(db)
			if err := list(s, "Neverland"); err == nil {
				t.Errorf("List by bad country should fail, got nil instead")
			}
		})
	}
}
