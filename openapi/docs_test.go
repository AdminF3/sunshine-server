package openapi

import "testing"

func tSchema(t *testing.T) []Schema {
	schemas, err := ReadSchemas("./")
	if err != nil {
		t.Fatalf("fail to read jsons with: %v", err)
	}

	return schemas
}

// from json doc files
func TestGetModelProps(t *testing.T) {
	schemas := tSchema(t)

	res := parseDocsModel(schemas[0].Components)

	if len(res) == 0 {
		t.Fatal("expected results be found")
	}

	for _, m := range res {
		if m.Name == "" ||
			len(m.fields) == 0 {
			t.Fatal("empty model found")
		}
	}
}

func TestSchemaComp(t *testing.T) {
	sch := tSchema(t)

	models := make([]XGoType, 0)
	for _, s := range sch {
		models = append(models, getXGoStruct(s.Components)...)
	}

	var b bool
	for _, m := range models {
		if m.ID == "stageai.tech/sunshine/sunshine/models.User" {
			b = true
		}
	}
	if !b {
		t.Fatal("expect to find user model")
	}
}
