package graphql

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// LoadGQLTestFile loads a string with a pre-saved query or response
func LoadGQLTestFile(t *testing.T, name string, a ...interface{}) string {
	f := filepath.Join(basepath, "testdata", name)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		t.Fatalf("Failed to load response data %q: %s", name, err)
		return ""
	}
	return fmt.Sprintf(string(b), a...)
}
