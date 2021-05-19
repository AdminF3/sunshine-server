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

func TestGetNotification(t *testing.T) {
	e := services.NewTestEnv(t)

	su := stores.NewTestAdmin(t, e.UserStore)
	no := stores.NewTestNotification(t, e.Notifier, su.ID)

	cases := []struct {
		name   string
		id     uuid.UUID
		ctx    context.Context
		errors []string
		result string
	}{
		{
			name:   "valid",
			id:     no.ID,
			ctx:    services.NewTestContext(t, e, su),
			result: LoadGQLTestFile(t, "query_getNotification_response.json", no.Old),
			errors: []string{},
		},
		{
			name:   "unauthorized",
			id:     no.ID,
			ctx:    ctx,
			result: `{"getNotification":null}`,
			errors: []string{"unauthorized"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, "query_getNotification_request.json", no.ID),
				Result:  c.result,
			})

		})
	}

}

func TestSeeNotification(t *testing.T) {
	e := services.NewTestEnv(t)

	su := stores.NewTestAdmin(t, e.UserStore)
	no := stores.NewTestNotification(t, e.Notifier, su.ID)

	cases := []struct {
		name   string
		id     uuid.UUID
		ctx    context.Context
		errors []string
		result string
	}{
		{
			name:   "valid",
			id:     no.ID,
			ctx:    services.NewTestContext(t, e, su),
			result: `{"seeNotification": {"message":"Success!"}}`,
			errors: []string{},
		},
		{
			name:   "unauthorized",
			id:     no.ID,
			ctx:    ctx,
			result: `{"seeNotification":null}`,
			errors: []string{"unauthorized"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, "mutation_seeNotification_request.json", no.ID),
				Result:  c.result,
			})

		})
	}
}

func TestListNotifications(t *testing.T) {
	e := services.NewTestEnv(t)

	su := stores.NewTestAdmin(t, e.UserStore)
	makeSomeTestNotifications(t, e.Notifier, su.ID, 4)

	cases := []struct {
		name   string
		ctx    context.Context
		action models.UserAction
		errors []string
		result string
	}{
		{
			name:   "valid",
			ctx:    services.NewTestContext(t, e, su),
			result: listNotificationsResponse(t, 4),
			errors: []string{},
		},
		{
			name:   "action match",
			ctx:    services.NewTestContext(t, e, su),
			action: models.UserActionUpload,
			result: listNotificationsResponse(t, 4),
			errors: []string{},
		},
		{
			name:   "action miss",
			ctx:    services.NewTestContext(t, e, su),
			action: models.UserActionCreate,
			result: listNotificationsResponse(t, 0),
			errors: []string{},
		},
		{
			name:   "unauthorized",
			ctx:    ctx,
			result: `{"listNotifications":null}`,
			errors: []string{"unauthorized"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var action string
			if c.action != "" {
				action = fmt.Sprintf("action: %s", strings.ToUpper(string(c.action)))
			}
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, "query_listNotifications_request.json", action),
				Result:  c.result,
			})
		})
	}
}

func TestNotificationListing(t *testing.T) {
	e := services.NewTestEnv(t)

	a := stores.NewTestAsset(t, e.UserStore)
	o := stores.NewTestOrg(t, e.UserStore)
	u := stores.NewTestUser(t, e.UserStore)
	u2 := stores.NewTestUser(t, e.UserStore)
	makeSomeTestNotifications(t, e.Notifier, u.ID, 7)
	stores.NewTestNotification(t, e.Notifier, u.ID, stores.TNWithTarget(a))
	stores.NewTestNotification(t, e.Notifier, u.ID, stores.TNWithOrg(o.Data.(*models.Organization)))
	stores.NewTestNotification(t, e.Notifier, u.ID, stores.TNWithAction(models.UserActionCreate))
	stores.NewTestNotification(t, e.Notifier, u.ID, stores.TNWithUser(u2.Data.(*models.User)))

	argActionAll := "action: [CREATE, UPDATE, UPLOAD, ASSIGN, GDPR, REQUEST_MEMBERSHIP, LEAR_APPLY, REQUEST_PROJECT_CREATION]"
	targetKey := "Goo"
	targetType := "ORGANIZATION"
	cursors := func(from, to int) string {
		var result strings.Builder
		for i := from; i < to; i++ {
			if i > from {
				result.WriteRune(',')
			}
			result.WriteString(fmt.Sprintf(`{"cursor":"%s"}`, encodeCursor(i)))
		}
		return result.String()
	}

	cases := []struct {
		name   string
		ctx    context.Context
		args   string
		errors []string
		result string
	}{
		{
			name:   "all",
			ctx:    services.NewTestContext(t, e, u),
			args:   argActionAll,
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 11), 11),
		},
		{
			name:   "first 5",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("first: 5, %s", argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", true, false, cursors(0, 5), 11),
		},
		{
			name:   "last 3",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("last: 3, %s", argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, true, cursors(8, 11), 11),
		},
		{
			name:   "last 5 before [5]",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("last: 5, before: %q, %s", encodeCursor(10), argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", true, true, cursors(5, 10), 11),
		},
		{
			name:   "first 5 after [2]",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("first: 5, after: %q, %s", encodeCursor(2), argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", true, true, cursors(3, 8), 11),
		},
		{
			name:   "action upload",
			ctx:    services.NewTestContext(t, e, u),
			args:   "action: [UPLOAD]",
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 10), 10),
		},
		{
			name:   "action create",
			ctx:    services.NewTestContext(t, e, u),
			args:   "action: [CREATE]",
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 1), 1),
		},
		{
			name:   "with target",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("targetID: %q, %s", a.ID, argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 1), 1),
		},
		{
			name:   "with user",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("userID: %q, %s", u2.ID, argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 1), 1),
		},
		{
			name:   "with seen",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("seen: true, %s", argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 0), 0),
		},
		{
			name:   "with not seen",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("first: 5, seen: false, %s", argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", true, false, cursors(0, 5), 11),
		},
		{
			name:   "by target type",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf("targetType: %s, %s", targetType, argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 1), 1),
		},
		{
			name:   "by target key",
			ctx:    services.NewTestContext(t, e, u),
			args:   fmt.Sprintf(`targetKey: "%s", %s`, targetKey, argActionAll),
			result: LoadGQLTestFile(t, "query_notificationListing_response.json", false, false, cursors(0, 1), 1),
		},
		{
			name:   "unauthorized",
			ctx:    ctx,
			args:   argActionAll,
			errors: []string{"unauthorized"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RunGraphQLTest(t, GraphQLTest{
				Context: c.ctx,
				Handler: Handler(e),
				Errors:  c.errors,
				Query:   LoadGQLTestFile(t, "query_notificationListing_request.json", c.args),
				Result:  c.result,
			})
		})
	}
}

func makeSomeTestNotifications(t *testing.T, not stores.Notifier, recp uuid.UUID, count int) []models.Notification {
	var ns = make([]models.Notification, count)
	for i := range ns {
		ns[i] = *stores.NewTestNotification(t, not, recp)
	}
	return ns
}

func listNotificationsResponse(t *testing.T, n int) string {
	records := strings.Repeat(`{"old":"hello"},`, n)
	if len(records) > 0 {
		records = records[:len(records)-1] // remove trailing comma
	}
	return LoadGQLTestFile(t, "query_listNotifications_response.json", records)
}
