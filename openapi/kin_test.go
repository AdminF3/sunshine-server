package openapi

import (
	"bytes"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

func cat(t *testing.T) []byte {
	var b bytes.Buffer
	if err := Concatenate(&b, "../openapi"); err != nil {
		t.Fatalf("openapi.Concatenate: %v", err)
	}
	return b.Bytes()
}

func TestOpenAPI3Validate(t *testing.T) {
	openapi3.DefineStringFormat("uuid", openapi3.FormatOfStringForUUIDOfRFC4122)
	router := openapi3filter.NewRouter()
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(cat(t))
	if err != nil {
		t.Fatal(err)
	}

	if err := router.AddSwagger(swagger); err != nil {
		t.Fatal(err)
	}
}
