package graphql

import (
	"context"
	"fmt"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestMeetings(t *testing.T) {
	t.Run("new", testNew)
	t.Run("get", testGet)
	t.Run("update", testUpdate)
	t.Run("create", testCreate)
	t.Run("delete", testDelete)
	t.Run("list", testList)
}

func testNew(t *testing.T) {
	cases := []struct {
		name  string
		input *models.Document
		check func(*Meeting) bool
	}{
		{
			name:  "nil",
			input: nil,
			check: func(m *Meeting) bool { return m == nil },
		},
		{
			name:  "ok",
			input: models.Wrap(&models.Meeting{Name: "foobar"}),
			check: func(m *Meeting) bool { return m.Name == "foobar" },
		},
		{
			name: "bad kind",
			input: &models.Document{
				Kind: "user",
				Data: &models.User{Name: "foobar"},
			},
			check: func(m *Meeting) bool { return m == nil },
		},
		{
			name: "one attachment",
			input: &models.Document{
				Kind: "meeting",
				Data: &models.Meeting{Name: "foobar"},
				Attachments: map[string]models.Attachment{
					"file1": models.Attachment{Name: "file1"},
				},
			},
			check: func(m *Meeting) bool {
				return m != nil && m.Name == "foobar" && len(m.Attachments) == 1
			},
		},
		{
			name: "many attachments",
			input: &models.Document{
				Kind: "meeting",
				Data: &models.Meeting{Name: "foobar"},
				Attachments: map[string]models.Attachment{
					"file1": models.Attachment{Name: "file1"},
					"file2": models.Attachment{Name: "file2"},
				},
			},
			check: func(m *Meeting) bool {
				return m != nil && m.Name == "foobar" && len(m.Attachments) == 2
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if m := newMeeting(c.input); !c.check(m) {
				t.Fatalf("check failed. Got:\n%#v", m)
			}
		})
	}
}

func testGet(t *testing.T) {
	env := services.NewTestEnv(t)

	admin := stores.NewTestAdmin(t, env.UserStore)
	user := stores.NewTestUser(t, env.UserStore)
	org := stores.NewTestOrg(t, env.OrganizationStore, user.ID, user.ID)
	prj := stores.NewTestProject(t, env.ProjectStore)
	m := stores.NewTestMeeting(t,
		env.MeetingsStore,
		stores.TMeetingWithOrg(org.ID),
		stores.TMeetingWithPrj(prj.ID))

	randomm := stores.NewTestMeeting(t, env.MeetingsStore)
	adminctx := services.NewTestContext(t, env, admin)
	userctx := services.NewTestContext(t, env, user)

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
			query:  LoadGQLTestFile(t, "query_getMeeting_request.json", m.ID.String()),
			result: LoadGQLTestFile(t, "query_getMeeting_response.json"),
		},
		{
			name:   "not-logged-in",
			ctx:    ctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "query_getMeeting_request.json", m.ID.String()),
			result: `{"getMeeting":null}`,
		},
		{
			name:   "not-in-org",
			ctx:    userctx,
			query:  LoadGQLTestFile(t, "query_getMeeting_request.json", randomm.ID.String()),
			errors: []string{"unauthorized"},
			result: `{"getMeeting":null}`,
		},
		{
			name:   "user",
			ctx:    userctx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "query_getMeeting_request.json", m.ID.String()),
			result: LoadGQLTestFile(t, "query_getMeeting_response.json"),
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

func testUpdate(t *testing.T) {
	env := services.NewTestEnv(t)

	admin := stores.NewTestAdmin(t, env.UserStore)
	user := stores.NewTestUser(t, env.UserStore)
	org := stores.NewTestOrg(t, env.OrganizationStore, user.ID, user.ID)
	m := stores.NewTestMeeting(t, env.MeetingsStore, stores.TMeetingWithOrg(org.ID))
	_ = stores.NewTestAttachment(t, env.MeetingsStore, m.ID)

	otherm := stores.NewTestMeeting(t, env.MeetingsStore)
	_ = stores.NewTestAttachment(t, env.MeetingsStore, otherm.ID)
	adminctx := services.NewTestContext(t, env, admin)
	userctx := services.NewTestContext(t, env, user)

	cases := []struct {
		name       string
		ctx        context.Context
		errors     []string
		query      string
		result     string
		respQuery  string
		respResult string
	}{
		{
			name:       "admin",
			ctx:        adminctx,
			errors:     []string{},
			query:      LoadGQLTestFile(t, "mutation_updateMeeting_request.json", m.ID.String()),
			result:     LoadGQLTestFile(t, "mutation_updateMeeting_response.json"),
			respQuery:  fmt.Sprintf("query{getMeeting(mID:\"%s\"){name guests {phone}}}", m.ID.String()),
			respResult: `{"getMeeting":{ "guests": [ {"phone": "911"}],"name": "Final meeting"}}`,
		},
		{
			name:       "not-logged-in",
			ctx:        ctx,
			errors:     []string{"unauthorized"},
			query:      LoadGQLTestFile(t, "mutation_updateMeeting_request.json", m.ID.String()),
			result:     `{"updateMeeting":null}`,
			respQuery:  fmt.Sprintf("query{getMeeting(mID:\"%s\"){name}}", m.ID.String()),
			respResult: `{"getMeeting":null}`,
		},
		{
			name:       "not-in-org",
			ctx:        userctx,
			query:      LoadGQLTestFile(t, "mutation_updateMeeting_request.json", otherm.ID.String()),
			errors:     []string{"unauthorized"},
			result:     `{"updateMeeting": null}`,
			respQuery:  fmt.Sprintf("query{getMeeting(mID:\"%s\"){name}}", otherm.ID.String()),
			respResult: `{"getMeeting": null}`,
		},
		{
			name:       "user",
			ctx:        userctx,
			errors:     []string{},
			query:      LoadGQLTestFile(t, "mutation_updateMeeting_request.json", m.ID.String()),
			result:     LoadGQLTestFile(t, "mutation_updateMeeting_response.json"),
			respQuery:  fmt.Sprintf("query{getMeeting(mID:\"%s\"){name}}", m.ID.String()),
			respResult: `{"getMeeting":{ "name": "Final meeting"}}`,
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

			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(env),
				Errors:  c.errors,
				Query:   c.respQuery,
				Result:  c.respResult,
			})
		})
	}
}

func testDelete(t *testing.T) {
	env := services.NewTestEnv(t)

	user := stores.NewTestUser(t, env.UserStore)
	org := stores.NewTestOrg(t, env.OrganizationStore, user.ID, user.ID)
	m := stores.NewTestMeeting(t, env.MeetingsStore, stores.TMeetingWithOrg(org.ID))
	userctx := services.NewTestContext(t, env, user)

	cases := []struct {
		name   string
		ctx    context.Context
		errors []string
		query  string
		result string
	}{

		{
			name:   "not-logged-in",
			ctx:    ctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_deleteMeeting_request.json", m.ID.String()),
			result: `{"deleteMeeting":null}`,
		},
		{
			name:   "user",
			ctx:    userctx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_deleteMeeting_request.json", m.ID.String()),
			result: `{"deleteMeeting": {"message":"Success!"}}`,
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

func testCreate(t *testing.T) {
	env := services.NewTestEnv(t)

	admin := stores.NewTestAdmin(t, env.UserStore)
	user := stores.NewTestUser(t, env.UserStore)
	uorg := stores.NewTestOrg(t, env.OrganizationStore, user.ID)

	org := stores.NewTestOrg(t, env.OrganizationStore)
	adminctx := services.NewTestContext(t, env, admin)
	userctx := services.NewTestContext(t, env, user)

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
			query:  LoadGQLTestFile(t, "mutation_createMeeting_request.json", uorg.ID.String()),
			result: LoadGQLTestFile(t, "mutation_createMeeting_response.json"),
		},
		{
			name:   "not-logged-in",
			ctx:    ctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_createMeeting_request.json", uorg.ID.String()),
			result: `{"createMeeting": null}`,
		},
		{
			name:   "not-in-org",
			ctx:    userctx,
			errors: []string{"unauthorized"},
			query:  LoadGQLTestFile(t, "mutation_createMeeting_request.json", org.ID.String()),
			result: `{"createMeeting": null}`,
		},
		{
			name:   "user",
			ctx:    userctx,
			errors: []string{},
			query:  LoadGQLTestFile(t, "mutation_createMeeting_request.json", uorg.ID.String()),
			result: LoadGQLTestFile(t, "mutation_createMeeting_response.json"),
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

func testList(t *testing.T) {
	env := services.NewTestEnv(t)

	var (
		admin = stores.NewTestAdmin(t, env.UserStore)

		user = stores.NewTestUser(t, env.UserStore)
		uorg = stores.NewTestOrg(t, env.OrganizationStore, user.ID)

		org = stores.NewTestOrg(t, env.OrganizationStore)
		ro  = stores.NewTestOrg(t, env.OrganizationStore)

		m1 = stores.NewTestMeeting(t, env.MeetingsStore, stores.TMeetingWithOrg(ro.ID))
		m2 = stores.NewTestMeeting(t, env.MeetingsStore, stores.TMeetingWithOrg(uorg.ID))
		m3 = stores.NewTestMeeting(t, env.MeetingsStore, stores.TMeetingWithOrg(org.ID))

		_ = stores.NewTestAttachment(t, env.MeetingsStore, m1.ID)
		_ = stores.NewTestAttachment(t, env.MeetingsStore, m2.ID)
		_ = stores.NewTestAttachment(t, env.MeetingsStore, m3.ID)

		adminctx = services.NewTestContext(t, env, admin)
		userctx  = services.NewTestContext(t, env, user)

		cases = []struct {
			name   string
			ctx    context.Context
			errors []string
			query  string
			result string
		}{
			{
				name:   "admin-list-all",
				ctx:    adminctx,
				errors: []string{},
				query:  LoadGQLTestFile(t, "query_listMeetings_all_request.json"),
				result: LoadGQLTestFile(t, "query_listMeetings_all_response.json", org.ID.String(), uorg.ID.String(), ro.ID.String()),
			},
			{
				name:   "admin-list-one",
				ctx:    adminctx,
				errors: []string{},
				query:  LoadGQLTestFile(t, "query_listMeetings_1_request.json", org.ID.String()),
				result: LoadGQLTestFile(t, "query_listMeetings_1_response.json", org.ID.String()),
			},
			{
				name:   "not-logged-in",
				ctx:    ctx,
				errors: []string{"unauthorized"},
				query:  LoadGQLTestFile(t, "query_listMeetings_all_request.json"),
			},
			{
				name:   "not-in-org",
				ctx:    userctx,
				query:  LoadGQLTestFile(t, "query_listMeetings_1_request.json", org.ID.String()),
				errors: []string{"unauthorized"},
			},
			{
				name:   "user",
				ctx:    userctx,
				query:  LoadGQLTestFile(t, "query_listMeetings_1_request.json", uorg.ID.String()),
				errors: []string{},
				result: LoadGQLTestFile(t, "query_listMeetings_response.json", uorg.ID.String()),
			},
		}
	)

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
