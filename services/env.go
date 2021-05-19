package services

import (
	"encoding/base64"
	"fmt"
	"testing"

	"stageai.tech/sunshine/sunshine"
	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"

	raven "github.com/getsentry/raven-go"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

type Env struct {
	General           config.General
	Paths             config.Paths
	AssetStore        stores.Store
	ContractStore     stores.Store
	OrganizationStore stores.Store
	UserStore         stores.Store
	ProjectStore      stores.Store
	IndoorClimaStore  stores.Store
	MeetingsStore     stores.Store
	FAStore           stores.Store
	FPStore           stores.Store
	WPStore           stores.Store
	MPStore           stores.Store
	GDPRStore         stores.Store
	CountryStore      stores.Store
	Notifier          stores.Notifier
	Portfolio         stores.Portfolio
	SessionStore      sessions.Store
	TokenStore        stores.TokenStore
	Mailer            Mailer
	Validator         *validator.Validate
	Debug             bool
	DB                *gorm.DB
}

// NewSessionStore returns a new instance of session store with sensible
// options and proper auth and encrypt keys.
func NewSessionStore(cfg config.Session) (sessions.Store, error) {
	auth, err := base64.StdEncoding.DecodeString(cfg.Auth)
	if err != nil {
		return nil, err
	}
	enc, err := base64.StdEncoding.DecodeString(cfg.Encr)
	if err != nil {
		return nil, err
	}

	store := sessions.NewCookieStore(auth, enc)
	store.Options.Path = cfg.Path
	store.Options.HttpOnly = cfg.HTTPOnly
	store.Options.Secure = cfg.Secure
	store.Options.SameSite = cfg.SameSite
	return store, nil
}

// NewEnv returns a new ready to use environment with PostgreSQL.
func NewEnv() (*Env, error) {
	var (
		cfg      = config.Load()
		db, err  = models.NewGORM(cfg.DB)
		validate = constructValidator()

		sender SendMail
	)

	if err != nil {
		return nil, err
	}

	switch cfg.Mail.Backend {
	case "smtp":
		sender = Send
	case "file":
		sender = SendToFile
	default:
		return nil, fmt.Errorf("mail backend %q not implemented", cfg.Mail.Backend)
	}

	sessionStore, err := NewSessionStore(cfg.Session)
	if err != nil {
		return nil, err
	}

	raven.SetRelease(sunshine.Version())
	return &Env{
		General:           cfg.General,
		Paths:             cfg.Paths,
		AssetStore:        stores.NewAssetStore(db, validate),
		ContractStore:     stores.NewContractStore(db, validate),
		OrganizationStore: stores.NewOrganizationStore(db, validate),
		UserStore:         stores.NewUserStore(db, validate),
		IndoorClimaStore:  stores.NewIndoorClimaStore(db, validate),
		MeetingsStore:     stores.NewMeetingsStore(db, validate),
		WPStore:           stores.NewWorkPhaseStore(db, validate),
		MPStore:           stores.NewMonitoringPhaseStore(db, validate),
		Notifier:          stores.NewNotifier(db, validate),
		Portfolio:         stores.NewPortfolioStore(db),
		GDPRStore:         stores.NewGDPRStore(db, validate),
		CountryStore:      stores.NewCountryStore(db, validate),
		SessionStore:      sessionStore,
		ProjectStore:      stores.NewProjectStore(db, validate),
		TokenStore:        stores.NewTokenStore(db, validate),
		FAStore:           stores.NewForfaitingApplicationStore(db, validate),
		FPStore:           stores.NewForfaitingPaymentStore(db, validate),
		Mailer:            NewMailer(cfg.General, cfg.Mail, sender),
		Validator:         validate,
		DB:                db,
	}, raven.SetDSN(cfg.General.SentryDSN)
}

// NewTestEnv returns a new ready to use PostgreSQL environment and a delete func.
func NewTestEnv(t *testing.T) *Env {
	var (
		cfg      = config.Load()
		db       = models.NewTestGORM(t)
		validate = constructValidator()
	)

	sessionStore, err := NewSessionStore(cfg.Session)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range models.Countries() {
		if c.IsConsortium() {
			stores.NewTestPortfolioRole(t, stores.NewUserStore(db, validate), models.PortfolioDirectorRole, c)
			stores.NewTestPortfolioRole(t, stores.NewUserStore(db, validate), models.DataProtectionOfficerRole, c)
			stores.NewTestPortfolioRole(t, stores.NewUserStore(db, validate), models.CountryAdminRole, c)
		}
	}

	return &Env{
		General:           cfg.General,
		Paths:             cfg.Paths,
		AssetStore:        stores.NewAssetStore(db, validate),
		ContractStore:     stores.NewContractStore(db, validate),
		OrganizationStore: stores.NewOrganizationStore(db, validate),
		UserStore:         stores.NewUserStore(db, validate),
		IndoorClimaStore:  stores.NewIndoorClimaStore(db, validate),
		MeetingsStore:     stores.NewMeetingsStore(db, validate),
		Notifier:          stores.NewNotifier(db, validate),
		Portfolio:         stores.NewPortfolioStore(db),
		WPStore:           stores.NewWorkPhaseStore(db, validate),
		MPStore:           stores.NewMonitoringPhaseStore(db, validate),
		SessionStore:      sessionStore,
		ProjectStore:      stores.NewProjectStore(db, validate),
		TokenStore:        stores.NewTokenStore(db, validate),
		FAStore:           stores.NewForfaitingApplicationStore(db, validate),
		FPStore:           stores.NewForfaitingPaymentStore(db, validate),
		GDPRStore:         stores.NewGDPRStore(db, validate),
		CountryStore:      stores.NewCountryStore(db, validate),
		Mailer:            NewMailer(cfg.General, cfg.Mail, SendToFile),
		Debug:             true,
		Validator:         validate,
		DB:                db,
	}
}

func constructValidator() *validator.Validate {
	var result = validator.New()

	result.RegisterValidation("upload_type", func(fl validator.FieldLevel) bool {
		_, ok := models.UploadTypes[fl.Field().String()]
		return ok
	})

	return result
}
