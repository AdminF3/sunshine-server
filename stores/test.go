package stores

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var (
	ctx     = context.Background()
	columns = []contract.Column{
		{Name: "Month", Kind: contract.Name, Headers: []string{"", "Symbol", "Unit"}},

		{Name: "Heating days", Kind: contract.Count, Headers: []string{"n-2", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: contract.Energy, Headers: []string{"n-2", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: contract.Volume, Headers: []string{"n-2", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: contract.Temperature, Headers: []string{"n-2", "θkū", "°C"}},

		{Name: "Heating days", Kind: contract.Count, Headers: []string{"n-1", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: contract.Energy, Headers: []string{"n-1", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: contract.Volume, Headers: []string{"n-1", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: contract.Temperature, Headers: []string{"n-1", "θkū", "°C"}},

		{Name: "Heating days", Kind: contract.Count, Headers: []string{"n", "DApk", "Days"}},
		{Name: "Total heat energy consumption", Kind: contract.Energy, Headers: []string{"n", "QT", "MWh"}},
		{Name: "Domestic hot water consumption", Kind: contract.Volume, Headers: []string{"n", "V", "m³"}},
		{Name: "Domestic hot water temperature", Kind: contract.Temperature, Headers: []string{"n", "θkū", "°C"}},
	}

	rows = []contract.Row{ // using months like that in order to keep tabular outline
		{contract.Cell(time.Month(1).String()), "5", "10.1231", "45", "60.0", "11", "20", "105", "120", "20", "40", "200", "240"},
		{contract.Cell(time.Month(2).String()), "6", "11.3412", "44", "61.0", "10", "21", "30", "121", "21", "41", "201", "241"},
		{contract.Cell(time.Month(3).String()), "7", "12.0001", "43", "62.2", "12", "19", "102", "122", "22", "42", "202", "239"},
		{contract.Cell(time.Month(4).String()), "6", "13.0000", "42", "61.9", "15", "24", "101", "121", "21", "41", "201", "246"},
		{contract.Cell(time.Month(5).String()), "9", "14.1235", "386", "371.9", "455.1", "25", "102", "122", "22", "42", "202", "242"},
		{contract.Cell(time.Month(6).String()), "8", "15.0000", "40", "63.0", "13", "23", "103", "123", "23", "43", "203", "243"},
		{contract.Cell(time.Month(6).String()), "8", "15.0000", "40", "63.0", "13", "23", "103", "123", "23", "43", "203", "243"},
	}
)

type TOpts func(*testing.T, Store, *models.Document)

func NewTestUser(t *testing.T, s Store) *models.Document {
	t.Helper()
	return createUser(t, s, false, false, false)
}

func NewTestAdmin(t *testing.T, s Store) *models.Document {
	t.Helper()
	return createUser(t, s, true, false, false)
}

func NewTestPlatformManager(t *testing.T, s Store) *models.Document {
	t.Helper()
	return createUser(t, s, false, true, false)
}

func NewTestAdminNwManager(t *testing.T, s Store) *models.Document {
	t.Helper()
	return createUser(t, s, false, false, true)
}

func NewTestPortfolioRole(t *testing.T, s Store, role models.PortfolioRole, countries ...models.Country) *models.Document {
	t.Helper()
	u := createUser(t, s, false, false, false)

	if len(countries) == 0 {
		countries = []models.Country{models.CountryLatvia}
	}

	for _, c := range countries {
		if err := s.Portfolio().Put(context.Background(), u.ID, c, role); err != nil {
			t.Fatalf("Create %v in %v: %v", role, c, err)
		}
	}
	return u
}

func createUser(t *testing.T, s Store, isAdmin, isPfManager, isANM bool) *models.Document {
	t.Helper()
	userStore := s.FromKind("user")

	u := models.User{
		Name:            "John Doe",
		Email:           fmt.Sprintf("john_%s@example.com", uuid.New().String()),
		Password:        "foo",
		SuperUser:       isAdmin,
		PlatformManager: isPfManager,
		AdminNwManager:  isANM,
		Country:         models.CountryLatvia,
		IsActive:        true,
		Valid:           models.ValidationStatusValid,
	}

	d, err := userStore.Create(context.Background(), &u)

	if err != nil {
		t.Fatalf("create test user: %s", err)
	}
	return d
}

func NewTestOrg(t *testing.T, s Store, u ...uuid.UUID) *models.Document {
	t.Helper()

	o := models.Organization{
		Name:               "Goo Corporation",
		VAT:                "GB" + uuid.New().String(),
		Address:            "End of the world",
		Telephone:          "+359888123456",
		Website:            "https://goocorp.example",
		LegalForm:          models.LegalFormPublicOrganization,
		Registered:         time.Date(2017, time.May, 2, 12, 30, 10, 5, time.UTC),
		Email:              "notfakeorg@real.com",
		Valid:              models.ValidationStatusValid,
		Country:            models.CountryLatvia,
		RegistrationNumber: "1234567890",
	}

	var roles models.OrgRoles
	us := s.FromKind("user")
	orgS := s.FromKind("organization")

	if len(u) == 0 {
		roles = models.OrgRoles{LEAR: NewTestUser(t, us).ID}
		o.OrganizationRoles = append(o.OrganizationRoles, models.OrganizationRole{
			UserID:         NewTestUser(t, us).ID,
			OrganizationID: o.ID,
			Position:       "lear"})
	} else {
		roles = models.OrgRoles{
			LEAR:  u[0],
			LEAAs: u[1:],
		}
		o.OrganizationRoles = append(o.OrganizationRoles, models.OrganizationRole{
			UserID:         u[0],
			OrganizationID: o.ID,
			Position:       "lear"})
		for _, uid := range u[1:] {
			o.OrganizationRoles = append(o.OrganizationRoles, models.OrganizationRole{
				UserID:         uid,
				OrganizationID: o.ID,
				Position:       "leaa"})
		}
	}
	o.Roles = roles

	d, err := orgS.Create(context.Background(), &o)
	if err != nil {
		t.Fatalf("create test org: %s", err)
	}

	return d
}

func NewTestAsset(t *testing.T, s Store, opts ...TOpts) *models.Document {
	t.Helper()

	orgID := NewTestOrg(t, s.FromKind("organization")).ID
	as := s.FromKind("asset")
	category := models.NROfficeBuildings

	a := models.Asset{
		Owner:   orgID,
		ESCO:    &orgID,
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

	d, err := as.Create(context.Background(), &a)
	if err != nil {
		t.Fatalf("create test asset: %s", err)
	}

	for _, opt := range opts {
		opt(t, s, d)
	}

	d, err = as.Update(context.Background(), d)
	if err != nil {
		t.Fatalf("create test asset update opts: %s", err)
	}

	return d
}

func NewTestProject(t *testing.T, s Store, opts ...TOpts) *models.Document {
	t.Helper()
	var (
		prjS = s.FromKind("project")
		us   = s.FromKind("user")
		orgS = s.FromKind("organization")

		u     = NewTestUser(t, us)
		admin = NewTestAdmin(t, us)

		p = models.Project{
			Name:              uuid.New().String(),
			Owner:             NewTestOrg(t, orgS).ID,
			Asset:             NewTestAsset(t, s.FromKind("asset")).ID,
			Status:            models.ProjectStatusPlanning,
			AirTemperature:    20,
			WaterTemperature:  40,
			GuaranteedSavings: 51.16,
			Country:           models.CountryLatvia,
			PortfolioDirector: admin.ID,
			Milestone:         models.MilestoneAcquisitionMeeting,
		}

		d, err = prjS.Create(context.Background(), &p)
	)

	if err != nil {
		t.Fatalf("create test project: %s", err)
	}

	d.Data.(*models.Project).Roles.PM = make([]uuid.UUID, 1)
	d.Data.(*models.Project).Roles.PM[0] = u.ID

	pr := d.Data.(*models.Project).ProjectRoles
	pr = append(pr, models.ProjectRole{
		UserID:    u.ID,
		ProjectID: p.ID,
		Position:  "pm"})
	d.Data.(*models.Project).ProjectRoles = pr

	for _, opt := range opts {
		opt(t, prjS, d)
	}

	d, err = prjS.Update(context.Background(), d)
	if err != nil {
		t.Fatalf("create test project updating opts: %s", err)
	}
	return d
}

func NewTestContract(t *testing.T, s Store, p *models.Document) (doc, prj *models.Document) {
	t.Helper()
	var cs = s.FromKind("contract")

	if p == nil {
		prj = NewTestProject(t, cs)
	} else {
		prj = p
	}

	contr := contract.New(prj.ID)
	contr.Fields = map[string]string{
		"date-of-meeting":     "33-02-1990",
		"address-of-building": "dragan tzankov 33",
		"meeting-opened-by":   "mincho slivata",
		"chair-of-meeting":    "bache Kilo",
	}
	contr.Agreement = map[string]string{
		"assignor-address":           "Tintqva 14",
		"assignor-bank-account-iban": "9999-9999-9999-9999-9999",
	}

	doc, err := cs.Create(context.Background(), contr)

	if err != nil {
		t.Fatalf("Creating test contract failed: %s", err)
	}

	return
}

func NewTestInClima(t *testing.T, s Store) (ic *models.Document, p *models.Document) {
	t.Helper()
	var err error

	ctx := context.Background()

	ics := s.FromKind("indoorclima")
	ps := s.FromKind("project")
	cst := s.FromKind("contract")

	p = NewTestProject(t, ps)
	icm := contract.IndoorClima{
		Project: p.ID,
		Zones:   make(contract.JSONZone),
	}

	icm.AirexWindows.N = 67
	icm.AirexWindows.N1 = 67
	icm.AirexWindows.N2 = 67
	icm.HeatedVolumeBuilding = 7277

	z1 := contract.ZoneModel{
		Area:   111,
		UValue: 1.1,
	}

	icm.Zones["window_env1_zone1"] = z1

	ic, err = ics.Create(ctx, &icm)
	if err != nil {
		t.Fatalf("create test indoor clima: %s", err)
	}

	cdoc, err := cst.Create(ctx, contract.New(p.ID))
	if err != nil {
		t.Fatalf("creating contract fails with: %v", err)
	}

	table, _ := contract.NewTable(columns, rows...)
	cdoc.Data.(*contract.Contract).Tables["baseline"] = table
	cdoc.Data.(*contract.Contract).Tables["baseyear_n_2"] = table
	cdoc.Data.(*contract.Contract).Tables["baseyear_n_1"] = table
	cdoc.Data.(*contract.Contract).Tables["baseyear"] = table
	cst.Update(ctx, cdoc)

	return ic, p
}

type TNOpt func(*testing.T, *models.Notification)

func NewTestNotification(t *testing.T, ns Notifier, recp uuid.UUID, opts ...TNOpt) *models.Notification {
	t.Helper()
	n := models.Notification{
		Action:      models.UserActionUpload,
		RecipientID: recp,
		UserID:      uuid.New(),
		Old:         "hello",
		New:         "bye",
		Country:     models.CountryBulgaria,
	}

	for _, opt := range opts {
		opt(t, &n)
	}

	err := ns.Notify(context.Background(), &n)
	if err != nil {
		t.Fatalf("create test notification: %s", err)
	}
	not, err := ns.Get(context.Background(), n.ID, recp)
	if err != nil {
		t.Fatalf("could not get created notification: %s", err)
	}
	return not
}

func TNWithAction(action models.UserAction) TNOpt {
	return func(_ *testing.T, n *models.Notification) {
		n.Action = action
	}
}

func TNWithUser(u *models.User) TNOpt {
	return func(_ *testing.T, n *models.Notification) {
		n.UserID = u.ID
		n.UserKey = u.Key()
	}
}

func TNWithTarget(d *models.Document) TNOpt {
	return func(_ *testing.T, n *models.Notification) {
		n.TargetID = d.ID
		n.TargetKey = d.Data.Key()
	}
}

func TNWithOrg(o *models.Organization) TNOpt {
	return func(_ *testing.T, n *models.Notification) {
		n.TargetID = o.ID
		n.TargetKey = o.Name
		n.TargetType = models.OrganizationT
	}
}

func NewTestProjectCreationRequest(t *testing.T, store Store, opts ...TOpts) *models.ProjectCreationRequest {
	t.Helper()

	org := NewTestOrg(t, store)
	asset := NewTestAsset(t, store, TAWithOrg(org.ID))

	req := models.ProjectCreationRequest{
		Asset:        asset.ID,
		Organization: org.ID,
		User:         NewTestUser(t, store).ID,
		Status:       models.OpenedStatus,
	}

	for _, opt := range opts {
		opt(t, store, models.NewDocument(&req))
	}

	err := store.DB().Create(&req).Error
	if err != nil {
		t.Fatalf("creating request fails with: %v", err)
	}

	return &req
}

func NewTestGDPRRequest(t *testing.T, db *gorm.DB, aT models.GDPRType) *models.GDPRRequest {
	t.Helper()

	req := models.GDPRRequest{
		RequesterName:    "Ivan",
		RequesterPhone:   "0088112233",
		RequesterEmail:   "i.ivaonv@test.com",
		RequesterAddress: "Test address",
		Name:             "Petar",
		Phone:            "0088112234",
		Email:            "p.petrov@test.com",
		Address:          "test address peter",
		Action:           aT,
		Reason:           "Want to delete",
		Information:      "all my stuff.",
	}
	err := db.Create(&req).Error
	if err != nil {
		t.Fatalf("fail to create gdpr request: %v", err)
	}

	return &req
}

func NewTestWorkPhase(t *testing.T, s Store, pid uuid.UUID) *models.WorkPhase {
	t.Helper()

	wp := models.WorkPhase{
		Project: pid,
	}

	wp.Reviews = []models.WPReview{
		models.WPReview{
			Approved: true,
			Type:     models.WPReviewTypeExecutive,
			Comment:  "Lorem",
		},
		models.WPReview{
			Approved: false,
			Type:     models.WPReviewTypeFinancial,
			Comment:  "Lorem",
		},
		models.WPReview{
			Approved: false,
			Type:     models.WPReviewTypeBankAccount,
			Comment:  "Lorem",
		},
		models.WPReview{
			Approved: false,
			Type:     models.WPReviewTypeTechnical,
			Comment:  "Lorem",
		},
	}

	err := s.DB().Create(&wp).Error
	if err != nil {
		t.Fatalf("fail to create work phase: %v", err)
	}
	return &wp
}

func NewTestMonitoringPhase(t *testing.T, s Store, pid uuid.UUID) *models.MonitoringPhase {
	t.Helper()

	mp := models.MonitoringPhase{
		Project: pid,
	}

	// Number of reviews should be equal to years of
	// project. It is set to 30 because there should be
	// no project with more than that.
	mp.Reviews = make([]models.MPReview, 30)
	for i := 0; i < 30; i++ {
		mp.Reviews[i] = models.MPReview{
			Approved: true,
			Type:     models.MPReviewTypeForfaiting,
			Comment:  "Lorem",
		}
	}

	err := s.DB().Create(&mp).Error
	if err != nil {
		t.Fatalf("fail to create monitoring phase: %v", err)
	}
	return &mp
}

func TPrjReqWithUser(gu uuid.UUID) TOpts {
	return func(t *testing.T, _ Store, d *models.Document) {
		d.Data.(*models.ProjectCreationRequest).User = gu
	}
}

func TPrjReqWithAsset(a uuid.UUID) TOpts {
	return func(t *testing.T, _ Store, d *models.Document) {
		d.Data.(*models.ProjectCreationRequest).Asset = a
	}
}

func TPrjWithMilestone(m models.Milestone) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Project).Milestone = m
	}
}

// TPrjWithOrg updates Project with custom Organization.
func TPrjWithOrg(id uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Project).Owner = id
	}
}

// TPrjWithConsrOrg updates Project with custom consortium
// Organization.
func TPrjWithConsrOrg(id uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		o := []string{id.String()}
		d.Data.(*models.Project).ConsortiumOrgs = o
	}
}

// TPrjWithAsset updates Project with custom asset.
func TPrjWithAsset(id uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Project).Asset = id
	}
}

// TAWithESCO updates Project with custom ESCO Organization.
func TAWithESCO(id uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		o := id
		d.Data.(*models.Asset).ESCO = &o
	}
}

func TPrjWithCountry(c models.Country) TOpts {
	return func(_ *testing.T, _ Store, d *models.Document) {
		d.Data.(*models.Project).Country = c
	}
}

// TPrjWithPm modify the PM of an project.
func TPrjWithPm(id uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		tmp := d.Data.(*models.Project).Roles.PM
		tmp = append(tmp, id)
		d.Data.(*models.Project).Roles.PM = tmp

		pr := d.Data.(*models.Project).ProjectRoles
		pr = append(pr, models.ProjectRole{
			UserID:    id,
			ProjectID: d.ID,
			Position:  "pm"})
		d.Data.(*models.Project).ProjectRoles = pr
	}
}

// TPrjWithRole takes roles as map of role and slice of uuid.UUIDs. Append each
// slice of ids to corresponding project roles.
func TPrjWithRole(roles map[string][]uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		pr := d.Data.(*models.Project).ProjectRoles
		for r, ids := range roles {
			for _, id := range ids {
				pr = append(pr, models.ProjectRole{
					UserID:    id,
					ProjectID: d.ID,
					Position:  r})

			}
			d.Data.(*models.Project).ProjectRoles = pr
		}
	}
}

// TAWithOrg modify the owner of the asset.
func TAWithOrg(org uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Asset).Owner = org
	}
}

// TAWithEsco modifies the owner of the asset.
func TAWithEsco(org *uuid.UUID) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Asset).ESCO = org
	}
}

// TAWithAddr modifies the address of an asset.
func TAWithAddr(addr string) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Asset).Address = addr
	}
}

// TAWithBuildingType modify the building type of an asset.
func TAWithBuildingType(btype models.Building) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Asset).BuildingType = btype
	}
}

// TAWithStatus modify the validation status of an asset.
func TAWithStatus(vtype models.ValidationStatus) TOpts {
	return func(t *testing.T, st Store, d *models.Document) {
		d.Data.(*models.Asset).Valid = vtype
	}
}

func TAWithCountry(c models.Country) TOpts {
	return func(_ *testing.T, _ Store, d *models.Document) {
		d.Data.(*models.Asset).Country = c
	}
}

func NewTestMeeting(t *testing.T, s Store, opts ...func(*models.Meeting)) models.Meeting {
	org := NewTestOrg(t, s.FromKind("organization"))
	topic := models.MTypeOther

	m := models.Meeting{
		Name:        "init meeting",
		Host:        org.ID,
		Location:    "Tintyava 15-17",
		Date:        time.Now(),
		Objective:   "create eating fascility",
		Stakeholder: models.LegalFormNGO,
		Stage:       "init meeting",
		Notes:       "nice one",
		Topic:       &topic,
		Guests: []models.MeetingGuest{
			models.MeetingGuest{Name: "John", Email: "john@stageia.tech", Phone: "0888123123", Type: models.StakeHoldersTypeNGO},
			models.MeetingGuest{Name: "Michael", Email: "michael@stageia.tech", Phone: "0888123123", Type: models.StakeHoldersTypeNGO},
		},
	}

	for _, opt := range opts {
		opt(&m)
	}

	d, err := s.Create(context.Background(), &m)
	if err != nil {
		t.Fatalf("create meeting fails: %s", err)
	}

	result := d.Data.(*models.Meeting)

	return *result
}

func TMeetingWithOrg(org uuid.UUID) func(m *models.Meeting) {
	return func(m *models.Meeting) {
		m.Host = org
	}
}

func TMeetingWithPrj(prj uuid.UUID) func(m *models.Meeting) {
	return func(m *models.Meeting) {
		m.Project = &prj
	}
}

func NewTestAttachment(t *testing.T, st Store, ids ...uuid.UUID) *models.Attachment {
	t.Helper()

	var id = uuid.New()
	if len(ids) != 0 {
		id = ids[0]
	}

	att := models.Attachment{
		Name:  "new-attachment",
		Owner: id,
		Size:  int64(1024),
	}

	if err := st.DB().Create(&att).Error; err != nil {
		t.Fatalf("fails to create attachment: %s", err)
	}

	return &att
}

func NewTestFA(t *testing.T, st Store, opts ...TOpts) *models.ForfaitingApplication {
	t.Helper()

	st = st.FromKind("forfaiting_application")

	manager := NewTestUser(t, st)
	fa := models.ForfaitingApplication{
		Value:     models.Value{ID: uuid.New()},
		Project:   NewTestProject(t, st).ID,
		ManagerID: manager.ID,
		Manager:   *manager.Data.(*models.User),
		Finance:   models.FinanceEquity,
	}

	fa.Reviews = []models.FAReview{
		models.FAReview{
			Approved: true,
			Type:     models.FAReviewTypeExecutive,
			Comment:  "Lorem",
		},
		models.FAReview{
			Approved: false,
			Type:     models.FAReviewTypeFinancial,
			Comment:  "Lorem",
		},
		models.FAReview{
			Approved: false,
			Type:     models.FAReviewTypeGuidelines,
			Comment:  "Lorem",
		},
		models.FAReview{
			Approved: false,
			Type:     models.FAReviewTypeTechnical,
			Comment:  "Lorem",
		},
	}

	dd := models.NewDocument(&fa)
	for _, opt := range opts {
		opt(t, st, dd)
	}

	doc, err := st.Create(ctx, &fa)
	if err != nil {
		t.Fatalf("fail to create fa: %v", err)
	}

	ba := models.BankAccount{
		BeneficiaryName: "Jon",
		BankNameAddress: "eow",
		IBAN:            "111111111",
		FAID:            fa.ID,
	}
	_, err = st.Create(ctx, &ba)
	if err != nil {
		t.Errorf("fail to create bank acc: %v", err)
	}
	doc.Data.(*models.ForfaitingApplication).BankAccount = ba
	doc, err = st.Update(ctx, doc)
	if err != nil {
		t.Errorf("failed to update FA with bank acc: %v", err)
	}

	return doc.Data.(*models.ForfaitingApplication)
}

func NewTestFP(t *testing.T, st Store, opts ...TOpts) *models.ForfaitingPayment {
	t.Helper()

	st = st.FromKind("forfaiting_payment")

	fp := models.ForfaitingPayment{
		Value:         models.Value{ID: uuid.New()},
		Project:       NewTestProject(t, st).ID,
		Currency:      models.CurrencyEUR,
		TransferValue: 9001,
	}

	dd := models.NewDocument(&fp)
	for _, opt := range opts {
		opt(t, st, dd)
	}
	doc, err := st.Create(ctx, &fp)
	if err != nil {
		t.Fatalf("fail to create fp: %v", err)
	}

	return doc.Data.(*models.ForfaitingPayment)
}

func TFAWithProject(id uuid.UUID) TOpts {
	return func(_ *testing.T, _ Store, d *models.Document) {
		d.Data.(*models.ForfaitingApplication).Project = id
	}
}

func TFPWithProject(id uuid.UUID) TOpts {
	return func(_ *testing.T, _ Store, d *models.Document) {
		d.Data.(*models.ForfaitingPayment).Project = id
	}
}
