// Package contract is responsible for creating contracts, performing calculation over them and generaing PDF documents.
package contract

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// Make sure we can initialize a contract.
var _ = New(uuid.New())

// JSONMap is a map that implements sql.Scanner and driver.Valuer.
type JSONMap map[string]string

// Contract holds general fields, agreement fields, raw markdown and tables
// data of a contract for given project.
type Contract struct {
	models.Value

	Project     uuid.UUID       `json:"project" validate:"required" gorm:"column:project_id"`
	Fields      JSONMap         `json:"fields"`
	Agreement   JSONMap         `json:"agreement_fields"`
	Markdown    []byte          `json:"markdown"`
	Tables      Tables          `json:"tables"`
	Maintenance JSONMap         `json:"maintenance"`
	Eurobor     decimal.Decimal `json:"-" gorm:"-"`
	VAT         float64         `json:"-" gorm:"-"`
}

func (c *Contract) AfterFind(tx *gorm.DB) (err error) {
	c.Eurobor = GetEurobor(tx)

	var proj models.Project
	if err = tx.Where("id = ?", c.Project).First(&proj).Error; err != nil {
		return err
	}
	c.VAT = GetVat(tx, proj.Country)
	return nil
}

func (c *Contract) BeforeSave(tx *gorm.DB) (err error) {
	var proj models.Project
	if err = tx.Where("id = ?", c.Project).First(&proj).Error; err != nil {
		return err
	}

	c.VAT = GetVat(tx, proj.Country)
	// recalculate the tables with the new values
	calc, err := calculate(c, proj)
	if err != nil {
		return err
	}

	c.Fields["calculations_qietg"] = strconv.FormatFloat(calc.QIetG, 'f', 2, 64)
	c.Fields["calculations_qapkczg"] = strconv.FormatFloat(calc.QApkCzG, 'f', 2, 64)
	c.Fields["calculations_om1"] = strconv.FormatFloat(calc.OM1, 'f', 2, 64)

	return nil
}

func (Contract) Kind() string {
	return "contract"
}

func (c Contract) Key() string {
	return c.Project.String()
}

func (c Contract) Dependencies() []config.Dependency {
	return []config.Dependency{}
}

func (Contract) TableName() string {
	return "contracts"
}

// New creates an empty contract with initialized agreement fields and tables.
func New(id uuid.UUID) *Contract {
	return &Contract{
		Project:     id,
		Fields:      NewFields(),
		Agreement:   NewAgreement(),
		Tables:      NewTables(),
		Maintenance: NewMaintenance(),
	}
}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, j)
}

// GetEurobor retrieve the the most recent value.
func GetEurobor(db *gorm.DB) decimal.Decimal {
	var v struct {
		ID    uuid.UUID
		Value float64
	}
	if err := db.
		Raw("SELECT id, value FROM eurobor ORDER BY created_at DESC").
		Scan(&v).
		Error; err != nil {

		return decimal.NewFromInt(0)
	}

	return decimal.NewFromFloatWithExponent(v.Value, -4)
}

// GetVat retrieves a given vat % by a country name and returns it as a decimal value.
func GetVat(db *gorm.DB, c models.Country) float64 {
	var cvt models.CountryVat
	if err := db.Where("country = ?", c).First(&cvt).Error; err != nil {
		return float64(0)
	}

	vat := float64(cvt.VAT)
	return vat / 100
}
