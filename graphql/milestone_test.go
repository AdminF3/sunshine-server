package graphql

import (
	"context"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestWPMP(t *testing.T) {
	t.Run("advance_wp", testAdvanceToWP)
	t.Run("advance_mp", testAdvanceToMP)
	t.Run("get_wp", testGetWP)
	t.Run("get_mp", testGetMP)
	t.Run("review_wp", testReviewWP)
	t.Run("review_mp", testReviewMP)
}

func testAdvanceToWP(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	admin := stores.NewTestAdmin(t, e.UserStore)
	user := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore)
	adminctx := services.NewTestContext(t, e, admin)
	userctx := services.NewTestContext(t, e, user)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "admin",
			ctx:    adminctx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_advanceWP_request.json", prj.ID),
			result: LoadGQLTestFile(t, "mutation_advanceWP_result.json", prj.ID),
		},
		{
			name:   "unauth",
			ctx:    userctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_advanceWP_request.json", prj.ID),
			result: `{"advanceProjectToWorkPhase":null}`,
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

func testAdvanceToMP(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	admin := stores.NewTestAdmin(t, e.UserStore)
	user := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore)
	adminctx := services.NewTestContext(t, e, admin)
	userctx := services.NewTestContext(t, e, user)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "admin",
			ctx:    adminctx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_advanceMP_request.json", prj.ID),
			result: LoadGQLTestFile(t, "mutation_advanceMP_result.json", prj.ID),
		},
		{
			name:   "unauth",
			ctx:    userctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_advanceMP_request.json", prj.ID),
			result: `{"advanceProjectToMonitoringPhase":null}`,
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

func testGetWP(t *testing.T) {
	e := services.NewTestEnv(t)

	admin := stores.NewTestAdmin(t, e.UserStore)
	user := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore)
	wp := stores.NewTestWorkPhase(t, e.WPStore, prj.ID)
	adminctx := services.NewTestContext(t, e, admin)
	userctx := services.NewTestContext(t, e, user)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "admin",
			ctx:    adminctx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "query_getWP_request.json", wp.ID),
			result: LoadGQLTestFile(t, "query_getWP_result.json", prj.ID),
		},
		{
			name:   "unauth",
			ctx:    userctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "query_getWP_request.json", wp.ID),
			result: `{"getWorkPhase":null}`,
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

func testGetMP(t *testing.T) {
	e := services.NewTestEnv(t)

	admin := stores.NewTestAdmin(t, e.UserStore)
	user := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore)
	mp := stores.NewTestMonitoringPhase(t, e.MPStore, prj.ID)
	adminctx := services.NewTestContext(t, e, admin)
	userctx := services.NewTestContext(t, e, user)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{
		{
			name:   "admin",
			ctx:    adminctx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "query_getMP_request.json", mp.ID),
			result: LoadGQLTestFile(t, "query_getMP_result.json", prj.ID),
		},
		{
			name:   "unauth",
			ctx:    userctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "query_getMP_request.json", mp.ID),
			result: `{"getMonitoringPhase":null}`,
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

func testReviewWP(t *testing.T) {
	env := services.NewTestEnv(t)
	pd := stores.NewTestPortfolioRole(t, env.UserStore,
		models.PortfolioDirectorRole, "Bulgaria")
	pm := stores.NewTestUser(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm.ID), stores.TPrjWithCountry("Bulgaria"),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))

	wp := stores.NewTestWorkPhase(t, env.WPStore, prj.ID)
	var revid uuid.UUID
	for _, r := range wp.Reviews {
		if r.Type == models.WPReviewTypeFinancial {
			revid = r.ID
			break
		}
	}

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:   "ok",
			ctx:    services.NewTestContext(t, env, pd),
			query:  LoadGQLTestFile(t, "mutation_reviewWorkPhase_request.json", wp.ID, revid),
			result: `{"reviewWorkPhase": {"message": "Success!"}}`,
		},
		{
			name:   "comment",
			ctx:    services.NewTestContext(t, env, pd),
			query:  LoadGQLTestFile(t, "mutation_reviewWorkPhase_request.json", wp.ID, revid),
			result: `{"reviewWorkPhase": {"message": "Success!"}}`,
		},
		{
			name:   "unauthorized",
			ctx:    context.Background(),
			query:  LoadGQLTestFile(t, "mutation_reviewWorkPhase_request.json", wp.ID, revid),
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

func testReviewMP(t *testing.T) {
	env := services.NewTestEnv(t)
	pd := stores.NewTestPortfolioRole(t, env.UserStore,
		models.PortfolioDirectorRole, "Bulgaria")
	pm := stores.NewTestUser(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm.ID), stores.TPrjWithCountry("Bulgaria"),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))

	mp := stores.NewTestMonitoringPhase(t, env.MPStore, prj.ID)

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:   "ok",
			ctx:    services.NewTestContext(t, env, pd),
			query:  LoadGQLTestFile(t, "mutation_reviewMonitoringPhase_request.json", mp.ID, mp.Reviews[0].ID),
			result: `{"reviewMonitoringPhase": {"message": "Success!"}}`,
		},
		{
			name:   "comment",
			ctx:    services.NewTestContext(t, env, pd),
			query:  LoadGQLTestFile(t, "mutation_reviewMonitoringPhase_request.json", mp.ID, mp.Reviews[0].ID),
			result: `{"reviewMonitoringPhase": {"message": "Success!"}}`,
		},
		{
			name:   "unauthorized",
			ctx:    context.Background(),
			query:  LoadGQLTestFile(t, "mutation_reviewMonitoringPhase_request.json", mp.ID, mp.Reviews[0].ID),
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
