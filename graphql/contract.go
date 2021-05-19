package graphql

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"strconv"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/google/uuid"
)

func (r *queryResolver) GetTable(ctx context.Context, projectID uuid.UUID, annexN *int, tableName string) (*contract.Table, error) {
	vars := map[string]string{
		"annexn": "",
		"table":  tableName,
	}
	if annexN != nil {
		vars["annexn"] = strconv.Itoa(*annexN)
	}

	t, err := r.ctr.GetTable(ctx, projectID, vars)
	return t, err
}

func (r *mutationResolver) UpdateTable(ctx context.Context, projectID uuid.UUID, annexN *int, tableName string, table *UpdateTable) (*contract.Table, error) {
	vars := map[string]string{
		"annexn": "",
		"table":  tableName,
	}
	if annexN != nil {
		vars["annexn"] = strconv.Itoa(*annexN)
	}
	t, err := unmarshalUpdateTable(table)
	if err != nil {
		return nil, err
	}

	c, err := r.ctr.UpdateTable(ctx, projectID, *t, vars)

	return c, err
}

func (r *queryResolver) GetIndoorClima(ctx context.Context, projectID uuid.UUID) (*contract.IndoorClima, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, errors.New("unauthorized")
	}

	p, err := r.env.ProjectStore.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, r := range p.Data.(*models.Project).ProjectRoles {
		if r.UserID == cv.User.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		user, err := r.env.UserStore.Get(ctx, cv.User.ID)
		if err != nil || !user.Data.(*models.User).SuperUser {
			return nil, errors.New("unauthorized")
		}
	}

	ic, err := r.env.IndoorClimaStore.GetByIndex(ctx, projectID.String())
	if err != nil {
		ic, err = r.env.IndoorClimaStore.Create(ctx, contract.NewIndoorClima(projectID))
		if err != nil {
			return nil, err
		}
	}

	return ic.Data.(*contract.IndoorClima), nil
}

func (r *mutationResolver) UpdateIndoorClima(ctx context.Context, projectID uuid.UUID, values contract.IndoorClima) (*contract.IndoorClima, error) {
	gob.Register(contract.IndoorClima{})

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	dec := gob.NewDecoder(&buff)

	e := models.Entity(values)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}

	decode := controller.MarshalGOB(dec)

	doc, err := r.ctr.UpdateIndoorClima(ctx, projectID, decode)
	if err != nil {
		return nil, err
	}

	return doc.Data.(*contract.IndoorClima), nil
}

func unmarshalUpdateTable(table *UpdateTable) (*contract.Table, error) {
	rows := make([]contract.Row, len(table.Rows))
	for i, r := range table.Rows {
		cells := make([]contract.Cell, len(r))
		for j, c := range r {
			cells[j] = contract.Cell(*c)
		}
		rows[i] = cells
	}
	col := make([]contract.Column, len(rows[0]))
	t, err := contract.NewTable(col, rows...)

	return &t, err
}
