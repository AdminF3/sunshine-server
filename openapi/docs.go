package openapi

// Helper functions that manipulate openapi documentation files (as json) and
// convert the data to appropriate types.

// parseDocsModel by given schema (from json) returns deconstructed objects as
// Model.
func parseDocsModel(json SchemaComponents) []Model {
	var res = make([]Model, 0, len(json.Schemas))

	for tp, sch := range json.Schemas {
		fields := make(map[string]string)

		xType := xgoType(sch)
		if xType.Ignore {
			continue
		}

		props := sch.(map[string]interface{})["properties"].(map[string]interface{})

		for k, v := range props {
			i := v.(map[string]interface{})
			fields[k] = i["type"].(string)
		}
		res = append(res, Model{Name: tp, fields: fields, mType: "doc"})
	}
	return res
}

// getXGoStruct find and return the x-go-type property value of a
// model. x-go-struct hold the absolute path of the model in the source of the
// project.
func getXGoStruct(file SchemaComponents) (models []XGoType) {
	models = make([]XGoType, 0, len(file.Schemas))

	for _, sch := range file.Schemas {
		models = append(models, xgoType(sch))
	}
	return models
}

func xgoType(p interface{}) XGoType {
	prop := p.(map[string]interface{})["x-go-type"]
	ignore, _ := prop.(map[string]interface{})["ignore"].(bool)

	return XGoType{
		Ignore: ignore,
		ID:     prop.(map[string]interface{})["id"].(string),
	}
}
