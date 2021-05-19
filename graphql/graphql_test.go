package graphql

// This is taken from "github.com/graph-gophers/graphql-go/gqltesting" in order
// to be able to test errors.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/golang/mock/gomock"
)

var any = gomock.Any()

// GraphQLTest is a GraphQL test case to be used with RunGraphQLTest(s).
type GraphQLTest struct {
	Context       context.Context
	Handler       http.Handler
	Query         string
	OperationName string
	Variables     map[string]interface{}
	Result        string
	Errors        []string
}

func (g *GraphQLTest) Exec(t *testing.T) ([]byte, []string) {
	t.Helper()

	opts := make([]client.Option, 0, len(g.Variables))
	for k, v := range g.Variables {
		opts = append(opts, client.Var(k, v))
	}

	cl := client.New(mockCtx(g.Context, g.Handler), opts...)
	resp, err := cl.RawPost(g.Query)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatal(err)
	}

	var errs []Message
	if err := json.Unmarshal(resp.Errors, &errs); err != nil && len(resp.Errors) > 0 {
		t.Fatalf("Unmarshal errors: %v", err)
	}
	errors := make([]string, len(errs))
	for i, v := range errs {
		errors[i] = v.Message
	}
	return data, errors
}

// RunGraphQLTest runs a single GraphQL test case.
func RunGraphQLTest(t *testing.T, test GraphQLTest) {
	t.Helper()
	if test.Context == nil {
		test.Context = context.Background()
	}
	result, errors := test.Exec(t)
	cmpErorrs(t, test.Errors, errors)

	if len(test.Result) > 0 {
		got := formatJSON(t, result)
		want := formatJSON(t, []byte(test.Result))
		if !bytes.Equal(got, want) {
			t.Logf("got:  %s", got)
			t.Logf("want: %s", want)
			t.Fail()
		}
	}
}

func formatJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil && len(data) > 0 {
		t.Fatalf("invalid JSON: %s; raw data:\n%s", err, data)
	}
	formatted, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return formatted
}

func cmpErorrs(t *testing.T, exp []string, got []string) {
	t.Helper()
	if len(exp) != len(got) {
		t.Errorf("Expected %d errors, got %d: %v", len(exp), len(got), got)
		return
	}

	for i := range got {
		if !strings.Contains(got[i], exp[i]) {
			t.Errorf("Expected error %d to be %q, got %q", i, exp[i], got[i])
		}
	}
}

func mockCtx(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
