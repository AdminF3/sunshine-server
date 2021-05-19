package controller

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func (c *Contract) GetIndoorClima(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	ctr, err := c.buildContext(ctx, id, nil, withIndoorClima)
	if err != nil {
		return nil, err
	}

	if !Can(ctx, GetProjectIndoorClima, ctr.project.ID, ctr.project.Country) {
		return nil, ErrUnauthorized
	}

	t := ctr.contract.Tables["calc_energy_fee"]
	aapk, _ := strconv.ParseFloat(t.Row(t.Len()-1).Cell(4).String(), 64)
	ctr.ic.Data.(*contract.IndoorClima).Calculate(
		ctr.contract.Tables,
		aapk,
		float64(ctr.asset.Floors))

	return ctr.ic, nil
}

func (c *Contract) UpdateIndoorClima(ctx context.Context, id uuid.UUID, d decode) (*models.Document, error) {
	ctr, err := c.buildContext(ctx, id, nil, withIndoorClima)
	if err != nil {
		return nil, err
	}

	if !Can(ctx, UpdateProjectIndoorClima, id, ctr.project.Country) {
		return nil, ErrUnauthorized
	}

	icValue := ctr.ic.Data.(*contract.IndoorClima).Value
	ctr.ic.Data.(*contract.IndoorClima).BasementPipes = nil
	ctr.ic.Data.(*contract.IndoorClima).AtticPipes = nil
	ctr.ic.Data.(*contract.IndoorClima).Zones = nil

	if err := d(ctr.ic.Data); err != nil {
		return nil, err
	}
	ctr.ic.Data.(*contract.IndoorClima).Value = icValue

	if err := validateZones(ctr.ic.Data.(*contract.IndoorClima)); err != nil {
		return nil, err
	}

	t := ctr.contract.Tables["calc_energy_fee"]
	Aapk, _ := strconv.ParseFloat(t.Row(t.Len()-1).Cell(4).String(), 64)
	floors := ctr.asset.Floors
	ctr.ic.Data.(*contract.IndoorClima).Calculate(ctr.contract.Tables, Aapk, float64(floors))
	ctr.ic.Data.(*contract.IndoorClima).Project = id

	if err = c.cst.DB().Save(ctr.ic.Data.(*contract.IndoorClima)).Error; err != nil {
		return nil, err
	}

	return c.GetIndoorClima(ctx, id)

}

type decode func(models.Entity) error

func MarshalJSON(r io.Reader) decode {
	return func(e models.Entity) error {
		return json.NewDecoder(r).Decode(e)
	}
}

func MarshalGOB(g *gob.Decoder) decode {
	return func(e models.Entity) error {
		return g.Decode(e)
	}
}

func withIndoorClima(ctx context.Context, ctr *contractCTX, s stores.Store, _ map[string]string) error {
	store := s.FromKind("indoorclima")

	var err error
	ctr.ic, err = store.GetByIndex(ctx, ctr.id.String())
	if err != nil {
		ctr.ic, err = store.Create(ctx, contract.NewIndoorClima(ctr.id))
		if err != nil {
			log.Printf("Failed creating new indoor clima: %s", err)
			return fmt.Errorf("failed creating new indoor clima: %w", ErrFatal)
		}
	}

	return nil
}

func validateZones(ic *contract.IndoorClima) error {
	re := regexp.MustCompile(`(attic|basement_ceiling|ground|roof|basewall|external_door|window|external_wall)_[a-z]+[0-9]_zone[1-2]`)
	for k := range ic.Zones {
		if i := re.FindStringIndex(k); i == nil {
			return fmt.Errorf("%w: zone type does not met structure: %v", ErrBadInput, k)
		}
	}
	return nil
}
