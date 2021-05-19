package openapi

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

func IsDocumented() bool {
	return areModelsDocumented() //&& areRoutes()
}

func areModelsDocumented() bool {
	// schemas as model from json files
	schemas, err := ReadSchemas("./")
	if err != nil {
		log.Fatalf("could not read schemas: %v", err)
		return false
	}

	var srcModels = make([]Model, 0)
	var docModels = make([]Model, 0)

	for _, sch := range schemas {
		srcModels = append(srcModels, extractSrcModels(sch)...)
		docModels = append(docModels, parseDocsModel(sch.Components)...)
	}

	return compare(srcModels, docModels, "source code") && compare(docModels, srcModels, "openapi")
}

// diff is some sort of deep equal of two structs of type openapi.Model
func diff(doc, src Model) bool {
	for k, v := range doc.fields {
		// Just in case! Only doc.field.type should be converted, but
		// it is not clearer that the user of the api would not reverse
		// the doc and src models.
		f1 := convert(src.fields[k])
		f2 := convert(v)
		if f1 != f2 {
			fmt.Printf("field %q of type %q is found to be %q\n", k, v, src.fields[k])
			return false
		}
	}
	return true
}

func compare(mod, comp []Model, from string) bool {
	var mMod = make(map[string]Model)

	for _, m := range mod {
		mMod[m.Name] = m
	}

	for _, c := range comp {
		mm, ok := mMod[c.Name]
		if !ok {
			fmt.Printf("%s is not documented in %s\n", c.Name, from)
		}
		if !diff(c, mm) {
			fmt.Printf("-> model 1: %#v and \n-> model 2: %#v\n", mm, c)
			return false
		}

		delete(mMod, c.Name)
	}

	if len(mMod) != 0 {
		fmt.Printf("models are not verified %q\n", mMod)
		return false
	}

	return true
}

func convert(t string) string {
	switch t {
	case "int",
		"integer",
		"uint16",
		"uint8",
		"Building",
		"Heating",
		"ValidationStatus",
		"LegalForm",
		"Kind",
		"ProjectStatus":
		return "int"
	case "float64",
		"float32",
		"number":
		return "float"
	case "string",
		"uuid.UUID",
		"time.Time",
		"Milestone",
		"EntityType",
		"UserAction",
		"PortfolioRole",
		"Country",
		"AssetCategory":
		return "string"
	case "object",
		"AssetSnapshot",
		"Coords":
		return "object"
	case "array",
		"OrgRoles",
		"ProjRoles",
		"[]SocialProfile",
		"[]Notification",
		"[]uuid.UUID",
		"[]string",
		"[]Column",
		"[]Row",
		"[]OrganizationRole",
		"[]ProjectRole",
		"pq.StringArray",
		"[]CountryRole":
		return "array"
	case "boolean",
		"bool":
		return "bool"
	default:
		// If both types are not defined here, they will be valued as
		// `nil`, thus this will false report that the two fields are
		// equal. Generating random string every time default case is
		// called mayfix this.
		return "nil" + uuid.New().String()
	}
}
