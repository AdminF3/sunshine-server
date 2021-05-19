package graphql

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestAssignPM(t *testing.T) {
	e := services.NewTestEnv(t)

	su := stores.NewTestAdmin(t, e.UserStore)
	plsign := stores.NewTestUser(t, e.UserStore)
	u := stores.NewTestUser(t, e.UserStore)
	var proles = make(map[string][]uuid.UUID, 1)
	proles["plsign"] = []uuid.UUID{plsign.ID}
	p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithRole(proles))

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		result string
	}{
		{
			name:   "unauth",
			ctx:    services.NewTestContext(t, e, u),
			result: `{"assignPM": null}`,
			errors: []string{"unauthorized"},
		},
		{
			name:   "ok",
			ctx:    services.NewTestContext(t, e, plsign),
			result: `{"assignPM": {"message":"Success!"}}`,
			errors: []string{},
		},
		{
			name:   "superuser",
			ctx:    services.NewTestContext(t, e, su),
			result: `{"assignPM": {"message":"Success!"}}`,
			errors: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, "mutation_assignPM_request.json", p.ID, []uuid.UUID{u.ID}),
				Result:  c.result,
			})
		})
	}
}

func TestAddProjectRole(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	su := stores.NewTestAdmin(t, e.UserStore)
	pm := stores.NewTestUser(t, e.UserStore)
	u := stores.NewTestUser(t, e.UserStore)
	u2 := stores.NewTestUser(t, e.UserStore)

	var proles = make(map[string][]uuid.UUID, 1)
	proles["pm"] = []uuid.UUID{pm.ID}
	p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithRole(proles))

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		result string
		query  string
	}{
		{
			name:   "pm-PLSIGN",
			ctx:    services.NewTestContext(t, e, pm),
			result: `{"addProjectRole": {"message":"Success!"}}`,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_addProjectRole_request.json", p.ID, u2.ID, "PLSIGN"),
		},
		{
			name:   "pm-PACO",
			ctx:    services.NewTestContext(t, e, pm),
			result: `{"addProjectRole": {"message":"Success!"}}`,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_addProjectRole_request.json", p.ID, u2.ID, "PACO"),
		},
		{
			name:   "admin",
			ctx:    services.NewTestContext(t, e, su),
			result: `{"addProjectRole": {"message":"Success!"}}`,
			query:  LoadGQLTestFile(t, "mutation_addProjectRole_request.json", p.ID, u.ID, "PACO"),
			errors: []string{},
		},
		{
			name:   "unauth",
			ctx:    services.NewTestContext(t, e, u),
			result: `{"addProjectRole": null}`,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_addProjectRole_request.json", p.ID, u2.ID, "PACO"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			not := mocks.NewMockNotifier(mock)
			e.Notifier = not
			not.EXPECT().Broadcast(any, any, any, any, any, any, any, any).AnyTimes()

			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}

func TestRemoveProjectRole(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	su := stores.NewTestAdmin(t, e.UserStore)
	pm := stores.NewTestUser(t, e.UserStore)
	paco := stores.NewTestUser(t, e.UserStore)
	teme := stores.NewTestUser(t, e.UserStore)
	u := stores.NewTestUser(t, e.UserStore)

	var proles = make(map[string][]uuid.UUID, 3)
	proles["pm"] = []uuid.UUID{pm.ID}
	proles["paco"] = []uuid.UUID{paco.ID}
	proles["teme"] = []uuid.UUID{teme.ID}

	p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithRole(proles))

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		result string
		query  string
	}{
		{
			name:   "unauth",
			ctx:    services.NewTestContext(t, e, u),
			result: `{"removeProjectRole": null}`,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_removeProjectRole_request.json", p.ID, paco.ID, "PACO"),
		},
		{
			name:   "pm",
			ctx:    services.NewTestContext(t, e, pm),
			result: `{"removeProjectRole": {"message":"Success!"}}`,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_removeProjectRole_request.json", p.ID, paco.ID, "PACO"),
		},
		{
			name:   "admin",
			ctx:    services.NewTestContext(t, e, su),
			result: `{"removeProjectRole": {"message":"Success!"}}`,
			query:  LoadGQLTestFile(t, "mutation_removeProjectRole_request.json", p.ID, teme.ID, "TEME"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			not := mocks.NewMockNotifier(mock)
			e.Notifier = not
			not.EXPECT().Broadcast(any, any, any, any, any, any, any, any).AnyTimes()
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}

func TestRequestProjectCreation(t *testing.T) {
	e := services.NewTestEnv(t)

	asset := stores.NewTestAsset(t, e.AssetStore)
	org := stores.NewTestOrg(t, e.OrganizationStore)
	u := stores.NewTestUser(t, e.UserStore)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name: "unauthorized",
			ctx:  ctx,
			query: LoadGQLTestFile(
				t,
				"mutation_requestProjectCreation_request.json",
				asset.ID.String(),
				org.ID.String(),
			),
			result: `{"requestProjectCreation": null }`,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			name: "default",
			ctx:  services.NewTestContext(t, e, u),
			query: LoadGQLTestFile(
				t,
				"mutation_requestProjectCreation_request.json",
				asset.ID.String(),
				org.ID.String(),
			),
			result: `{"requestProjectCreation":{"message":"Success!"}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}

func TestProcessProjectRequest(t *testing.T) {
	e := services.NewTestEnv(t)

	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	u := stores.NewTestUser(t, e.UserStore)

	org := stores.NewTestOrg(t, e.OrganizationStore, u.ID)
	asset := stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))

	r := stores.NewTestProjectCreationRequest(t,
		e.ProjectStore,
		stores.TPrjReqWithAsset(asset.ID),
	)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name: "unauthorized",
			ctx:  ctx,
			query: LoadGQLTestFile(
				t,
				"mutation_processProjectCreationRequest_request.json",
				r.User,
				r.Asset,
			),
			result: `{"processProjectCreation": null }`,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			name: "default",
			ctx:  services.NewTestContext(t, e, u),
			query: LoadGQLTestFile(
				t,
				"mutation_processProjectCreationRequest_request.json",
				r.User,
				r.Asset,
			),
			result: `{"processProjectCreation":{"message":"Success!"}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}

func TestCommentProject(t *testing.T) {
	e := services.NewTestEnv(t)

	u := stores.NewTestUser(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(u.ID))

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name: "unauthorized",
			ctx:  ctx,
			query: LoadGQLTestFile(
				t,
				"mutation_commentProject_request.json",
				prj.ID,
			),
			result: `{"commentProject": null }`,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			name: "default",
			ctx:  services.NewTestContext(t, e, u),
			query: LoadGQLTestFile(
				t,
				"mutation_commentProject_request.json",
				prj.ID,
			),
			result: LoadGQLTestFile(t, "mutation_commentProject_response.json"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}

func TestAdvanceToMilestone(t *testing.T) {
	e := services.NewTestEnv(t)

	user := stores.NewTestUser(t, e.UserStore)
	project := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(user.ID))
	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name: "unauthorized",
			ctx:  ctx,
			query: LoadGQLTestFile(
				t,
				"mutation_advanceToMilestone_request.json",
				project.ID.String(),
			),
			result: `{"advanceToMilestone": null }`,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			name: "default",
			ctx:  services.NewTestContext(t, e, user),
			query: LoadGQLTestFile(
				t,
				"mutation_advanceToMilestone_request.json",
				project.ID.String(),
			),
			result: `{"advanceToMilestone":{"message":"Success!"}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.query,
				Result:  c.result,
			})
		})
	}
}
