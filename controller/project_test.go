package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

const (
	projectTmpl = `{
    	"name": "%s",
    	"owner": "%s",
    	"asset": "%s",
	"roles": {
		"pm": [%q]
        },
    	"status": 1,
	"airtemp": 20,
	"watertemp": 40,
	"savings": 51.16,
	"portfolio_director": %q,
	"country": "Latvia"
    }`
)

func TestAssignProjectRoles(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	p := NewProject(e)
	ustore := p.st.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	lear := stores.NewTestUser(t, ustore)
	plsign := stores.NewTestUser(t, ustore)
	pm := stores.NewTestUser(t, ustore)
	u := stores.NewTestUser(t, ustore)
	u2 := stores.NewTestUser(t, ustore)

	o := stores.NewTestOrg(t, p.st.FromKind("organization"), lear.ID)

	var proles = make(map[string][]uuid.UUID, 1)
	proles["plsign"] = []uuid.UUID{plsign.ID}
	proles["pm"] = []uuid.UUID{pm.ID}

	proj1 := stores.NewTestProject(t, p.st, stores.TPrjWithOrg(o.ID),
		stores.TPrjWithRole(proles))
	prj1 := proj1.Data.(*models.Project)
	rc1 := len(prj1.Roles.PM)
	updater1 := []uuid.UUID{u.ID, u2.ID}

	proj3 := stores.NewTestProject(t, p.st, stores.TPrjWithOrg(o.ID),
		stores.TPrjWithRole(proles))
	prj3 := proj3.Data.(*models.Project)
	rc3 := len(prj3.Roles.PM)

	cases := []struct {
		name     string
		ctx      context.Context
		proj     uuid.UUID
		action   Action
		roles    []uuid.UUID
		expected error
		count    int
	}{
		{
			name:     "ok PLSIGN assigning PM",
			ctx:      services.NewTestContext(t, e, plsign),
			proj:     proj1.ID,
			action:   AssignPM,
			roles:    updater1,
			expected: nil,
			count:    rc1 + len(updater1),
		},
		{
			name:     "ok PM assigning PM",
			ctx:      services.NewTestContext(t, e, pm),
			proj:     proj1.ID,
			action:   AssignPM,
			roles:    updater1,
			expected: nil,
			count:    rc1 + len(updater1),
		},
		{
			name:     "no proj",
			ctx:      services.NewTestContext(t, e, su),
			proj:     uuid.New(),
			action:   AssignPM,
			roles:    []uuid.UUID{u.ID},
			expected: ErrNotFound,
			count:    0,
		},
		{
			name:     "unauth PM",
			ctx:      services.NewTestContext(t, e, lear),
			proj:     proj3.ID,
			action:   AssignPM,
			roles:    []uuid.UUID{u.ID},
			expected: ErrUnauthorized,
			count:    rc3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := p.AssignProjectRoles(c.ctx, c.proj, c.roles, c.action)
			if err != c.expected {
				t.Errorf("expected err: %v, \n got: %v", c.expected, err)
			}

			if c.expected == ErrNotFound {
				// project is not found
				t.Skip()
			}

			pdoc, err := e.ProjectStore.Get(c.ctx, c.proj)
			if err != nil {
				t.Errorf("Failed to get project; %v", err)
			}

			pr := pdoc.Data.(*models.Project).Roles
			upCount := len(pr.PM)
			if upCount != c.count {
				t.Errorf("expected: %v, got: %v roles", c.count, upCount)
			}
		})
	}
}

func TestProject(t *testing.T) {
	t.Run("create", testCreate)
	t.Run("requestPrjCreation", testRequestPrjCreation)
	t.Run("processPrjCreation", testProcessPrjRequest)
	t.Run("listsByIDs", testListByIDs)
	t.Run("commentProject", testCommentProject)
	t.Run("advanceToMilestone", testAdvanceToMilestone)
	t.Run("reports", testReports)
}

func testRequestPrjCreation(t *testing.T) {
	e := services.NewTestEnv(t)

	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	p := NewProject(e)

	org := stores.NewTestOrg(t, e.OrganizationStore)
	u := stores.NewTestUser(t, e.UserStore)

	cases := []struct {
		name  string
		ctx   context.Context
		asset uuid.UUID
		err   error
		// preCreate function for setup dependencies
		pre func(context.Context, uuid.UUID)
	}{
		{
			name:  "ok",
			ctx:   services.NewTestContext(t, e, u),
			asset: stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID)).ID,
			err:   nil,
			pre:   func(context.Context, uuid.UUID) {},
		},
		{
			name: "unauthorized",
			ctx:  context.Background(),
			err:  ErrUnauthorized,
			pre:  func(context.Context, uuid.UUID) {},
		},
		{
			name:  "duplicate",
			ctx:   services.NewTestContext(t, e, u),
			asset: stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID)).ID,
			pre: func(ctx context.Context, aid uuid.UUID) {
				err := p.RequestProjectCreation(ctx, aid, org.ID)
				if err != nil {
					t.Errorf("error occurs:  %v", err)
				}
			},
			err: ErrDuplicate,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.pre(c.ctx, c.asset)

			err := p.RequestProjectCreation(c.ctx, c.asset, org.ID)
			if !errors.Is(err, c.err) {
				t.Errorf("expected err: %v, got: %v", c.err, err)
			}

			if c.err != nil {
				return
			}

			var res models.ProjectCreationRequest
			p.st.DB().Where("asset_id = ? AND organization_id = ? AND status = 'opened'", c.asset, org.ID).First(&res)

			if res.Asset != c.asset ||
				res.Organization != org.ID ||
				res.Status != models.OpenedStatus {
				t.Errorf("cannot fetch request: %v", res)
			}
		})
	}
}

func testProcessPrjRequest(t *testing.T) {
	e := services.NewTestEnv(t)

	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	p := NewProject(e)

	u := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, u.ID)

	ru := stores.NewTestUser(t, e.UserStore)

	cases := []struct {
		name      string
		ctx       context.Context
		target    *models.Document
		asset     uuid.UUID
		appr      bool
		status    models.ProjectCreationRequestStatus
		err       error
		haveToken bool
	}{
		{
			name:   "ok_approve",
			ctx:    services.NewTestContext(t, e, u),
			target: stores.NewTestUser(t, e.UserStore),
			asset: stores.NewTestAsset(t,
				e.AssetStore,
				stores.TAWithOrg(org.ID),
			).ID,
			appr:      true,
			status:    models.AcceptedStatus,
			haveToken: true,
		},
		{
			name:   "ok_reject",
			ctx:    services.NewTestContext(t, e, u),
			target: stores.NewTestUser(t, e.UserStore),
			asset: stores.NewTestAsset(t,
				e.AssetStore,
				stores.TAWithOrg(org.ID),
			).ID,
			appr:   false,
			status: models.RejectedStatus,
		},
		{
			name:   "unauthorized",
			ctx:    context.Background(),
			target: stores.NewTestUser(t, e.UserStore),
			asset: stores.NewTestAsset(t,
				e.AssetStore,
				stores.TAWithOrg(org.ID),
			).ID,
			appr:   false,
			status: models.OpenedStatus,
			err:    ErrUnauthorized,
		},
		{
			name:   "user-not-exists",
			ctx:    services.NewTestContext(t, e, u),
			target: stores.NewTestUser(t, e.UserStore),
			asset: stores.NewTestAsset(t,
				e.AssetStore,
				stores.TAWithOrg(org.ID),
			).ID,
			appr:   false,
			status: models.OpenedStatus,
			err:    ErrNotFound,
		},
		{
			name:   "random-user-appr",
			ctx:    services.NewTestContext(t, e, ru),
			target: stores.NewTestUser(t, e.UserStore),
			asset: stores.NewTestAsset(t,
				e.AssetStore,
				stores.TAWithOrg(org.ID),
			).ID,
			appr:   true,
			status: models.OpenedStatus,
			err:    ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// prerequisites
			p.RequestProjectCreation(
				services.NewTestContext(t, e, c.target),
				c.asset,
				org.ID,
			)

			targetu := c.target.ID
			if c.name == "user-not-exists" {
				// this hack is specifically for that
				// test cases. it is easier that way
				// to create prj request with regular
				// user and then process with fake ID.
				targetu = uuid.New()
			}

			err := p.ProcessProjectRequest(c.ctx, targetu, c.asset, c.appr)
			if !errors.Is(err, c.err) {
				t.Fatalf("expected err: %v, got: %v", c.err, err)
			}

			// check if token exists
			tk := new(models.Token)
			err = e.ProjectStore.DB().
				Where("user_id = ? AND purpose = ?", c.target.ID, models.CreateProjectToken).
				First(tk).Error
			if c.err == nil && err != nil && c.appr {
				t.Error("token not created")
			}

			// check if request change state
			var res models.ProjectCreationRequest
			p.st.DB().Where("asset_id = ? AND organization_id = ?", c.asset, org.ID).First(&res)

			if res.Status != c.status {
				t.Errorf("request status do not match: exp: %v, got: %v", res.Status, c.status)
			}

			if c.haveToken && res.Token == nil {
				t.Errorf("get prj request token %v, expected: %v", res.Token, tk.ID)
			}
		})
	}
}

func testCreate(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	contr := NewProject(e)

	admin := stores.NewTestAdmin(t, e.UserStore)

	// host organization
	orgu := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, orgu.ID)
	asset := stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))

	// guest organization
	guestu := stores.NewTestUser(t, e.UserStore)
	guesto := stores.NewTestOrg(t, e.OrganizationStore, guestu.ID)

	// Project templates
	randomguy := stores.NewTestUser(t, e.UserStore)

	cases := []struct {
		name   string
		ctx    context.Context
		reader io.Reader
		// token assumes that the TC rely on created token.
		token bool
		pre   func(uuid.UUID) error
		err   error
	}{

		{
			name: "su-default",
			ctx:  services.NewTestContext(t, e, admin),
			reader: strings.NewReader(
				fmt.Sprintf(projectTmpl, "su-default", org.ID, asset.ID, randomguy.ID, randomguy.ID),
			),
			token: false,
			pre:   func(u uuid.UUID) error { return nil },
		},
		{
			name: "lear-default",
			ctx:  services.NewTestContext(t, e, orgu),
			reader: strings.NewReader(
				fmt.Sprintf(projectTmpl, "lear-default", org.ID, asset.ID, randomguy.ID, randomguy.ID),
			),
			token: false,
			pre:   func(u uuid.UUID) error { return nil },
		},
		{
			name: "guest-org",
			ctx:  services.NewTestContext(t, e, guestu),
			reader: strings.NewReader(
				fmt.Sprintf(projectTmpl, "guest-org", guesto.ID, asset.ID, randomguy.ID, randomguy.ID),
			),
			token: true,
			pre: func(u uuid.UUID) error {
				stores.NewTestProjectCreationRequest(t,
					e.ProjectStore,
					stores.TPrjReqWithUser(guestu.ID),
					stores.TPrjReqWithAsset(asset.ID))

				ctx := services.NewTestContext(t, e, orgu)
				return contr.ProcessProjectRequest(ctx, u, asset.ID, true)

			},
		},
		{
			name: "guest org without request",
			ctx:  services.NewTestContext(t, e, guestu),
			reader: strings.NewReader(
				fmt.Sprintf(projectTmpl, "guest org without request", guesto.ID, asset.ID, randomguy.ID, randomguy.ID),
			),
			token: false,
			pre:   func(u uuid.UUID) error { return nil },
			err:   ErrUnauthorized,
		},
		{
			name: "guest-noauth",
			ctx:  services.NewTestContext(t, e, randomguy),
			reader: strings.NewReader(
				fmt.Sprintf(projectTmpl, "guest-noauth", guesto.ID, asset.ID, randomguy.ID, randomguy.ID),
			),
			token: false,
			pre:   func(u uuid.UUID) error { return nil },
			err:   ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// prerequisites
			us := services.FromContext(c.ctx).User
			if err := c.pre(us.ID); err != nil {
				t.Fatalf("fail to process project creation request: %v", err)
			}

			// call
			doc, _, err := contr.Create(c.ctx, c.reader)
			if err != c.err {
				t.Fatalf("got err: %v but expected: %v", err, c.err)
			}

			// check whether it is successful
			if doc != nil {
				if doc.Data.(*models.Project).Name != c.name {
					t.Fatalf("expected name: %s, got: %s", c.name, doc.Data)
				}
				if len(doc.Data.(*models.Project).Roles.PM) == 0 {
					t.Fatal("No pm assigned on create project")
				}

			}

			// additional checks for project request and token
			if !c.token {
				// default flow of creating project does not create token.
				return
			}

			tk := new(models.Token)
			contr.st.DB().
				Where("user_id = ?", us.ID).
				First(tk)

			if _, terr := contr.token.Get(c.ctx, models.CreateProjectToken, tk.ID); terr == nil {
				t.Errorf("token found but should not: %v", tk)
			}

			pr := new(models.ProjectCreationRequest)
			contr.st.DB().
				Where("user_id = ?", us.ID).
				First(pr)

			if pr.Status != models.AcceptedStatus {
				t.Fatalf("project request fails to invalidate: %v", pr)
			}
		})
	}
}

func testListByIDs(t *testing.T) {
	e := services.NewTestEnv(t)
	c := NewProject(e)
	prjs := []*models.Document{
		stores.NewTestProject(t, e.ProjectStore),
		stores.NewTestProject(t, e.ProjectStore),
		stores.NewTestProject(t, e.ProjectStore),
	}

	v, err := c.ListByIDs(context.Background(), prjs[0].ID, prjs[1].ID, prjs[2].ID)
	if err != nil {
		t.Fatal(err)
	}

	for i, prj := range prjs {
		if name := prj.Data.(*models.Project).Name; v[prj.ID].Name != name {
			t.Fatalf("Got project %d: %v; expected %v", i, v[prj.ID], name)
		}
	}
}

func testAdvanceToMilestone(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewProject(e)

	user := stores.NewTestUser(t, contr.st)
	milestone := models.MilestoneKickOffMeeting
	project := stores.NewTestProject(t, contr.st, stores.TPrjWithPm(user.ID))

	cases := []struct {
		name string
		ctx  context.Context
		err  error
		prj  uuid.UUID
	}{
		{
			name: "default",
			ctx:  services.NewTestContext(t, e, user),
			err:  nil,
			prj:  project.ID,
		},
		{
			name: "random user",
			ctx:  services.NewTestContext(t, e, stores.NewTestUser(t, contr.st)),
			err:  ErrUnauthorized,
			prj:  project.ID,
		},
		{
			name: "random project",
			ctx:  services.NewTestContext(t, e, user),
			err:  ErrUnauthorized,
			prj:  uuid.New(),
		},
	}

	for _, c := range cases {

		err := contr.AdvanceToMilestone(c.ctx, c.prj, milestone)
		if err != c.err {
			t.Errorf("got err: %v, but expected: %v", err, c.err)
		}

		if c.err != nil {
			continue
		}

		var res models.Project
		contr.st.DB().Where("id = ?", project.ID).First(&res)
		if res.Milestone != milestone {
			t.Errorf("milestone did not match")
		}
	}
}
func testCommentProject(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewProject(e)

	user := stores.NewTestUser(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(user.ID))

	randomu := stores.NewTestUser(t, e.UserStore)
	topic := "topic"

	cases := []struct {
		name  string
		err   error
		ctx   context.Context
		pid   uuid.UUID
		uid   uuid.UUID
		topic *string
	}{
		{
			name:  "default",
			err:   nil,
			ctx:   services.NewTestContext(t, e, user),
			pid:   prj.ID,
			uid:   user.ID,
			topic: &topic,
		},
		{
			name: "random user",
			err:  ErrUnauthorized,
			ctx:  services.NewTestContext(t, e, randomu),
			uid:  randomu.ID,
			pid:  prj.ID,
		},
		{
			name:  "no topic",
			err:   nil,
			ctx:   services.NewTestContext(t, e, user),
			pid:   prj.ID,
			uid:   user.ID,
			topic: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			prj, err := contr.CommentProject(c.ctx, c.pid, "content", c.topic)
			if err != c.err {
				t.Fatalf("exp error: %v, got: %v", c.err, err)
			}

			if c.err != nil {
				return
			}

			if len(prj.Comments) == 0 {
				t.Fatalf("missing comments")
			}

			comment := prj.Comments[0]

			if comment.Content != "content" {
				t.Fatalf("exp comment to be: \"content\", but got: %v", comment.Content)
			}

			if comment.Author.ID != c.uid {
				t.Fatalf("exp author to be: \"%v\", but got: %v", c.uid, comment.Author.ID)
			}
		})
	}
}

func testReports(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewProject(e)

	// anm
	anm := stores.NewTestAdminNwManager(t, e.UserStore)

	// random user
	u := stores.NewTestUser(t, e.UserStore)

	// user countries of CA role
	ca := stores.NewTestPortfolioRole(t, e.UserStore, models.CountryAdminRole, models.CountryBulgaria)

	lear := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)

	consru := stores.NewTestUser(t, e.OrganizationStore)
	consr := stores.NewTestOrg(t, e.OrganizationStore, consru.ID)

	escou := stores.NewTestUser(t, e.OrganizationStore)
	esco := stores.NewTestOrg(t, e.OrganizationStore, escou.ID)

	tl := makeTestProjects(t, e.AssetStore, org.ID, consr.ID, esco.ID)

	cases := []struct {
		name  string
		ctx   context.Context
		err   error
		count int
		// useful for pagination
		exp int
	}{
		{
			name:  "default",
			ctx:   services.NewTestContext(t, e, anm),
			count: tl,
			exp:   tl,
		},
		{
			name: "unauth",
			ctx:  services.NewTestContext(t, e, u),
			err:  ErrUnauthorized,
		},
		{
			name:  "ca",
			ctx:   services.NewTestContext(t, e, ca),
			count: 4,
			exp:   4,
		},
		{
			name:  "lear",
			ctx:   services.NewTestContext(t, e, lear),
			count: 4,
			exp:   4,
		},
		{
			name:  "consortium orgs",
			ctx:   services.NewTestContext(t, e, consru),
			count: 1,
			exp:   1,
		},
		{
			name:  "esco orgs",
			ctx:   services.NewTestContext(t, e, escou),
			count: 1,
			exp:   1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			docs, deps, total, err := contr.Reports(c.ctx, stores.Filter{})
			if err != c.err {
				t.Errorf("exp err: %v, but got: %v", c.err, err)
			}

			if total != c.count {
				t.Fatalf("expected count: %d, but got: %d", c.count, total)
			}

			if len(docs) != c.exp {
				t.Fatalf("expected docs: %d, but got: %d", c.exp, len(docs))
			}

			if len(deps) == 0 && c.err == nil {
				t.Fatal("deps are not preloaded properly")
			}
		})
	}
}

func makeTestProjects(t *testing.T, st stores.Store, org, consr, esco uuid.UUID) int {
	i := 0
	for i < 10 {
		if i%3 == 0 {
			stores.NewTestProject(t, st,
				stores.TPrjWithCountry(models.CountryBulgaria),
				stores.TPrjWithOrg(org),
			)
		} else {
			stores.NewTestProject(t, st)
		}

		i++
	}

	i++
	stores.NewTestProject(t, st,
		stores.TPrjWithCountry(models.CountryAustria),
		stores.TPrjWithConsrOrg(consr),
	)

	i++
	a := stores.NewTestAsset(t, st,
		stores.TAWithESCO(esco))
	stores.NewTestProject(t, st,
		stores.TPrjWithCountry(models.CountryBelgium),
		stores.TPrjWithAsset(a.ID),
	)

	return i
}
