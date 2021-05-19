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

func TestAddOrganizationRole(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	su := stores.NewTestAdmin(t, e.UserStore)
	lear := stores.NewTestUser(t, e.UserStore)
	u := stores.NewTestUser(t, e.UserStore)
	u2 := stores.NewTestUser(t, e.UserStore)

	o := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)

	cases := []struct {
		name   string
		org    uuid.UUID
		ctx    context.Context
		errors []string
		result string
		query  string
	}{
		{
			name:   "unauth",
			org:    o.ID,
			ctx:    services.NewTestContext(t, e, u),
			result: `{"addOrganizationRole": null}`,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_addOrganizationRole_request.json", o.ID, u2.ID, "LEAAS"),
		},
		{
			name:   "lear",
			org:    o.ID,
			ctx:    services.NewTestContext(t, e, lear),
			result: `{"addOrganizationRole": {"message":"Success!"}}`,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_addOrganizationRole_request.json", o.ID, u2.ID, "LEAAS"),
		},
		{
			name:   "admin",
			org:    o.ID,
			ctx:    services.NewTestContext(t, e, su),
			result: `{"addOrganizationRole": {"message":"Success!"}}`,
			query:  LoadGQLTestFile(t, "mutation_addOrganizationRole_request.json", o.ID, u.ID, "LEAAS"),
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

func TestRemoveOrganizationRole(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	su := stores.NewTestAdmin(t, e.UserStore)
	lear := stores.NewTestUser(t, e.UserStore)
	u := stores.NewTestUser(t, e.UserStore)
	u2 := stores.NewTestUser(t, e.UserStore)
	random := stores.NewTestUser(t, e.UserStore)

	o := stores.NewTestOrg(t, e.OrganizationStore, lear.ID, u.ID, u2.ID)

	cases := []struct {
		name   string
		org    uuid.UUID
		ctx    context.Context
		errors []string
		result string
		query  string
	}{
		{
			name:   "unauth",
			org:    o.ID,
			ctx:    services.NewTestContext(t, e, random),
			result: `{"removeOrganizationRole": null}`,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_removeOrganizationRole_request.json", o.ID, u2.ID, "LEAAS"),
		},
		{
			name:   "lear",
			org:    o.ID,
			ctx:    services.NewTestContext(t, e, lear),
			result: `{"removeOrganizationRole": {"message":"Success!"}}`,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_removeOrganizationRole_request.json", o.ID, u2.ID, "LEAAS"),
		},
		{
			name:   "admin",
			org:    o.ID,
			ctx:    services.NewTestContext(t, e, su),
			result: `{"removeOrganizationRole": {"message":"Success!"}}`,
			query:  LoadGQLTestFile(t, "mutation_removeOrganizationRole_request.json", o.ID, u.ID, "LEAAS"),
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

func TestOrganizationReports(t *testing.T) {
	e := services.NewTestEnv(t)

	u := stores.NewTestPlatformManager(t, e.UserStore)

	stores.NewTestProject(t, e.ProjectStore)
	stores.NewTestProject(t, e.ProjectStore)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		result string
		args   string
	}{
		{
			name:   "unauth",
			ctx:    ctx,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			name:   "all",
			ctx:    services.NewTestContext(t, e, u),
			result: LoadGQLTestFile(t, "query_organizationReportsListing_response.json", 4),
		},
		{
			name:   "first 2",
			ctx:    services.NewTestContext(t, e, u),
			result: LoadGQLTestFile(t, "query_organizationReportsListing_response.json", 4),
			args:   `first: 2, offset: 0`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, "query_organizationReportsListing_request.json", c.args),
				Result:  c.result,
			})
		})
	}
}
