package graphql

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestPortfolio(t *testing.T) {
	t.Run("data_protec_officer_add", testAssignDataProtectionOfficer)
	t.Run("data_protec_officer_remove", testRemoveDataProtectionOfficer)
	t.Run("country_admin_add", testCountryAdminAddRemove)
}

func testAssignDataProtectionOfficer(t *testing.T) {
	env := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	admin := stores.NewTestAdmin(t, env.UserStore)
	user := stores.NewTestUser(t, env.UserStore)
	adminCtx := services.NewTestContext(t, env, admin)
	userCtx := services.NewTestContext(t, env, user)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "add_dpo",
			ctx:    adminCtx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_assignDPO_request.json", admin.ID.String(), "Bulgaria"),
			result: `{"addPortfolioRole": {"message":"Success!"}}`,
		},
		{
			name:   "add_existant",
			ctx:    adminCtx,
			errors: []string{"duplicate entry"},
			query:  LoadGQLTestFile(t, "mutation_assignDPO_request.json", admin.ID.String(), "Bulgaria"),
			result: `{"addPortfolioRole":null}`,
		},
		{
			name:   "not_an_admin",
			ctx:    userCtx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_assignDPO_request.json", user.ID.String(), "Bulgaria"),
			result: `{"addPortfolioRole":null}`,
		},
		{
			name:   "not_valid_country",
			ctx:    adminCtx,
			errors: []string{"invalid country"},
			query:  LoadGQLTestFile(t, "mutation_assignDPO_request.json", admin.ID.String(), "Durjava"),
			result: `{"addPortfolioRole":null}`,
		},
		{
			name:   "not_logged",
			ctx:    ctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_assignDPO_request.json", admin.ID.String(), "Bulgaria"),
			result: `{"addPortfolioRole": null}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			not := mocks.NewMockNotifier(mock)
			env.Notifier = not
			not.EXPECT().Notify(any, any).AnyTimes()

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

func testRemoveDataProtectionOfficer(t *testing.T) {
	env := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	admin := stores.NewTestAdmin(t, env.UserStore)
	user := stores.NewTestUser(t, env.UserStore)
	adminCtx := services.NewTestContext(t, env, admin)
	userCtx := services.NewTestContext(t, env, user)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "RemoveDPO",
			ctx:    adminCtx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_removeDPO_request.json", admin.ID.String(), "Bulgaria"),
			result: `{"removePortfolioRole": {"message":"Success!"}}`,
		},
		{
			name:   "RemoveUnexistant",
			ctx:    adminCtx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_removeDPO_request.json", admin.ID.String(), "Bulgaria"),
			result: `{"removePortfolioRole": {"message":"Success!"}}`,
		},
		{
			name:   "NotAnAdmin",
			ctx:    userCtx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_removeDPO_request.json", user.ID.String(), "Bulgaria"),
			result: `{"removePortfolioRole":null}`,
		},
		{
			name:   "NotValidCountry",
			ctx:    adminCtx,
			errors: []string{"invalid country"},
			query:  LoadGQLTestFile(t, "mutation_removeDPO_request.json", admin.ID.String(), "Durjava"),
			result: `{"removePortfolioRole":null}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			not := mocks.NewMockNotifier(mock)
			env.Notifier = not
			not.EXPECT().Notify(any, any).AnyTimes()

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

func testCountryAdminAddRemove(t *testing.T) {
	e := services.NewTestEnv(t)

	admin := stores.NewTestAdmin(t, e.UserStore)
	user := stores.NewTestUser(t, e.UserStore)
	pfm := stores.NewTestPlatformManager(t, e.UserStore)
	anm := stores.NewTestAdminNwManager(t, e.UserStore)

	adminCtx := services.NewTestContext(t, e, admin)
	userCtx := services.NewTestContext(t, e, user)
	pfmCtx := services.NewTestContext(t, e, pfm)
	anmCtx := services.NewTestContext(t, e, anm)

	cases := []struct {
		name         string
		ctx          context.Context
		errors       []string
		queryAdd     string
		queryRemove  string
		resultAdd    string
		resultRemove string
	}{
		{
			name:         "ok pfm",
			ctx:          pfmCtx,
			errors:       []string{},
			queryAdd:     LoadGQLTestFile(t, "mutation_addCA_request.json", admin.ID.String(), "Bulgaria"),
			resultAdd:    `{"addCountryAdmin": {"message":"Success!"}}`,
			queryRemove:  LoadGQLTestFile(t, "mutation_removeCA_request.json", admin.ID.String(), "Bulgaria"),
			resultRemove: `{"removeCountryAdmin": {"message":"Success!"}}`,
		},
		{
			name:         "ok anm",
			ctx:          anmCtx,
			errors:       []string{},
			queryAdd:     LoadGQLTestFile(t, "mutation_addCA_request.json", admin.ID.String(), "Bulgaria"),
			resultAdd:    `{"addCountryAdmin": {"message":"Success!"}}`,
			queryRemove:  LoadGQLTestFile(t, "mutation_removeCA_request.json", admin.ID.String(), "Bulgaria"),
			resultRemove: `{"removeCountryAdmin": {"message":"Success!"}}`,
		},
		{
			name:         "ok admin",
			ctx:          adminCtx,
			errors:       []string{},
			queryAdd:     LoadGQLTestFile(t, "mutation_addCA_request.json", admin.ID.String(), "Bulgaria"),
			resultAdd:    `{"addCountryAdmin": {"message":"Success!"}}`,
			queryRemove:  LoadGQLTestFile(t, "mutation_removeCA_request.json", admin.ID.String(), "Bulgaria"),
			resultRemove: `{"removeCountryAdmin": {"message":"Success!"}}`,
		},
		{
			name:         "unauth",
			ctx:          userCtx,
			errors:       []string{"unauthorized"},
			queryAdd:     LoadGQLTestFile(t, "mutation_addCA_request.json", admin.ID.String(), "Bulgaria"),
			resultAdd:    `{"addCountryAdmin":null}`,
			queryRemove:  LoadGQLTestFile(t, "mutation_removeCA_request.json", admin.ID.String(), "Bulgaria"),
			resultRemove: `{"removeCountryAdmin":null}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.queryAdd,
				Result:  c.resultAdd,
			})
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   c.queryRemove,
				Result:  c.resultRemove,
			})

		})
	}
}
