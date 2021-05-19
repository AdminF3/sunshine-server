package models

import (
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/google/uuid"
)

type ForfaitingApplication struct {
	Value

	Project     uuid.UUID   `validate:"required" gorm:"column:project_id"`
	Reviews     []FAReview  `gorm:"foreignkey:ForfaitingApplicationID;PRELOAD:true"`
	BankAccount BankAccount `gorm:"foreignkey:FAID;PRELOAD:true"`
	PrivateBond bool
	Finance     Finance
	ManagerID   uuid.UUID `gorm:"column:manager_id"`
	Manager     User
}

func (ForfaitingApplication) TableName() string { return "forfaiting_applications" }
func (ForfaitingApplication) Kind() string      { return "forfaiting_application" }
func (fa ForfaitingApplication) Key() string    { return fa.ID.String() }
func (fa ForfaitingApplication) Dependencies() []Dependency {
	var dep = []config.Dependency{
		{ID: fa.ManagerID, Kind: "user"},
	}
	return dep
}

type FAReview struct {
	Value

	ForfaitingApplicationID uuid.UUID
	Author                  *uuid.UUID
	Approved                bool
	Comment                 string
	Type                    FAReviewType
}

func (FAReview) TableName() string          { return "fa_reviews" }
func (FAReview) Kind() string               { return "fa_review" }
func (fa FAReview) Key() string             { return fa.ID.String() }
func (FAReview) Dependencies() []Dependency { return nil }

type Finance int8

const (
	_ Finance = iota
	FinanceEquity
	FinanceBankFunding
	FinanceOther
)

type BankAccount struct {
	Value

	BeneficiaryName string
	BankNameAddress string
	IBAN            string
	SWIFT           string

	FAID uuid.UUID
}

func (BankAccount) TableName() string          { return "bank_accounts" }
func (BankAccount) Kind() string               { return "bank_account" }
func (ba BankAccount) Key() string             { return ba.ID.String() }
func (BankAccount) Dependencies() []Dependency { return nil }

type FAReviewType int8

const (
	_ FAReviewType = iota
	FAReviewTypeFinancial
	FAReviewTypeTechnical
	FAReviewTypeGuidelines
	FAReviewTypeExecutive
)

type ForfaitingPayment struct {
	Value

	TransferValue int
	Currency      Currency
	Project       uuid.UUID `validate:"required" gorm:"column:project_id"`
	TransferDate  time.Time
}

func (ForfaitingPayment) TableName() string          { return "forfaiting_payments" }
func (ForfaitingPayment) Kind() string               { return "forfaiting_payment" }
func (fp ForfaitingPayment) Key() string             { return fp.ID.String() }
func (ForfaitingPayment) Dependencies() []Dependency { return nil }
func (ForfaitingPayment) IsEntity()                  {}

type Currency string

const (
	CurrencyEUR Currency = "EUR"
	CurrencyALL Currency = "ALL"
	CurrencyAMD Currency = "AMD"
	CurrencyBYN Currency = "BYN"
	CurrencyBAM Currency = "BAM"
	CurrencyBGN Currency = "BGN"
	CurrencyHRK Currency = "HRK"
	CurrencyCZK Currency = "CZK"
	CurrencyDKK Currency = "DKK"
	CurrencyGEL Currency = "GEL"
	CurrencyHUF Currency = "HUF"
	CurrencyISK Currency = "ISK"
	CurrencyCHF Currency = "CHF"
	CurrencyMDL Currency = "MDL"
	CurrencyMKD Currency = "MKD"
	CurrencyNOK Currency = "NOK"
	CurrencyPLN Currency = "PLN"
	CurrencyRON Currency = "RON"
	CurrencyRUB Currency = "RUB"
	CurrencyRSD Currency = "RSD"
	CurrencySEK Currency = "SEK"
	CurrencyTRY Currency = "TRY"
	CurrencyUAH Currency = "UAH"
	CurrencyGBP Currency = "GBP"
)
