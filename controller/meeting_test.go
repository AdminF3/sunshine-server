package controller

import (
	"context"
	"testing"
	"time"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func TestMeetings(t *testing.T) {
	t.Run("create", testCreateDeleteMeeting)
	t.Run("get", testGetMeeting)
	t.Run("update", testUpdateMeeting)
}

func testCreateDeleteMeeting(t *testing.T) {
	e := services.NewTestEnv(t)

	m := NewMeeting(e)
	admin := stores.NewTestAdmin(t, e.UserStore)
	lear := stores.NewTestUser(t, e.UserStore)
	random := stores.NewTestUser(t, e.UserStore)

	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)
	m1 := models.Meeting{
		Name:        "da",
		Host:        org.ID,
		Location:    "da",
		Date:        time.Now(),
		Stakeholder: models.LegalFormNGO,
	}

	m2 := models.Meeting{
		Name:        "daa",
		Host:        org.ID,
		Location:    "daa",
		Date:        time.Now(),
		Stakeholder: models.LegalFormNGO,
	}

	cases := []struct {
		name    string
		ctx     context.Context
		meeting *models.Meeting
		err     error
	}{
		{
			name:    "admin",
			ctx:     services.NewTestContext(t, e, admin),
			meeting: &m1,
			err:     nil,
		},
		{
			name:    "lear",
			ctx:     services.NewTestContext(t, e, lear),
			meeting: &m2,
			err:     nil,
		},
		{
			name:    "unauth",
			ctx:     services.NewTestContext(t, e, random),
			meeting: &m2,
			err:     ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mm, err := m.Create(c.ctx, c.meeting)
			if err != nil && err != c.err {
				t.Fatalf("expected: %v; got: %v", c.err, err)
			}
			if err == nil {
				if mm.Data.(*models.Meeting).Name != c.meeting.Name {
					t.Fatalf("expected created meeting's name to be da")
				}

				derr := m.Delete(c.ctx, mm.ID)
				if derr != nil {
					t.Fatalf("got %v err on del meeting", derr)
				}
			}
		})
	}
}

func testGetMeeting(t *testing.T) {
	e := services.NewTestEnv(t)

	m := NewMeeting(e)
	admin := stores.NewTestAdmin(t, e.UserStore)
	adminctx := services.NewTestContext(t, e, admin)
	lear := stores.NewTestUser(t, e.UserStore)
	rand := stores.NewTestUser(t, e.UserStore)

	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)
	meet := stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithOrg(org.ID))

	// project consortium case
	prj2 := stores.NewTestProject(t, e.ProjectStore)
	prj2.Data.(*models.Project).ConsortiumOrgs = []string{org.ID.String()}
	if _, err := e.ProjectStore.Update(adminctx, prj2); err != nil {
		t.Errorf("Fail to add consortium org to project: %v", err)
	}
	meet2 := stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithPrj(prj2.ID))

	cases := []struct {
		name string
		ctx  context.Context
		meet uuid.UUID
		err  error
	}{
		{
			name: "admin",
			ctx:  adminctx,
			meet: meet.ID,
			err:  nil,
		},
		{
			name: "lear",
			ctx:  services.NewTestContext(t, e, lear),
			meet: meet.ID,
			err:  nil,
		},
		{
			name: "unauth",
			ctx:  services.NewTestContext(t, e, rand),
			meet: meet.ID,
			err:  ErrUnauthorized,
		},
		{
			name: "consortium get",
			ctx:  services.NewTestContext(t, e, lear),
			meet: meet2.ID,
			err:  nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mm, err := m.Get(c.ctx, c.meet)
			if err != nil && err != c.err {
				t.Errorf("expected: %v; got: %v", c.err, err)
			}
			if c.err == nil {
				if mm.Data.(*models.Meeting).Name != "init meeting" {
					t.Errorf("expected different guests")
				}
			}
		})
	}
}

func testUpdateMeeting(t *testing.T) {
	e := services.NewTestEnv(t)

	m := NewMeeting(e)
	admin := stores.NewTestAdmin(t, e.UserStore)
	lear := stores.NewTestUser(t, e.UserStore)
	rand := stores.NewTestUser(t, e.UserStore)

	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)
	meet := stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithOrg(org.ID))

	cases := []struct {
		name string
		ctx  context.Context
		new  string
		err  error
	}{
		{
			name: "admin",
			ctx:  services.NewTestContext(t, e, admin),
			new:  "da",
			err:  nil,
		},
		{
			name: "lear",
			ctx:  services.NewTestContext(t, e, lear),
			new:  "ne",
			err:  nil,
		},
		{
			name: "unauth",
			ctx:  services.NewTestContext(t, e, rand),
			new:  "fail",
			err:  ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			meet.Name = c.new
			upd, err := m.Update(c.ctx, meet)
			if err != nil && err != c.err {
				t.Errorf("expected: %v; got: %v", c.err, err)
			}
			if c.err == nil {
				if upd.Data.(*models.Meeting).Name != c.new {
					t.Errorf("expected meet name to be %v; got %v", c.new, upd.Data.(*models.Meeting).Name)
				}
			}
		})
	}
}
