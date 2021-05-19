package graphql

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
)

func TestUsers(t *testing.T) {
	t.Run("users_resendActivationEmail", testResendActivationEmail)
}

func testResendActivationEmail(t *testing.T) {
	env := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	user := stores.NewTestUser(t, env.UserStore)
	userCtx := services.NewTestContext(t, env, user)
	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "normal",
			ctx:    userCtx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_resendActivationEmail_request.json", user.Data.(*models.User).Email),
			result: `{"resendActivationEmail": {"message":"Success!"}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ts := mocks.NewMockTokenStore(mock)
			env.TokenStore = ts
			ts.EXPECT().Create(gomock.Any(), models.CreateToken, user.ID).Return(
				&models.Token{}, nil)
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

func TestAdminsListing(t *testing.T) {
	e := services.NewTestEnv(t)

	// create admin users
	su := stores.NewTestAdmin(t, e.UserStore)
	stores.NewTestAdminNwManager(t, e.UserStore)
	stores.NewTestPlatformManager(t, e.UserStore)
	stores.NewTestPortfolioRole(t, e.UserStore, models.InvestorRole, models.CountryLatvia)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		args   string
		result string
	}{
		{
			name:   "unauth",
			ctx:    ctx,
			errors: []string{controller.ErrUnauthorized.Error()},
		},
		{
			name:   "all",
			ctx:    services.NewTestContext(t, e, su),
			result: LoadGQLTestFile(t, "query_adminsListing_response.json"),
		},
		{
			name:   "first 3",
			ctx:    services.NewTestContext(t, e, su),
			result: LoadGQLTestFile(t, "query_adminsListing_3_response.json"),
			args:   `first: 3, offset: 0`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, "query_adminsListing_request.json", c.args),
				Result:  c.result,
			})
		})
	}
}
