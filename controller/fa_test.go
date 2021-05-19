package controller

import (
	"context"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func TestFA(t *testing.T) {
	t.Run("create", createFA)
	t.Run("submit review", submitReview)
	t.Run("update", updateFA)
	t.Run("createFP", createFP)
	t.Run("updateFP", updateFP)
}

func createFA(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewForfaitingAgreement(e)

	admin := stores.NewTestAdmin(t, e.UserStore)
	u := NewUser(e)
	pm := stores.NewTestUser(t, e.UserStore)
	randomu := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore,
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation),
		stores.TPrjWithPm(pm.ID))

	randomprj := stores.NewTestProject(t, e.ProjectStore)

	fa := models.ForfaitingApplication{
		ManagerID:   admin.ID,
		Finance:     models.FinanceBankFunding,
		PrivateBond: true,
		BankAccount: models.BankAccount{
			IBAN:            "iban",
			BankNameAddress: "end of the world",
			BeneficiaryName: "John Doe",
			SWIFT:           "BC1234XX",
		},
	}

	cases := []struct {
		name  string
		error error
		ctx   context.Context
		prj   uuid.UUID
	}{
		{
			name: "admin",
			ctx:  services.NewTestContext(t, e, admin),
			prj:  prj.ID,
		},
		{
			name:  "same-project-id",
			ctx:   services.NewTestContext(t, e, admin),
			prj:   prj.ID,
			error: ErrDuplicate,
		},
		{
			name:  "unauthorized",
			error: ErrUnauthorized,
			ctx:   context.Background(),
			prj:   prj.ID,
		},
		{
			name:  "random-user",
			ctx:   services.NewTestContext(t, e, randomu),
			error: ErrUnauthorized,
			prj:   prj.ID,
		},
		{
			name:  "early-milestone",
			ctx:   services.NewTestContext(t, e, admin),
			prj:   randomprj.ID,
			error: ErrBadInput,
		},
		{
			name:  "project-not-found",
			ctx:   services.NewTestContext(t, e, admin),
			prj:   uuid.New(),
			error: ErrNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fa.Project = c.prj

			fares, err := contr.Create(c.ctx, fa)
			if !isError(err, c.error) {
				t.Fatalf("expected err: %v, but got: %v", c.error, err)
			}

			if err != nil {
				return
			}

			if len(fares.Reviews) != 4 {
				t.Errorf("expected 4 reviews, but got: %d", len(fares.Reviews))
			}

			// Check for tama role for forfaiting manager
			usd, _, err := u.st.Unwrap(c.ctx, fares.ManagerID)
			if err != nil {
				t.Errorf("getting forfaiting manager failed with: %v", err)
			}

			usr := usd.Data.(*models.User)
			isTama := false
			for _, r := range usr.ProjectRoles {
				if r.Position == "tama" {
					isTama = true
				}
			}
			if !isTama {
				t.Errorf(" the forfaitin manager is missing tama role")
			}
		})
	}
}

func submitReview(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewForfaitingAgreement(e)

	pm := stores.NewTestUser(t, e.UserStore)
	randomu := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore,
		stores.TPrjWithPm(pm.ID))

	fa := stores.NewTestFA(t, e.FAStore, stores.TFAWithProject(prj.ID))
	var revid uuid.UUID
	for _, v := range fa.Reviews {
		if v.Type == models.FAReviewTypeFinancial {
			revid = v.ID
			break
		}
	}

	cases := []struct {
		name  string
		ctx   context.Context
		faid  uuid.UUID
		revid uuid.UUID
		error error
	}{
		{
			name:  "ok",
			ctx:   services.NewTestContext(t, e, pm),
			faid:  fa.ID,
			revid: revid,
		},
		{
			name: "ok-additional-review",
			ctx:  services.NewTestContext(t, e, pm),
			faid: fa.ID,
		},
		{
			name:  "unauthorized",
			ctx:   services.NewTestContext(t, e, randomu),
			faid:  fa.ID,
			error: ErrUnauthorized,
		},
		{
			name:  "fa-not-found",
			ctx:   services.NewTestContext(t, e, pm),
			faid:  uuid.New(),
			error: ErrNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expectedRevNum := 5
			r := models.FAReview{
				Type:    models.FAReviewTypeFinancial,
				Comment: "Lorem Ipsum",
			}
			// If review.ID provided in ReviewWP request, controller
			// should update, rather than create a new review.
			if c.revid != uuid.Nil {
				r.ID = c.revid
				expectedRevNum = 4
			}

			err := contr.Review(c.ctx, c.faid, r)

			if !isError(err, c.error) {
				t.Fatalf("expected err: %v, but got: %v", c.error, err)
			}

			if err != nil {
				return
			}

			var res models.ForfaitingApplication
			e.FAStore.DB().Preload("Reviews").Where("id = ? ", c.faid).Find(&res)

			if len(res.Reviews) != expectedRevNum {
				t.Fatalf("reviews expected to be %v but got: %v", expectedRevNum, len(res.Reviews))
			}

			for _, r := range res.Reviews {
				if r.Type == models.FAReviewTypeFinancial {
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

func updateFA(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewForfaitingAgreement(e)

	pm := stores.NewTestUser(t, e.UserStore)
	randomu := stores.NewTestUser(t, e.UserStore)
	u := NewUser(e)

	prj := stores.NewTestProject(t, e.ProjectStore,
		stores.TPrjWithPm(pm.ID))

	fa := stores.NewTestFA(t, e.FAStore, stores.TFAWithProject(prj.ID))

	cases := []struct {
		name  string
		ctx   context.Context
		faid  uuid.UUID
		error error
	}{
		{
			name: "ok",
			ctx:  services.NewTestContext(t, e, pm),
			faid: fa.ID,
		},
		{
			name: "second update",
			ctx:  services.NewTestContext(t, e, pm),
			faid: fa.ID,
		},
		{
			name:  "unauth",
			ctx:   services.NewTestContext(t, e, randomu),
			faid:  fa.ID,
			error: ErrUnauthorized,
		},
		{
			name:  "fa not found",
			ctx:   services.NewTestContext(t, e, pm),
			faid:  uuid.New(),
			error: ErrNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := contr.Update(c.ctx, c.faid, models.ForfaitingApplication{
				PrivateBond: true,
				BankAccount: models.BankAccount{IBAN: "BG1818UNCR0111"},
				Finance:     models.FinanceOther,
			})

			if !isError(err, c.error) {
				t.Fatalf("expected err: %v, but got: %v", c.error, err)
			}

			if err != nil {
				return
			}

			if res.PrivateBond != true {
				t.Fatalf("expected privatebond to be %v, but got: %v", true, res.PrivateBond)
			}

			if res.BankAccount.IBAN != "BG1818UNCR0111" {
				t.Fatalf("expected privatebond to be %v, but got: %v", "BG1818UNCR0111", res.BankAccount.IBAN)
			}

			if res.Finance != models.FinanceOther {
				t.Fatalf("expected finance to be %v, but got: %v", models.FinanceOther, res.Finance)
			}

			// Check for tama role for forfaiting manager
			usd, _, err := u.st.Unwrap(c.ctx, res.ManagerID)
			if err != nil {
				t.Errorf("getting forfaiting manager failed with: %v", err)
			}

			usr := usd.Data.(*models.User)
			isTama := false
			for _, r := range usr.ProjectRoles {
				if r.Position == "tama" {
					isTama = true
				}
			}
			if !isTama {
				t.Errorf(" the forfaiting manager is missing tama role")
			}
			var cnt int
			contr.st.DB().Table("project_roles").
				Where("user_id = ? AND project_id = ? AND position = ? ",
					res.ManagerID, res.Project, "tama").
				Count(&cnt)
			if cnt != 1 {
				t.Errorf("Invalid count for managers with tama roles expected %d, got %d", 1, cnt)
			}
		})
	}
}

func isError(err, target error) bool {
	if err == target {
		return true
	}

	errs := ""
	errt := ""

	if err != nil {
		errs = err.Error()
	}

	if target != nil {
		errt = target.Error()
	}

	if errt == "" {
		return false
	}

	return strings.Contains(errs, errt)
}

func createFP(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewForfaitingAgreement(e)

	admin := stores.NewTestAdmin(t, e.UserStore)
	pm := stores.NewTestUser(t, e.UserStore)
	randomu := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore,
		stores.TPrjWithMilestone(models.MilestoneProjectPreparation),
		stores.TPrjWithPm(pm.ID))

	cases := []struct {
		name  string
		error error
		ctx   context.Context
		prj   uuid.UUID
	}{
		{
			name: "admin",
			ctx:  services.NewTestContext(t, e, admin),
			prj:  prj.ID,
		},
		{
			name:  "unauthorized",
			ctx:   context.Background(),
			prj:   prj.ID,
			error: ErrUnauthorized,
		},
		{
			name:  "random-user",
			ctx:   services.NewTestContext(t, e, randomu),
			error: ErrUnauthorized,
			prj:   prj.ID,
		},
		{
			name:  "project-not-found",
			ctx:   services.NewTestContext(t, e, admin),
			prj:   uuid.New(),
			error: ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fpres, err := contr.CreateFP(c.ctx, 123, models.CurrencyEUR, c.prj, nil)
			if !isError(err, c.error) {
				t.Fatalf("expected err: %v, but got: %v", c.error, err)
			}

			if err != nil {
				return
			}

			if fpres.TransferValue != 123 || fpres.Currency != models.CurrencyEUR {
				t.Fatalf("got different result than expected; got %v", fpres)
			}
		})
	}
}

func updateFP(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewForfaitingAgreement(e)

	pm := stores.NewTestUser(t, e.UserStore)
	randomu := stores.NewTestUser(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore,
		stores.TPrjWithPm(pm.ID))

	fp := stores.NewTestFP(t, e.FPStore, stores.TFPWithProject(prj.ID))
	tv := 8999
	lev := models.CurrencyBGN

	cases := []struct {
		name  string
		ctx   context.Context
		fpid  uuid.UUID
		error error
	}{
		{
			name: "ok",
			ctx:  services.NewTestContext(t, e, pm),
			fpid: fp.ID,
		},
		{
			name:  "unauth",
			ctx:   services.NewTestContext(t, e, randomu),
			fpid:  fp.ID,
			error: ErrUnauthorized,
		},
		{
			name:  "fp not found",
			ctx:   services.NewTestContext(t, e, pm),
			fpid:  uuid.New(),
			error: ErrNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := contr.UpdateFP(c.ctx, c.fpid, prj.ID, &tv, &lev, nil)

			if !isError(err, c.error) {
				t.Fatalf("expected err: %v, but got: %v", c.error, err)
			}

			if err != nil {
				return
			}

			if res.Currency != models.CurrencyBGN {
				t.Fatalf("expected currency to be updated; got: %v", res.Currency)
			}

			if res.TransferValue != tv {
				t.Fatalf("expected transfer value to be updated; got: %v", res.TransferValue)
			}
		})
	}
}
