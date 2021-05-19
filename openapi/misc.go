package openapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ReadSchemas takes dir as a destination for a folder with json files; reads
// them, marshal them into Schema and return slice of Schemas.
func ReadSchemas(dir string) ([]Schema, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list %s: %v", dir, err)
	}

	var schemas = make([]Schema, len(files))

	for i, fInfo := range files {
		if !strings.HasSuffix(fInfo.Name(), ".json") || fInfo.IsDir() {
			continue
		}

		s, err := decodeFile(dir, fInfo.Name())
		if err != nil {
			return nil, err
		}

		schemas[i] = *s
	}

	return schemas, nil
}

func decodeFile(dir, fileName string) (*Schema, error) {
	var s Schema

	f, err := os.Open(filepath.Join(dir, fileName))
	if err != nil {
		return nil, fmt.Errorf("open %s: %v", fileName, err)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&s)
	if err != nil {
		if syntaxError, ok := err.(*json.SyntaxError); ok {
			err = fmt.Errorf("%v: %w", syntaxError.Offset, err)
		}
	}
	return &s, err
}
