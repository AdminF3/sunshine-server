package openapi

// Helper functions that manipulate the types in the source code, extract their
// fields and package them into appropriate type.

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
)

var (
	regexpType = regexp.MustCompile(`([\w\[\]*.]+\s+){2}`)
	regexpTag  = regexp.MustCompile(`json:"(\w+)`)
)

// parseSrcModel returns the tags of a model's absolute path.
// e.g. model = "stage.tech/sunshine/sunshine/models.User"
func parseSrcModel(path string) *Model {
	out, err := exec.Command("go", "doc", path).Output()
	if err != nil {
		log.Printf("go doc %s failed: %v\n", path, err)
		return nil
	}

	scanner := strings.Split(string(out), "\n")

	rows := make([]string, 0)
	for _, v := range scanner {
		if v == "" {
			continue
		}

		if strings.Contains(v, "func") {
			continue
		}

		rows = append(rows, v)
	}
	model := Model{
		Name:   fieldName(path),
		fields: make(map[string]string, len(rows)),
	}
	for _, v := range rows {
		tag := fieldTag(v)
		ftype := fieldType(v)

		if tag != "" && ftype != "" {
			model.fields[tag] = ftype
		}
	}

	return &model
}

func extractSrcModels(sch Schema) []Model {
	var res = make([]Model, 0)
	for _, v := range getXGoStruct(sch.Components) {
		m := parseSrcModel(v.ID)
		if m != nil {
			m.mType = "src"
			res = append(res, *m)
		}
	}
	return res
}

func fieldName(path string) string {
	strs := strings.Split(path, ".")

	return strs[len(strs)-1]
}

// fieldType takes row as a string of a model and returns the type of the field
//
// e.g. 	SuperUser bool   `json:"superuser" gorm:"column:is_admin"`
// will return "bool". If the type is not found will return "".
func fieldType(row string) string {
	t := strings.TrimSpace(findSubmatch(row, regexpType))
	return strings.TrimPrefix(t, "*")
}

// fieldTag extract the value of the json tag of the given string.
//
// e.g. 	SuperUser bool   `json:"superuser" gorm:"column:is_admin"`
// will return "superuser". If the tag is not found it will return "".
func fieldTag(row string) string {
	return findSubmatch(row, regexpTag)
}

func findSubmatch(row string, regex *regexp.Regexp) string {
	res := regex.FindStringSubmatch(row)
	if len(res) > 0 {
		return res[1]
	}
	return ""
}
