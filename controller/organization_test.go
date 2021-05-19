package controller

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

const fmtValidateOrg = `{"valid": 3}`

// TestUpdateValidateOrganization validates that if update occurs, it
// will invalidate the organization
func TestUpdateValidateOrganization(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	o := NewOrganization(e)
	ustore := o.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	lear := stores.NewTestUser(t, ustore)
	u := stores.NewTestUser(t, ustore)

	org := stores.NewTestOrg(t, o.store.FromKind("organization"), lear.ID)
	org2 := stores.NewTestOrg(t, o.store.FromKind("organization"))

	cases := []struct {
		name     string
		ctx      context.Context
		org      uuid.UUID
		expected error
		valid    models.ValidationStatus
		body     io.Reader
	}{
		{
			name:     "lear validates own org",
			ctx:      services.NewTestContext(t, e, lear),
			org:      org.ID,
			expected: nil,
			valid:    models.ValidationStatusValid,
			body:     strings.NewReader(fmtValidateOrg),
		},
		{
			name:     "admin validates",
			ctx:      services.NewTestContext(t, e, su),
			org:      org2.ID,
			expected: nil,
			valid:    models.ValidationStatusDeclined,
			body:     strings.NewReader(fmtValidateOrg),
		},
		{
			name:     "lear validates unrelated org",
			ctx:      services.NewTestContext(t, e, lear),
			org:      org2.ID,
			expected: ErrUnauthorized,
			body:     strings.NewReader(fmtValidateOrg),
		},
		{
			name:     "random user validates",
			ctx:      services.NewTestContext(t, e, u),
			org:      org2.ID,
			expected: ErrUnauthorized,
			body:     strings.NewReader(fmtValidateOrg),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			updated, _, err := o.Update(c.ctx, c.org, c.body)
			if err != c.expected {
				t.Errorf("expected: %v, got: %v updated: %v", c.expected, err, updated)
			}
			if err == nil && updated.Data.(*models.Organization).Valid != c.valid {
				t.Errorf("expexted status: %v, got status: %v", c.valid, updated.Data.(*models.Organization).Valid)
			}
		})
	}
}

// TestValidateOrganization test validate organization.
func TestValidateOrganization(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	any := gomock.Any()
	not.EXPECT().Broadcast(any, any, any, any, any, any, any, any).AnyTimes()

	o := NewOrganization(e)
	ustore := o.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	ca := stores.NewTestPortfolioRole(t, ustore, models.CountryAdminRole, models.CountryLatvia)
	lear := stores.NewTestUser(t, ustore)

	org := stores.NewTestOrg(t, o.store.FromKind("organization"), lear.ID)
	asset := stores.NewTestAsset(t, o.store)

	org.Data.(*models.Organization).VAT = fmt.Sprintf("RC_%s", asset.ID.String())
	o.store.Update(services.NewTestContext(t, e, ca), org)

	cmnt := "da"
	cases := []struct {
		name      string
		ctx       context.Context
		org       uuid.UUID
		status    models.ValidationStatus
		oldStatus models.ValidationStatus
		expected  error
		comment   *string
	}{
		{
			name:     "ok ca",
			ctx:      services.NewTestContext(t, e, ca),
			org:      org.ID,
			status:   models.ValidationStatusRegistered,
			expected: nil,
		},
		{
			name:     "ok su",
			ctx:      services.NewTestContext(t, e, su),
			org:      org.ID,
			status:   models.ValidationStatusDeclined,
			expected: nil,
		},
		{
			name:      "lear unauth",
			ctx:       services.NewTestContext(t, e, lear),
			org:       org.ID,
			status:    models.ValidationStatusValid,
			oldStatus: models.ValidationStatusDeclined,
			expected:  ErrUnauthorized,
		},
		{
			name:     "with comment",
			ctx:      services.NewTestContext(t, e, ca),
			org:      org.ID,
			status:   models.ValidationStatusRegistered,
			expected: nil,
			comment:  &cmnt,
		},
		{
			name:     "assign community org",
			ctx:      services.NewTestContext(t, e, ca),
			org:      org.ID,
			status:   models.ValidationStatusValid,
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := o.Validate(c.ctx, c.org, c.status, c.comment)
			if err != c.expected {
				t.Errorf("expected %v err; got %v", c.expected, err)
			}

			if c.status == models.ValidationStatusValid && err == nil {
				var as models.Asset
				aid := strings.Split(org.Data.(*models.Organization).VAT, "_")[1]
				o.store.DB().Where("id = ?", aid).First(&as)

				if *as.ESCO != org.ID {
					t.Errorf("assign community org fails: %v", err)
				}
			}

			upd, _, err := o.Get(c.ctx, c.org)
			if err != nil {
				t.Errorf("could not fetch updated org; got %v", err)
			}
			if c.expected != nil {
				c.status = c.oldStatus
			}
			if upd.Data.(*models.Organization).Valid != c.status {
				t.Errorf("expected status %v; got %v", c.status, upd.Data.(*models.Organization).Valid)
			}

		})
	}
}

func TestGetOrganizationsReport(t *testing.T) {
	e := services.NewTestEnv(t)
	contr := NewOrganization(e)

	u := stores.NewTestAdminNwManager(t, e.UserStore)

	expTotal, targetOrg := generateDummyData(t, e)
	learID := targetOrg.OrganizationRoles[0].UserID
	lear, _ := e.UserStore.Get(context.Background(), learID)

	cases := []struct {
		name    string
		ctx     context.Context
		first   int
		offset  int
		total   int
		owned   []int
		related []int
	}{
		{
			name:    "all",
			ctx:     services.NewTestContext(t, e, u),
			first:   0,
			offset:  0,
			total:   expTotal,
			owned:   []int{2, 1, 1, 1},
			related: []int{1, 0, 1, 0},
		},
		{
			name:   "filter by pm",
			ctx:    services.NewTestContext(t, e, lear),
			first:  0,
			offset: 0,
			total:  1,
		},
		{
			name:   "subsection",
			ctx:    services.NewTestContext(t, e, u),
			first:  1,
			offset: 1,
			total:  1,
		},
	}

	prj := func(t *testing.T, r models.OrganizationProjectsReport, tp string, c []int) {
		if c == nil {
			return
		}

		if r.TotalCount != c[0] {
			t.Fatalf("from %s: total projects Count %d but expected: %d", tp, r.TotalCount, c[0])
		}

		if r.OngoingCount != c[1] {
			t.Fatalf("from %s: Ongoing count projects are %d but ext to be: %d", tp, r.OngoingCount, c[1])
		}

		if r.ApprovedForfaitingCount != c[2] {
			t.Errorf("from %s: Approved for forfaiting are %d but expected to be %d", tp, r.ApprovedForfaitingCount, c[2])
		}
		if r.MonitoringPhaseCount != c[3] {
			t.Errorf("from %s: Monitoring phases pojects are %d but expected to be %d", tp, r.ApprovedForfaitingCount, c[3])
		}

	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reports, _, err := contr.GetReport(c.ctx, c.first, c.offset)
			if err != nil {
				t.Fatal(err)
			}

			if len(reports) != c.total {
				t.Fatalf("expected reports to be %d, but got %d", c.total, len(reports))
			}

			var targetReport models.OrganizationReport
			for _, rep := range reports {
				if rep.ID == targetOrg.ID {
					targetReport = rep
				}
			}

			if !strings.Contains(targetReport.LearName, "John Doe") {
				t.Fatalf("lear name (John Doe) not found: %s", targetReport.LearName)
			}

			if !strings.Contains(targetReport.LearEmail, "john") {
				t.Fatalf("lear email (john_<uuid>)not found: %s", targetReport.LearEmail)
			}

			if targetReport.UsersCount != 1 {
				t.Fatalf("user count (%d) expected: %d", 1, targetReport.UsersCount)

			}

			prj(t, targetReport.OwnProjects, "owns", c.owned)
			prj(t, targetReport.RelatedProjects, "related", c.related)
		})
	}
}

func generateDummyData(t *testing.T, env *services.Env) (int, *models.Organization) {
	t.Helper()

	count := 2

	var org1, org2 *models.Document

	fareview := models.FAReview{
		Type:     models.FAReviewTypeExecutive,
		Approved: true,
		Author:   &stores.NewTestUser(t, env.UserStore).ID,
	}

	for i := 0; i < count; i++ {
		o := stores.NewTestOrg(t, env.OrganizationStore)
		_ = newTestPrj(t, env.ProjectStore, o.ID, uuid.Nil, models.ProjectStatusPlanning, true, fareview)
		pid := newTestPrj(t, env.ProjectStore, o.ID, uuid.Nil, models.ProjectStatusInProgress, false)
		stores.NewTestMonitoringPhase(t, env.OrganizationStore, pid)
		if i%2 == 0 {
			org1 = o
		} else {
			org2 = o
		}
	}

	// one time only
	// related
	a := newTestAsset(t, env.AssetStore, org2.ID)
	newTestPrj(t, env.ProjectStore, org1.ID, a, models.ProjectStatusPlanning, true, fareview)

	a2 := newTestAsset(t, env.AssetStore, org1.ID)
	newTestPrj(t, env.ProjectStore, org2.ID, a2, models.ProjectStatusPlanning, true)

	org1.Data.(*models.Organization).Country = models.CountryBulgaria
	env.OrganizationStore.Update(context.Background(), org1)

	return count, org1.Data.(*models.Organization)
}

func newTestPrj(t *testing.T,
	st stores.Store,
	oid, aid uuid.UUID,
	status models.ProjectStatus,
	accepted bool,
	reviews ...models.FAReview) uuid.UUID {
	t.Helper()

	p := models.Project{
		Name:              uuid.New().String(),
		Owner:             oid,
		Asset:             newTestAsset(t, st, oid),
		Status:            status,
		AirTemperature:    20,
		WaterTemperature:  40,
		GuaranteedSavings: 51.16,
		Country:           models.CountryLatvia,
		PortfolioDirector: stores.NewTestAdmin(t, st).ID,
		Milestone:         models.MilestoneAcquisitionMeeting,
	}

	if aid != uuid.Nil {
		p.Asset = aid
	}

	prj, err := st.Create(context.Background(), &p)
	if err != nil {
		t.Fatalf("fail to create project, %v", err)
	}

	manager := stores.NewTestUser(t, st)
	fa := models.ForfaitingApplication{
		Project:     prj.ID,
		ManagerID:   manager.ID,
		Manager:     *manager.Data.(*models.User),
		PrivateBond: false,
	}

	if len(reviews) > 0 {
		fa.Reviews = reviews
	}

	_, err = st.FromKind("forfaiting_application").Create(context.Background(), &fa)
	if err != nil {
		t.Fatalf("fail to create fa: %v", err)
	}
	return prj.ID
}

func newTestAsset(t *testing.T, st stores.Store, oid uuid.UUID) uuid.UUID {
	t.Helper()

	store := st.FromKind("asset")
	category := models.NROfficeBuildings

	a := models.Asset{
		Owner:   oid,
		Address: "End of the world",
		Coordinates: models.Coords{
			Lat: rand.Float32(),
			Lng: rand.Float32(),
		},
		Area:         9000,
		HeatedArea:   512,
		BillingArea:  8500,
		Flats:        250,
		Floors:       42,
		StairCases:   84,
		BuildingType: models.BuildingType318,
		HeatingType:  models.HeatingDistrict,
		Cadastre:     uuid.New().String(),
		Valid:        models.ValidationStatusValid,
		Country:      models.CountryLatvia,
		Category:     &category,
	}

	d, err := store.Create(context.Background(), &a)
	if err != nil {
		t.Fatalf("create test asset: %s", err)
	}
	return d.ID
}

func TestAcceptLearApplication(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	o := NewOrganization(e)
	ustore := o.store.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	lear := stores.NewTestUser(t, ustore)
	u := stores.NewTestUser(t, ustore)

	org := stores.NewTestOrg(t, o.store.FromKind("organization"), lear.ID)

	cases := []struct {
		name     string
		ctx      context.Context
		org      uuid.UUID
		expected error
	}{
		{
			name:     "lear",
			ctx:      services.NewTestContext(t, e, lear),
			org:      org.ID,
			expected: nil,
		},
		{
			name:     "admin",
			ctx:      services.NewTestContext(t, e, su),
			org:      org.ID,
			expected: nil,
		},
		{
			name:     "random user",
			ctx:      services.NewTestContext(t, e, u),
			org:      org.ID,
			expected: ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := o.AcceptLEARApplication(c.ctx, u.ID, c.org, "test", "testfile", true)
			if err != c.expected {
				t.Errorf("expected: %v, got: %v", c.expected, err)
			}
		})
	}

}
