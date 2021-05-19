package openapi

import "testing"

func TestDiffData(t *testing.T) {
	docModel := Model{
		Name: "Asset",
		fields: map[string]string{
			"name": "string",
			"area": "integer",
		},
	}

	srcModel := Model{
		Name: "Asset",
		fields: map[string]string{
			"name": "string",
			"area": "int",
		},
	}

	if !diff(docModel, srcModel) {
		t.Fatalf("docModel %#v and srcModel %#v are different", docModel, srcModel)
	}
}

func TestValidate(t *testing.T) {
	if !IsDocumented() {
		t.Fatalf("the project is not well documented")
	}
}
