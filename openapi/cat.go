package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"text/template"
)

// Concatenate OpenAPI files from given dir and write them to w.
func Concatenate(w io.Writer, dir string) error {
	schemas, err := ReadSchemas(dir)
	if err != nil {
		return err
	}

	sch := Schema{
		Tags:  make([]interface{}, 0, len(schemas)),
		Paths: make(map[string]interface{}),
		Components: SchemaComponents{
			Schemas: make(map[string]interface{}),
		},
	}

	for _, s := range schemas {
		sch.Tags = append(sch.Tags, s.Tags...)
		for path, v := range s.Paths {
			sch.Paths[path] = v

		}
		for comp, v := range s.Components.Schemas {
			sch.Components.Schemas[comp] = v
		}
	}

	funcs := template.FuncMap{"marshal": func(v interface{}) string {
		a, _ := json.MarshalIndent(v, "  ", "  ")
		return string(a)
	}}

	t := template.Must(template.New("base.tpl").
		Funcs(funcs).
		ParseFiles(filepath.Join(dir, "base.tpl")))

	if err := t.Execute(w, sch); err != nil {
		return fmt.Errorf("template execute: %v", err)
	}

	return nil
}
