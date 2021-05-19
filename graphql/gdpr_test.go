package graphql

import (
	"context"
	"fmt"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
)

func TestGDPRRequests(t *testing.T) {
	t.Run("send", testSendGDPRRequest)
	t.Run("get", testGetGDPRRequest)
	t.Run("list", testListGDPRRequests)

}

func testSendGDPRRequest(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	randomguy := stores.NewTestUser(t, e.UserStore)
	gRequest := models.GDPRRequest{
		RequesterName:    "Ivan",
		RequesterPhone:   "0088112233",
		RequesterEmail:   "i.ivaonv@test.com",
		RequesterAddress: "Test address",
		Name:             "Petar",
		Phone:            "0088112234",
		Email:            "p.petrov@test.com",
		Address:          "test address peter",
		Action:           models.GDPRTypeGet,
		Reason:           "get",
		Information:      "my stuff",
	}
	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		result string
		query  string
	}{
		{
			name: "ok get",
			ctx:  services.NewTestContext(t, e, randomguy),
			query: LoadGQLTestFile(t, "mutation_sendGDPR_request.json",
				gRequest.RequesterName, gRequest.RequesterPhone, gRequest.RequesterEmail,
				gRequest.RequesterAddress, gRequest.Name, gRequest.Phone, gRequest.Email,
				gRequest.Address, gRequest.Reason, gRequest.Information),

			result: `{"sendGDPRRequest": {"name":"Petar"}}`,
			errors: []string{},
		},
		{
			name: "ok_not_logged",
			ctx:  context.Background(),
			query: LoadGQLTestFile(t, "mutation_sendGDPR_request.json",
				gRequest.RequesterName, gRequest.RequesterPhone, gRequest.RequesterEmail,
				gRequest.RequesterAddress, gRequest.Name, gRequest.Phone, gRequest.Email,
				gRequest.Address, gRequest.Reason, gRequest.Information),
			result: `{"sendGDPRRequest": {"name":"Petar"}}`,
			errors: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			not := mocks.NewMockNotifier(mock)
			e.Notifier = not
			not.EXPECT().Notify(any, any).AnyTimes()

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

func testGetGDPRRequest(t *testing.T) {
	e := services.NewTestEnv(t)

	su := stores.NewTestAdmin(t, e.UserStore)
	randomguy := stores.NewTestUser(t, e.UserStore)
	req := stores.NewTestGDPRRequest(t, e.DB, models.GDPRTypeGet)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		result string
		query  string
	}{
		{
			name:   "ok get admin",
			ctx:    services.NewTestContext(t, e, su),
			result: fmt.Sprintf(`{"getGDPRRequest": {"name":"%s"}}`, req.Name),
			query:  fmt.Sprintf(`{getGDPRRequest(rID:"%s") {name}}`, req.ID),
			errors: []string{},
		},
		{
			name:   "unauth get random guy",
			ctx:    services.NewTestContext(t, e, randomguy),
			result: `{"getGDPRRequest": null}`,
			query:  fmt.Sprintf(`{getGDPRRequest(rID:"%s") {name}}`, req.ID),
			errors: []string{"unauthorized"},
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

func testListGDPRRequests(t *testing.T) {
	e := services.NewTestEnv(t)

	su := stores.NewTestAdmin(t, e.UserStore)
	randomguy := stores.NewTestUser(t, e.UserStore)
	stores.NewTestGDPRRequest(t, e.DB, models.GDPRTypeGet)
	stores.NewTestGDPRRequest(t, e.DB, models.GDPRTypeGet)
	stores.NewTestGDPRRequest(t, e.DB, models.GDPRTypeGet)
	stores.NewTestGDPRRequest(t, e.DB, models.GDPRTypeGet)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		result string
		query  string
	}{
		{
			name:   "unauth",
			ctx:    services.NewTestContext(t, e, randomguy),
			result: "",
			query:  LoadGQLTestFile(t, "query_listGdpr_request.json", ""),
			errors: []string{"unauthorized"},
		},
		{
			name:   "all",
			ctx:    services.NewTestContext(t, e, su),
			result: LoadGQLTestFile(t, "query_listGdpr_response.json", 4),
			query:  LoadGQLTestFile(t, "query_listGdpr_request.json", ""),
			errors: []string{},
		},
		{
			name:   "first 4",
			ctx:    services.NewTestContext(t, e, su),
			result: LoadGQLTestFile(t, "query_listGdpr_response.json", 4),
			query:  LoadGQLTestFile(t, "query_listGdpr_request.json", `first: 2, offset: 0`),
			errors: []string{},
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
