package graphql

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestCountry(t *testing.T) {
	t.Run("setVat", testSetVat)
	t.Run("getCountry", testGetCountry)
}

func testSetVat(t *testing.T) {
	env := services.NewTestEnv(t)
	admin := stores.NewTestAdmin(t, env.UserStore)
	randomu := stores.NewTestUser(t, env.UserStore)
	vt := 42

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:   "ok",
			ctx:    services.NewTestContext(t, env, admin),
			query:  LoadGQLTestFile(t, "mutation_setVat_request.json", vt),
			result: LoadGQLTestFile(t, "mutation_setVat_response.json"),
		},
		{
			name:   "unauth",
			ctx:    services.NewTestContext(t, env, randomu),
			query:  LoadGQLTestFile(t, "mutation_setVat_request.json", vt),
			errors: []string{"unauthorized"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(env),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}

func testGetCountry(t *testing.T) {
	env := services.NewTestEnv(t)
	admin := stores.NewTestAdmin(t, env.UserStore)
	randomu := stores.NewTestUser(t, env.UserStore)

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:   "ok",
			ctx:    services.NewTestContext(t, env, admin),
			query:  LoadGQLTestFile(t, "query_getCountry_request.json", "Bulgaria"),
			result: LoadGQLTestFile(t, "query_getCountry_response.json"),
		},
		{
			name:   "unauth",
			ctx:    services.NewTestContext(t, env, randomu),
			query:  LoadGQLTestFile(t, "query_getCountry_request.json", "Bulgaria"),
			errors: []string{"unauthorized"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(env),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}
