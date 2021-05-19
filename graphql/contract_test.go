package graphql

import (
	"context"
	"fmt"
	"testing"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

var ctx = context.Background()

func TestGetIndoorClima_empty(t *testing.T) {
	e := services.NewTestEnv(t)

	_, icprj := stores.NewTestInClima(t, e.IndoorClimaStore)
	pmID := icprj.Data.(*models.Project).Roles.PM[0]
	pm, err := e.UserStore.Get(ctx, pmID)
	if err != nil {
		t.Fatalf("fetch PM: %v", err)
	}
	prj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pmID))
	su := stores.NewTestAdmin(t, e.UserStore)

	cases := []struct {
		name   string
		id     uuid.UUID
		ctx    context.Context
		errors []string
		result string
	}{
		{
			ctx:    services.NewTestContext(t, e, pm),
			name:   "valid",
			id:     icprj.ID,
			result: LoadGQLTestFile(t, "query_getIndoorClima_response.json", icprj.ID),
		},
		{
			ctx:    services.NewTestContext(t, e, su),
			name:   "admin",
			id:     icprj.ID,
			result: LoadGQLTestFile(t, "query_getIndoorClima_response.json", icprj.ID),
		},
		{
			ctx:    ctx,
			name:   "not logged in",
			id:     icprj.ID,
			result: `{"getIndoorClima":null}`,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			ctx:    services.NewTestContext(t, e, stores.NewTestUser(t, e.UserStore)),
			name:   "not a member",
			id:     icprj.ID,
			result: `{"getIndoorClima":null}`,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			ctx:    services.NewTestContext(t, e, pm),
			name:   "missing project",
			id:     uuid.New(),
			result: `{"getIndoorClima":null}`,
			errors: []string{"record not found"},
		},
		{
			ctx:    services.NewTestContext(t, e, pm),
			name:   "project without indoor clima",
			id:     prj.ID,
			result: fmt.Sprintf(`{"getIndoorClima":{"projectID":%q,"zones":{}}}`, prj.ID.String()),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Query:   LoadGQLTestFile(t, "query_getIndoorClima_request.json", c.id),
				Errors:  c.errors,
				Result:  c.result,
			})
		})
	}
}

func TestGetTable(t *testing.T) {
	e := services.NewTestEnv(t)
	pm := stores.NewTestUser(t, e.UserStore)
	us := stores.NewTestUser(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	stores.NewTestContract(t, e.ContractStore, prj)

	testCases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "normal",
			ctx:    services.NewTestContext(t, e, pm),
			query:  LoadGQLTestFile(t, "query_getContract_Table_request.json", prj.ID),
			result: LoadGQLTestFile(t, "query_getContract_Table_response.json"),
		},
		{
			name:   "maintenance Log",
			ctx:    services.NewTestContext(t, e, pm),
			query:  LoadGQLTestFile(t, "query_getMaintenanceLog_Table_request.json", prj.ID),
			result: LoadGQLTestFile(t, "query_getMaintenanceLog_Table_response.json"),
		},
		{
			name:   "unauthorized",
			ctx:    services.NewTestContext(t, e, us),
			query:  LoadGQLTestFile(t, "query_getContract_Table_request.json", prj.ID),
			result: `null`,
			errors: []string{"unauthorized"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: tC.ctx,
				Handler: Handler(e),
				Query:   tC.query,
				Errors:  tC.errors,
				Result:  tC.result,
			})

		})
	}
}

func TestUpdateTable(t *testing.T) {
	e := services.NewTestEnv(t)
	pm := stores.NewTestUser(t, e.UserStore)
	us := stores.NewTestUser(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	stores.NewTestContract(t, e.ContractStore, prj)

	testCases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:   "normal",
			ctx:    services.NewTestContext(t, e, pm),
			query:  LoadGQLTestFile(t, "mutation_updateContract_Table_request.json", prj.ID),
			result: LoadGQLTestFile(t, "mutation_updateContract_Table_response.json"),
		},
		{
			name:   "maintenance table",
			ctx:    services.NewTestContext(t, e, pm),
			query:  LoadGQLTestFile(t, "mutation_updateMaintenanceLog_Table_request.json", prj.ID),
			result: LoadGQLTestFile(t, "mutation_updateMaintenanceLog_Table_response.json"),
		},
		{
			name:   "monitoring table",
			ctx:    services.NewTestContext(t, e, pm),
			query:  LoadGQLTestFile(t, "mutation_updateMonitoring_Table_request.json", prj.ID),
			result: LoadGQLTestFile(t, "mutation_updateMonitoring_Table_response.json"),
		},
		{
			name:   "measurements table",
			ctx:    services.NewTestContext(t, e, pm),
			query:  LoadGQLTestFile(t, "mutation_updateMeasurements_Table_request.json", prj.ID),
			result: LoadGQLTestFile(t, "mutation_updateMeasurements_Table_response.json"),
		},
		{
			name:   "unauthorized",
			ctx:    services.NewTestContext(t, e, us),
			query:  LoadGQLTestFile(t, "mutation_updateContract_Table_request.json", prj.ID),
			result: `null`,
			errors: []string{"unauthorized"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: tC.ctx,
				Handler: Handler(e),
				Query:   tC.query,
				Errors:  tC.errors,
				Result:  tC.result,
			})

		})
	}
}

func TestUpdateIndoorClima(t *testing.T) {
	env := services.NewTestEnv(t)

	_, icp := stores.NewTestInClima(t, env.IndoorClimaStore)
	pmID := icp.Data.(*models.Project).Roles.PM[0]
	pm, err := env.UserStore.Get(ctx, pmID)
	eduvalue := 7277
	if err != nil {
		t.Fatalf("fetch PM: %v", err)
	}

	ctx := services.NewTestContext(t, env, pm)

	RunGraphQLTest(t, GraphQLTest{
		Context: ctx,
		Handler: Handler(env),
		Query:   LoadGQLTestFile(t, "mutation_updateIndoorClima_request.json", icp.ID.String(), eduvalue),
		Errors:  []string{},
		Result:  LoadGQLTestFile(t, "mutation_updateIndoorClima_response.json", icp.ID.String(), eduvalue),
	})
}
