package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Asset struct {
	Value

	// Owner is the organization that owns the asset.
	Owner uuid.UUID `json:"owner" gorm:"column:owner_id" validate:"required"`

	// ESCO is the residence community organization.
	ESCO *uuid.UUID `json:"esco" gorm:"column:esco_id"`

	Address         string           `json:"address" validate:"required"`
	Coordinates     Coords           `json:"coordinates" validate:"required" gorm:"column:coords"`
	Area            int              `json:"area" validate:"required"`
	HeatedArea      int              `json:"heated_area"`
	BillingArea     int              `json:"billing_area"`
	CommonPartsArea int              `json:"common_parts_area"`
	Flats           int              `json:"flats"`
	Floors          int              `json:"floors"`
	StairCases      int              `json:"stair_cases"`
	BuildingType    Building         `json:"building_type" validate:"required"`
	HeatingType     Heating          `json:"heating_type"`
	Valid           ValidationStatus `json:"valid" gorm:"column:status"`
	Cadastre        string           `json:"cadastre" validate:"required"`
	Country         Country          `json:"country" validate:"required"`
	Category        *AssetCategory   `json:"category"`

	Projects []Project `gorm:"foreignkey:Asset"`
}

func (a Asset) ExcerptString() string {
	type coreInfo struct {
		Owner       uuid.UUID
		Address     string
		Cadastre    string
		Coordinates Coords
	}
	ci := coreInfo{Owner: a.Owner, Address: a.Address, Cadastre: a.Cadastre, Coordinates: a.Coordinates}
	out, err := json.Marshal(ci)
	if err != nil {
		return ""
	}
	return string(out)
}

func (Asset) Kind() string {
	return "asset"
}

func (Asset) IsEntity() {}

func (a Asset) Key() string {
	return cid(a.Coordinates)
}

func (a Asset) Dependencies() []config.Dependency {
	d := []config.Dependency{{ID: a.Owner, Kind: "organization"}}
	if a.ESCO != nil {
		d = append(d, config.Dependency{ID: *a.ESCO, Kind: "organization"})
	}
	return d
}

func (Asset) TableName() string {
	return "assets"
}

func (a Asset) Snapshot() AssetSnapshot {
	return AssetSnapshot{
		Owner:      a.Owner,
		ESCO:       a.ESCO,
		Area:       a.Area,
		HeatedArea: a.HeatedArea,

		BillingArea:  a.BillingArea,
		Flats:        a.Flats,
		Floors:       a.Floors,
		StairCases:   a.StairCases,
		BuildingType: a.BuildingType,
		HeatingType:  a.HeatingType,
		Cadastre:     a.Cadastre,
	}
}

type Coords struct {
	Lat float32 `json:"lat"`
	Lng float32 `json:"lng"`
}

func (c *Coords) Scan(src interface{}) error {
	var a pq.Float64Array
	if err := a.Scan(src); err != nil {
		return err
	}

	c.Lat = float32(a[0])
	c.Lng = float32(a[1])
	return nil
}

func (c Coords) Value() (driver.Value, error) {
	return pq.Float64Array{float64(c.Lat), float64(c.Lng)}.Value()
}

func cid(c Coords) string {
	return fmt.Sprintf("{%v, %v}", c.Lat, c.Lng)
}
