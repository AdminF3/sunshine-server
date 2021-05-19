package openapi

type (
	Schema struct {
		Tags       []interface{}          `json:"tags"`
		Paths      map[string]interface{} `json:"paths"`
		Components SchemaComponents       `json:"components"`
	}

	SchemaComponents struct {
		Schemas map[string]interface{} `json:"schemas"`
	}
)

type Model struct {
	Name   string
	fields map[string]string
	mType  string
}

type XGoType struct {
	Ignore bool
	ID     string
}
