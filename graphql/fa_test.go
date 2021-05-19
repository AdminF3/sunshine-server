package graphql

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func TestFA(t *testing.T) {
	t.Run("create", testCreateFA)
	t.Run("get", testGetFA)
	t.Run("list", testListFA)
	t.Run("getByProject", testgetFAByProject)
	t.Run("review", testReviewFA)
	t.Run("update", testUpdateFA)
}

func testCreateFA(t *testing.T) {
	env := services.NewTestEnv(t)
	pm := stores.NewTestUser(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm.ID),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))
	prjEarly := stores.NewTestProject(t, env.ProjectStore, stores.TPrjWithPm(pm.ID))
	user := stores.NewTestUser(t, env.UserStore)

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:   "ok",
			ctx:    services.NewTestContext(t, env, pm),
			query:  LoadGQLTestFile(t, "mutation_createForfaitingApplication_request.json", user.ID, prj.ID),
			result: `{"createForfaitingApplication": {"privateBond": true}}`,
		},
		{
			name:   "early",
			ctx:    services.NewTestContext(t, env, pm),
			query:  LoadGQLTestFile(t, "mutation_createForfaitingApplication_request.json", user.ID, prjEarly.ID),
			errors: []string{"milestone is lower than 'forfaiting_payout'"},
		},
		{
			name:   "unauthorized",
			ctx:    ctx,
			query:  LoadGQLTestFile(t, "mutation_createForfaitingApplication_request.json", user.ID, prj.ID),
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

func testGetFA(t *testing.T) {
	env := services.NewTestEnv(t)
	pm := stores.NewTestUser(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm.ID),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))
	fa := stores.NewTestFA(t, env.FAStore, stores.TFAWithProject(prj.ID))
	// attachments
	attB := models.Attachment{Owner: fa.ID, UploadType: "fa bank confirmation", Name: "bank"}
	attF := models.Attachment{Owner: fa.ID, UploadType: "fa financial statements", Name: "fin"}

	if err := env.AssetStore.DB().Create(&attB).Error; err != nil {
		t.Fatalf("fail to create attachment: %v", err)
	}
	if err := env.AssetStore.DB().Create(&attF).Error; err != nil {
		t.Fatalf("fail to create attachment: %v", err)
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
			ctx:    services.NewTestContext(t, env, pm),
			query:  LoadGQLTestFile(t, "query_getForfaitingApplication_request.json", fa.ID),
			result: fmt.Sprintf(`{"getForfaitingApplication": %s}`, faFmt(fa, fa.Manager, attF)),
		},
		{
			name:   "unauthorized",
			ctx:    ctx,
			query:  LoadGQLTestFile(t, "query_getForfaitingApplication_request.json", fa.ID),
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

func testReviewFA(t *testing.T) {
	env := services.NewTestEnv(t)
	pd := stores.NewTestPortfolioRole(t, env.UserStore,
		models.PortfolioDirectorRole, "Bulgaria")
	pm := stores.NewTestUser(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm.ID), stores.TPrjWithCountry("Bulgaria"),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))
	fa := stores.NewTestFA(t, env.FAStore, stores.TFAWithProject(prj.ID))

	var revid uuid.UUID
	for _, r := range fa.Reviews {
		if r.Type == models.FAReviewTypeFinancial {
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
			query:  LoadGQLTestFile(t, "mutation_reviewForfaitingApplication_request.json", fa.ID, revid),
			result: `{"reviewForfaitingApplication": {"message": "Success!"}}`,
		},
		{
			name:   "comment",
			ctx:    services.NewTestContext(t, env, pd),
			query:  LoadGQLTestFile(t, "mutation_reviewForfaitingApplication_request.json", fa.ID, revid),
			result: `{"reviewForfaitingApplication": {"message": "Success!"}}`,
		},
		{
			name:   "unauthorized",
			ctx:    context.Background(),
			query:  LoadGQLTestFile(t, "mutation_reviewForfaitingApplication_request.json", fa.ID, revid),
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

func faFmt(fa *models.ForfaitingApplication, m models.User, attF ...models.Attachment) string {
	reviews := make([]string, len(fa.Reviews))
	for i, r := range fa.Reviews {
		reviews[i] = reviewFmt(&r)
	}

	var attfmt = ""
	if len(attF) > 0 {
		attfmt = statementFmt(&attF[0])
	}

	return fmt.Sprintf(`{"project": {"ID": %q}, "privateBond": %v, "manager":{"ID": "%s", "name": "%s", "email": "%s"},
		"finance": %q, "reviews": [%s], "financialStatements": [%v]}`,
		fa.Project, fa.PrivateBond,
		m.ID, m.Name, m.Email,
		financeMap[fa.Finance],
		strings.Join(reviews, ", "),
		attfmt)
}

func statementFmt(att *models.Attachment) string {
	return fmt.Sprintf(`{"name": %q}`, att.Name)
}

func reviewFmt(r *models.FAReview) string {
	return fmt.Sprintf(`{"comment": %q}`, r.Comment)
}

func testListFA(t *testing.T) {
	env := services.NewTestEnv(t)
	pd := stores.NewTestPortfolioRole(t, env.UserStore,
		models.PortfolioDirectorRole, "Austria", "Bulgaria", "Latvia")
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithCountry("Bulgaria"),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))
	fas := []*models.ForfaitingApplication{
		stores.NewTestFA(t, env.FAStore, stores.TFAWithProject(prj.ID)),
		stores.NewTestFA(t, env.FAStore),
		stores.NewTestFA(t, env.FAStore),
		stores.NewTestFA(t, env.FAStore),
	}

	const listResp = `{ "listForfaitingApplications": [%s] }`

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name: "one",
			ctx:  services.NewTestContext(t, env, pd),
			query: LoadGQLTestFile(t, "query_listForfaitingApplications_request.json",
				`countries: ["Latvia"]`),
			result: fmt.Sprintf(listResp, strings.Join([]string{
				faFmt(fas[3], fas[3].Manager), faFmt(fas[2], fas[2].Manager), faFmt(fas[1], fas[1].Manager)}, ", ")),
		},
		{
			name: "two",
			ctx:  services.NewTestContext(t, env, pd),
			query: LoadGQLTestFile(t, "query_listForfaitingApplications_request.json",
				`countries: ["Bulgaria", "Latvia"]`),
			result: fmt.Sprintf(listResp, strings.Join([]string{
				faFmt(fas[3], fas[3].Manager), faFmt(fas[2], fas[2].Manager), faFmt(fas[1], fas[1].Manager), faFmt(fas[0], fas[0].Manager)}, ", ")),
		},
		{
			name: "empty",
			ctx:  services.NewTestContext(t, env, pd),
			query: LoadGQLTestFile(t, "query_listForfaitingApplications_request.json",
				`countries: ["Austria"]`),
			result: fmt.Sprintf(listResp, ""),
		},
		{
			name: "unauthorized",
			ctx:  ctx,
			query: LoadGQLTestFile(t, "query_listForfaitingApplications_request.json",
				`countries: ["Latvia"]`),
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

func testUpdateFA(t *testing.T) {
	env := services.NewTestEnv(t)

	pm := stores.NewTestUser(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm.ID),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))
	fa := stores.NewTestFA(t, env.FAStore, stores.TFAWithProject(prj.ID))

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:   "ok",
			ctx:    services.NewTestContext(t, env, pm),
			query:  LoadGQLTestFile(t, "mutation_updateForfaitingApplication_request.json", fa.ID.String()),
			result: `{"updateForfaitingApplication": {"privateBond": true}}`,
		},
		{
			name:   "unauthorized",
			ctx:    ctx,
			query:  LoadGQLTestFile(t, "mutation_updateForfaitingApplication_request.json", fa.ID.String()),
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

func testgetFAByProject(t *testing.T) {
	env := services.NewTestEnv(t)
	pm := stores.NewTestUser(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm.ID),
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation))
	fa := stores.NewTestFA(t, env.FAStore, stores.TFAWithProject(prj.ID))

	pm2 := stores.NewTestUser(t, env.UserStore)
	prj2 := stores.NewTestProject(t, env.ProjectStore,
		stores.TPrjWithPm(pm2.ID))

	cases := []struct {
		name   string
		ctx    context.Context
		query  string
		result string
		errors []string
	}{
		{
			name:  "ok",
			ctx:   services.NewTestContext(t, env, pm),
			query: LoadGQLTestFile(t, "query_getFAByProject_request.json", prj.ID.String()),
			result: fmt.Sprintf(`{"getForfaitingApplicationByProject":{
						"bankAccount": {
							"IBAN": "%s",
							"bankNameAddress": "%s",
							"beneficiaryName": "%s"
						},
						"privateBond": false
					}}`,
				fa.BankAccount.IBAN,
				fa.BankAccount.BankNameAddress,
				fa.BankAccount.BeneficiaryName),
		},
		{
			name:   "unauthorized",
			ctx:    ctx,
			query:  LoadGQLTestFile(t, "query_getFAByProject_request.json", prj.ID.String()),
			errors: []string{"unauthorized"},
		},
		{
			name:   "rnf",
			ctx:    services.NewTestContext(t, env, pm2),
			query:  LoadGQLTestFile(t, "query_getFAByProject_request.json", prj2.ID.String()),
			errors: []string{"record not found"},
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
