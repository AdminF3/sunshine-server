package controller

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

func TestMilestones(t *testing.T) {
	t.Run("create_WP_MP", testCreateWPMP)
	t.Run("get_WP", testGetWP)
	t.Run("get_MP", testGetMP)
	t.Run("review_WP", submitWPReview)
	t.Run("review_MP", submitMPReview)
}

func testCreateWPMP(t *testing.T) {
	e := services.NewTestEnv(t)

	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	any := gomock.Any()
	not.EXPECT().Broadcast(any, any, any, any, any, any, any, any).AnyTimes()

	wp := NewWorkPhase(e)
	mp := NewMonitoringPhase(e)
	ustore := wp.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	prj := stores.NewTestProject(t, wp.store.FromKind("project"))

	cases := []struct {
		name  string
		ctx   context.Context
		wperr error
		mperr error
		proj  uuid.UUID
	}{
		{
			name:  "ok su",
			ctx:   services.NewTestContext(t, e, su),
			wperr: nil,
			mperr: nil,
			proj:  prj.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			work, err := wp.AdvanceToWorkPhase(c.ctx, c.proj)
			if err != c.wperr {
				t.Errorf("expected err: %v;\n got: %v", c.wperr, err)
			}
			wph := work.Data.(*models.WorkPhase)
			if wph.Project != c.proj {
				t.Errorf("expected projID: %v;\n got: %v", c.proj,
					work.Data.(*models.WorkPhase).Project)
			}

			monitoring, err := mp.AdvanceToMonitoringPhase(c.ctx, c.proj)
			if err != c.mperr {
				t.Errorf("expected err: %v;\n got: %v", c.mperr, err)
			}
			if monitoring.Data.(*models.MonitoringPhase).Project != c.proj {
				t.Errorf("expected projID: %v;\n got: %v", c.proj,
					monitoring.Data.(*models.MonitoringPhase).Project)
			}
		})
	}
}

func testGetWP(t *testing.T) {
	e := services.NewTestEnv(t)

	wp := NewWorkPhase(e)
	ustore := wp.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	prj := stores.NewTestProject(t, wp.store.FromKind("project"))
	gw := stores.NewTestWorkPhase(t, wp.store.FromKind("work_phase"), prj.ID)

	cases := []struct {
		name  string
		ctx   context.Context
		wperr error
		proj  uuid.UUID
	}{
		{
			name:  "ok su",
			ctx:   services.NewTestContext(t, e, su),
			wperr: nil,
			proj:  prj.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			work, _, err := wp.GetWP(c.ctx, gw.ID)
			if err != c.wperr {
				t.Errorf("expected err: %v;\n got: %v", c.wperr, err)
			}
			if work.ID != gw.ID {
				t.Errorf("expected projID: %v;\n got: %v", c.proj,
					work.Data.(*models.WorkPhase).Project)
			}
		})
	}
}

func testGetMP(t *testing.T) {
	e := services.NewTestEnv(t)

	mp := NewMonitoringPhase(e)
	ustore := mp.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	prj := stores.NewTestProject(t, mp.store.FromKind("project"))
	gm := stores.NewTestMonitoringPhase(t, mp.store.FromKind("monitoring_phase"), prj.ID)

	cases := []struct {
		name  string
		ctx   context.Context
		mperr error
		proj  uuid.UUID
	}{
		{
			name:  "ok su",
			ctx:   services.NewTestContext(t, e, su),
			mperr: nil,
			proj:  prj.ID,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			monitoring, _, merr := mp.GetMP(c.ctx, gm.ID)
			if merr != c.mperr {
				t.Errorf("expected err: %v;\n got: %v", c.mperr, merr)
			}
			if monitoring.ID != gm.ID {
				t.Errorf("expected projID: %v;\n got: %v", c.proj,
					monitoring.Data.(*models.MonitoringPhase).Project)
			}
		})
	}

}

func submitWPReview(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewWorkPhase(e)

	pm := stores.NewTestUser(t, e.UserStore)
	randomu := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore,
		stores.TPrjWithPm(pm.ID))

	wp := stores.NewTestWorkPhase(t, e.WPStore, prj.ID)

	var rID uuid.UUID
	for _, v := range wp.Reviews {
		if v.Type == models.WPReviewTypeFinancial {
			rID = v.ID
			break
		}
	}

	cases := []struct {
		name  string
		ctx   context.Context
		wpid  uuid.UUID
		revid uuid.UUID
		error error
	}{
		{
			name:  "ok",
			ctx:   services.NewTestContext(t, e, pm),
			wpid:  wp.ID,
			revid: rID,
		},
		{
			name: "ok-additional-review",
			ctx:  services.NewTestContext(t, e, pm),
			wpid: wp.ID,
		},
		{
			name:  "unauthorized",
			ctx:   services.NewTestContext(t, e, randomu),
			wpid:  wp.ID,
			error: ErrUnauthorized,
		},
		{
			name:  "wp-not-found",
			ctx:   services.NewTestContext(t, e, pm),
			wpid:  uuid.New(),
			error: ErrNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expectedRevNum := 5
			r := models.WPReview{
				Type:    models.WPReviewTypeFinancial,
				Comment: "Lorem Ipsum",
			}
			// If review.ID provided in ReviewWP request, controller
			// should update, rather than create a new review.
			if c.revid != uuid.Nil {
				r.ID = c.revid
				expectedRevNum = 4
			}
			err := contr.ReviewWP(c.ctx, c.wpid, r)

			if !isError(err, c.error) {
				t.Fatalf("expected err: %v, but got: %v", c.error, err)
			}

			if err != nil {
				return
			}

			var res models.WorkPhase
			e.WPStore.DB().Preload("Reviews").Where("id = ? ", c.wpid).Find(&res)

			if len(res.Reviews) != expectedRevNum {
				t.Fatalf("reviews expected to be %v but got: %v", expectedRevNum, len(res.Reviews))
			}

			for _, r := range res.Reviews {
				if r.Type == models.WPReviewTypeFinancial {
					if r.Comment != "Lorem Ipsum" {
						t.Fatalf("expected comment to be updated with :%v but got: %v", "Lorem Ipsum", r.Comment)
					}

					tmpu := services.FromContext(c.ctx).User.ID
					if *r.Author != tmpu {
						t.Fatalf("author is not updated: %v, but got: %v", r.Author.String(), tmpu.String())
					}
				}
			}
		})
	}
}

func submitMPReview(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewMonitoringPhase(e)

	pm := stores.NewTestUser(t, e.UserStore)
	randomu := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore,
		stores.TPrjWithPm(pm.ID))

	mp := stores.NewTestMonitoringPhase(t, e.MPStore, prj.ID)
	var rID uuid.UUID
	for _, v := range mp.Reviews {
		if v.Type == models.MPReviewTypeForfaiting {
			rID = v.ID
			break
		}
	}

	cases := []struct {
		name  string
		ctx   context.Context
		mpid  uuid.UUID
		revid uuid.UUID
		error error
	}{
		{
			name:  "ok",
			ctx:   services.NewTestContext(t, e, pm),
			mpid:  mp.ID,
			revid: rID,
		},
		{
			name:  "unauthorized",
			ctx:   services.NewTestContext(t, e, randomu),
			mpid:  mp.ID,
			error: ErrUnauthorized,
		},
		{
			name:  "wp-not-found",
			ctx:   services.NewTestContext(t, e, pm),
			mpid:  uuid.New(),
			error: ErrNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expectedRevNum := 31
			r := models.MPReview{
				Type:    models.MPReviewTypeForfaiting,
				Comment: "Lorem Ipsum",
			}
			// If review.ID provided in ReviewMP request, controller
			// should update, rather than create a new review.
			if c.revid != uuid.Nil {
				r.ID = c.revid
				expectedRevNum = 30
			}

			err := contr.ReviewMP(c.ctx, c.mpid, r)

			if !isError(err, c.error) {
				t.Fatalf("expected err: %v, but got: %v", c.error, err)
			}

			if err != nil {
				return
			}

			var res models.MonitoringPhase
			e.MPStore.DB().Preload("Reviews").Where("id = ? ", c.mpid).Find(&res)

			if len(res.Reviews) != expectedRevNum {
				t.Fatalf("reviews expected to be %v but got: %v", expectedRevNum, len(res.Reviews))
			}

			for _, r := range res.Reviews {
				if r.Type == models.MPReviewTypeForfaiting {
					if r.Comment == "Lorem Ipsum" {
						tmpu := services.FromContext(c.ctx).User.ID
						if *r.Author != tmpu {
							t.Fatalf("author is not updated: %v, but got: %v", r.Author.String(), tmpu.String())
						}
						return
					}
				}
			}
			t.Fatalf("expected at least one review comment to be updated with :%v but none were", "Lorem Ipsum")
		})
	}
}
