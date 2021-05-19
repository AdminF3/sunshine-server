package http

import (
	"encoding/json"
	"fmt"
	"io"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"
)

// jsonDocument wraps *models.Document and add dependency map inside. It is
// only useful for encoding to JSON.
type jsonDocument struct {
	*models.Document
	Dependencies stores.Dependencies `json:"dependencies"`
	Errors       error               `json:"errors"`
}

// encode creates jsonDocument(s) from given document(s), dependency map and an
// error and encodes them into given io.Writer.
//
// It panics if given payload is not *models.Document nor []models.Document.
func encode(w io.Writer, payload interface{}, deps stores.Dependencies, err error) error {
	var data interface{}

	switch p := payload.(type) {
	case *models.Document:
		data = jsonDocument{
			Document:     p,
			Dependencies: deps,
			Errors:       err,
		}
	case []models.Document:
		data = jsonDocuments{
			Documents:    p,
			Dependencies: deps,
			Errors:       err,
		}
	default:
		panic(fmt.Sprintf("Given payload is of unexpected type: %T", payload))
	}

	return json.NewEncoder(w).Encode(data)
}

// jsonDocuments wraps []models.Document and add dependency map inside. It is
// only useful for encoding to JSON.
type jsonDocuments struct {
	Documents    []models.Document   `json:"documents"`
	Dependencies stores.Dependencies `json:"dependencies"`
	Errors       error               `json:"errors"`
}
