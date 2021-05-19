package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"stageai.tech/sunshine/sunshine/config"
)

type Country string

const (
	CountryAustria  Country = "Austria"
	CountryBulgaria Country = "Bulgaria"
	CountryLatvia   Country = "Latvia"
	CountryPoland   Country = "Poland"
	CountryRomania  Country = "Romania"
	CountrySlovakia Country = "Slovakia"

	// below are all non-consortium countries
	CountryAlbania           Country = "Albania"
	CountryAndorra           Country = "Andorra"
	CountryArmenia           Country = "Armenia"
	CountryBelarus           Country = "Belarus"
	CountryBelgium           Country = "Belgium"
	CountryBosniaHerzegovina Country = "Bosnia and Herzegovina"
	CountryCroatia           Country = "Croatia"
	CountryCyprus            Country = "Cyprus"
	CountryCzech             Country = "Czech Republic"
	CountryDenmark           Country = "Denmark"
	CountryEstonia           Country = "Estonia"
	CountryFinland           Country = "Finland"
	CountryFrance            Country = "France"
	CountryGeorgia           Country = "Georgia"
	CountryGermany           Country = "Germany"
	CountryGreece            Country = "Greece"
	CountryHungary           Country = "Hungary"
	CountryIreland           Country = "Ireland"
	CountryItaly             Country = "Italy"
	CountryKosovo            Country = "Kosovo"
	CountryLiechtenstein     Country = "Liechtenstein"
	CountryLithuania         Country = "Lithuania"
	CountryLuxembourg        Country = "Luxembourg"
	CountryMalta             Country = "Malta"
	CountryMoldova           Country = "Moldova"
	CountryMonaco            Country = "Monaco"
	CountryMontenegro        Country = "Montenegro"
	CountryNetherlands       Country = "Netherlands"
	CountryMacedonia         Country = "North Macedonia"
	CountryNorway            Country = "Norway"
	CountryPortugal          Country = "Portugal"
	CountryRussia            Country = "Russia"
	CountrySanMarino         Country = "San Marino"
	CountrySerbia            Country = "Serbia"
	CountrySlovenia          Country = "Slovenia"
	CountrySpain             Country = "Spain"
	CountrySweden            Country = "Sweden"
	CountrySwitzerland       Country = "Switzerland"
	CountryTurkey            Country = "Turkey"
	CountryUkraine           Country = "Ukraine"
	CountryUK                Country = "United Kingdom"
	CountryVatican           Country = "Vatican City"
)

// Countries this project operates in.
func Countries() []Country {
	return []Country{CountryAustria, CountryBulgaria, CountryLatvia, CountryPoland, CountryRomania, CountrySlovakia,
		CountryAlbania, CountryAndorra, CountryArmenia, CountryBelarus, CountryBelgium, CountryBosniaHerzegovina,
		CountryCroatia, CountryCyprus, CountryCzech, CountryDenmark, CountryEstonia, CountryFinland, CountryFrance,
		CountryGeorgia, CountryGermany, CountryGreece, CountryHungary, CountryIreland, CountryItaly, CountryKosovo,
		CountryLiechtenstein, CountryLithuania, CountryLuxembourg, CountryMalta, CountryMoldova, CountryMonaco,
		CountryMontenegro, CountryNetherlands, CountryMacedonia, CountryNorway, CountryPortugal, CountryRussia,
		CountrySanMarino, CountrySerbia, CountrySlovenia, CountrySpain, CountrySweden, CountrySwitzerland,
		CountryTurkey, CountryUkraine, CountryUK, CountryVatican}
}

func (c Country) IsConsortium() bool {
	switch c {
	case CountryAustria, CountryBulgaria, CountryLatvia, CountryPoland, CountryRomania, CountrySlovakia:
		return true
	default:
		return false
	}
}

func (c Country) HasAdaptedContract() bool {
	switch c {
	case CountryBulgaria, CountryRomania:
		return true
	default:
		return false
	}
}

func (c Country) Valid() error {
	for _, co := range Countries() {
		if co == c {
			return nil
		}
	}
	return errors.New("invalid country")
}

// Scan implements the database/sql.Scanner interface.
func (c *Country) Scan(value interface{}) error {
	var err error
	switch v := value.(type) {
	case []byte:
		*c = Country(v)
	case string:
		*c = Country(v)
	case fmt.Stringer:
		*c = Country(v.String())
	default:
		*c = Country(fmt.Sprintf("%v", value))
		err = fmt.Errorf("invalid country: %T(%v)", value, value)
	}

	return err
}

func (c Country) String() string               { return strings.Title(strings.ToLower(string(c))) }
func (c Country) Value() (driver.Value, error) { return c.String(), nil }

// LegalCountry returns the country of the project. If the project's
// country is not within the consortium, fallback to latvian. All
// results are formatted with toLower case.
func (c Country) LegalCountry() string {
	if c.IsConsortium() {
		return c.String()
	}

	return CountryLatvia.String()
}

// TODO: With the completion of #522 issue the above enum should become obsolete or somewhat changed.
// Also the below type CountryVat should become the only 'type Country'. It is done this way here
// not to introduce too many refactoring in addition to adding the new functionality and not make
// breaking changes to the frontend as well.

type CountryVat struct {
	Value

	VAT     int
	Country Country
}

func (CountryVat) Kind() string {
	return "country"
}

func (CountryVat) IsEntity() {}

func (c CountryVat) Key() string {
	return c.Country.String()
}

func (CountryVat) TableName() string {
	return "countries"
}

func (CountryVat) Dependencies() []config.Dependency {
	return []config.Dependency{}
}
