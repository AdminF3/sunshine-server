package models

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/DATA-DOG/go-txdb"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Value is basic model definition, which includes fields ID, CreatedAt,
// UpdatedAt, DeletedAt. It's intended to be embedded in your model type in
// order to pass them to Store.
//
// For instance:
//
//	type User struct {
//		store.Value
//
//		Name string
//		Email string
//		Password string
//		// ...
//	}
type Value struct {
	// ID value is unique for each record, initial value is set by the
	// database and should never be mutated.
	ID uuid.UUID `gorm:"type:uuid; primary_key"`

	// CreatedAt keeps when record has been created. Initial value gets
	// automatically set by the database and should never be mutated.
	CreatedAt time.Time

	// UpdatedAt keeps when was the list time record has been updated. Upon
	// creation initial value is automatically set by the database and GORM
	// automatically updates it on save.
	UpdatedAt time.Time

	// DeletedAt keeps when record has been deleted. nil value means that
	// record is NOT deleted.
	DeletedAt *time.Time
}

const fmtConn = "host='%s' port='%d' dbname='%s' search_path='%s' sslmode=disable"

var (
	logsql   = flag.Bool("logsql", false, "log sql queries while testing")
	gooseDir = flag.String(
		"goose",
		"",
		"custom location of the migrations; usefull when debugging to set full path for goose")
)

func NewGORM(cfg config.DB) (*gorm.DB, error) {
	var db, err = gorm.Open("postgres", pgConnectString("public", cfg))
	return configGORM(db), err
}

// configGORM explicitly sets some poorly documented GORM settings even though
// their default values might match.
func configGORM(db *gorm.DB) *gorm.DB {
	return db.
		Set("gorm:auto_preload", true).
		Set("gorm:association_autocreate", true).
		Set("gorm:association_autoupdate", true).
		Set("gorm:association_save_reference", true).
		Set("gorm:save_associations", true).
		LogMode(false)
}

func setupTestSchema(t *testing.T) string {
	cfg := config.Load()
	path, _ := PathName(t)
	schema := fmt.Sprintf("test_%s", uuid.New())
	mpath := filepath.Join(path, "models", "migrations")
	pconn := pgConnectString(schema, cfg.DB)

	db, err := gorm.Open("postgres", pconn)
	if err != nil {
		t.Fatalf("Failed to open a connection: %s", err)
	}
	defer db.Close()

	if _, err = db.Raw(fmt.Sprintf("CREATE SCHEMA %q", schema)).Rows(); err != nil {
		t.Fatalf("Create schema failed: %s", err)
	}
	dbSchemas.Store(schema, struct{}{})

	if *gooseDir != "" {
		mpath = *gooseDir
	}

	goose := exec.Command("goose", "-dir", mpath, "postgres", pconn, "up")
	if *logsql {
		goose.Stderr = os.Stderr
		goose.Stdout = os.Stdout
	}
	if err = goose.Run(); err != nil {
		t.Fatalf("Running goose failed: %q", err)
	}

	return pconn
}

func mustOpen(t *testing.T, args ...interface{}) *gorm.DB {
	db, err := gorm.Open("txdb", args...)
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return configGORM(db).LogMode(*logsql)
}

var (
	dbSchemas sync.Map
	dbSetup   sync.Once
	dsn       string
)

// NewTestGORM makes sure to open a transaction in a test schema with executed
// migrations and no data.
//
// Note that this function doesn't take care of cleaning up those test schemas.
// Users must call `ClearTestSchemas` at end of the whole test program.
func NewTestGORM(t *testing.T) *gorm.DB {
	dbSetup.Do(func() {
		dsn = setupTestSchema(t)
		txdb.Register("txdb", "postgres", dsn)

		d, ok := gorm.GetDialect("postgres")
		if !ok {
			t.Fatal(`You must import "github.com/jinzhu/gorm/dialects/postgres"`)
		}
		gorm.RegisterDialect("txdb", d)
	})
	db := mustOpen(t, "txdb", dsn)

	var txStatus string
	err := db.DB().QueryRow("SELECT txid_status(txid_current())").Scan(&txStatus)
	if err != nil || txStatus != "in progress" {
		t.Logf("Bad tx status %q (%v). Creating a new one.", txStatus, err)
		db = mustOpen(t, "txdb", setupTestSchema(t))
	}

	t.Cleanup(func() { db.Close() })
	return db
}

// ClearTestSchemas removes any schemas created by NewGORM. The functions is
// intended to be called from TestMain.
//
// Example:
//	func TestMain(m *testing.M) {
//		statusCode := m.Run()
//		if err := models.ClearTestSchemas(); err != nil {
//			fmt.Printf("Clear test schemas: %v", err)
//			statusCode = 1
//		}
//		os.Exit(statusCode)
//	}
func ClearTestSchemas() error {
	cfg := config.Load()
	schema := fmt.Sprintf("test_%s", uuid.New())
	pconn := pgConnectString(schema, cfg.DB)

	db, err := gorm.Open("postgres", pconn)
	if err != nil {
		return fmt.Errorf("gorm.Open: %w", err)
	}
	defer db.Close()

	dbSchemas.Range(func(key, _ interface{}) bool {
		if err := db.Exec(fmt.Sprintf("DROP SCHEMA %q CASCADE", key)).Error; err != nil {
			fmt.Printf("Failed to drop %q: %v. DROP IT MANUALLY!\n", key, err)
		}
		return true
	})
	return nil
}

// PathName returns base project path and project's name.
//
// This works reliably because Go guarantees that tests are being run from
// current file's directory.
func PathName(t *testing.T) (string, string) {
	var dirname, err = os.Getwd()
	if err != nil {
		t.Fatalf("base: %s", err)
	}

	prjPath := filepath.Join(dirname, "..")
	return prjPath, filepath.Base(prjPath)
}

func pgConnectString(schema string, cfg config.DB) string {
	var v = fmt.Sprintf(fmtConn, cfg.Host, cfg.Port, cfg.Name, schema)
	if cfg.Username != "" {
		v = fmt.Sprintf("%s user='%s'", v, cfg.Username)
	}
	if cfg.Password != "" {
		v = fmt.Sprintf("%s password='%s'", v, cfg.Password)
	}
	return v
}
