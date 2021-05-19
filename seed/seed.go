package seed

import (
	"context"
	"math/rand"
	"time"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"syreclabs.com/go/faker"
)

// Users holds type : user : pass
type Users map[string]map[*models.User]string

type Seeder struct {
	env *services.Env
}

func NewSeeder(env *services.Env) *Seeder {
	return &Seeder{
		env: env,
	}
}

func (s Seeder) Users(ctx context.Context, ucount int) (Users, error) {
	us := make(map[*models.User]string)
	res := make(Users)

	admin, pass, err := user(ctx, s.env.UserStore, true)
	if err != nil {
		return nil, err
	}
	us[admin] = pass
	res["admin"] = us

	us = make(map[*models.User]string)
	for i := 0; i < ucount; i++ {
		// create list of users
		u, p, err := user(ctx, s.env.UserStore, false)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			// create the first user as PD and exclude from the main user list
			err := s.env.UserStore.Portfolio().Put(ctx, u.ID, models.CountryBulgaria, models.PortfolioDirectorRole)
			if err != nil {
				return nil, err
			}
			res["pd"] = map[*models.User]string{u: p}
			continue
		}

		us[u] = p
	}

	res["user"] = us

	return res, nil
}

func (s Seeder) Organizations(ctx context.Context, users Users, ocount int) ([]*models.Organization, error) {
	var (
		usrs = make([]*models.User, 0)
		res  = make([]*models.Organization, ocount)
	)

	for k := range users["user"] {
		usrs = append(usrs, k)
	}

	for i := 0; i < ocount; i++ {
		org, err := org(ctx, s.env.OrganizationStore, usrs[i%len(usrs)].ID)
		if err != nil {
			return nil, err
		}

		res[i] = org
	}

	return res, nil
}

func (s Seeder) Assets(ctx context.Context, orgs []*models.Organization, count int) ([]*models.Asset, error) {
	res := make([]*models.Asset, count)

	for i := 0; i < count; i++ {
		asset, err := asset(ctx, s.env.AssetStore, orgs[i%len(orgs)].ID)
		if err != nil {
			return nil, err
		}
		res[i] = asset
	}

	return res, nil
}

func (s Seeder) Projects(ctx context.Context, assets []*models.Asset, u map[*models.User]string, count int) (int, error) {
	var admin *models.User
	for k := range u {
		admin = k
	}

	index := count
	if len(assets) < count {
		index = len(assets)
	}

	for i := 0; i < index; i++ {
		err := project(ctx, s.env.ProjectStore, assets[i].Owner, assets[i].ID, admin.ID)
		if err != nil {
			return i, err
		}
	}

	return index, nil
}

func project(ctx context.Context, st stores.Store, org, asset, admin uuid.UUID) error {
	comp := faker.Company()

	format := "2006-01-02 15:04:05"

	from, _ := time.Parse(format, "2007-09-18 11:58:06")
	to, _ := time.Parse(format, "2012-05-18 11:58:06")

	p := models.Project{
		Name:              comp.Name(),
		Owner:             org,
		Asset:             asset,
		Status:            models.ProjectStatusPlanning,
		AirTemperature:    20,
		WaterTemperature:  40,
		GuaranteedSavings: 51.16,
		Country:           "Latvia",
		PortfolioDirector: admin,
		Milestone:         models.MilestoneAcquisitionMeeting,
		ConstructionFrom:  from,
		ConstructionTo:    to,
	}

	prj, err := st.Create(ctx, &p)
	if err != nil {
		return err
	}

	if _, err := st.FromKind("contract").Create(ctx, contract.New(prj.ID)); err != nil {
		return err
	}

	um, err := st.FromKind("user").Get(ctx, admin)
	if err != nil {
		return err
	}

	u := um.Data.(*models.User)

	fa := models.ForfaitingApplication{
		Project: prj.ID,
		Manager: *u,
	}

	if err := st.DB().Create(&fa).Error; err != nil {
		return err
	}

	bank := models.BankAccount{
		FAID:            fa.ID,
		BeneficiaryName: faker.Name().String(),
		BankNameAddress: faker.Address().String(),
		IBAN:            faker.RandomString(18),
		SWIFT:           faker.RandomString(8),
	}

	if err := st.DB().Create(&bank).Error; err != nil {
		return err
	}

	return err
}

func asset(ctx context.Context, st stores.Store, orgID uuid.UUID) (*models.Asset, error) {
	addr := faker.Address()
	num := faker.Number()
	comp := faker.Company()

	a := models.Asset{
		Owner:   orgID,
		Address: addr.String(),
		Coordinates: models.Coords{
			Lat: rand.Float32(),
			Lng: rand.Float32(),
		},
		Area:         num.NumberInt(4),
		HeatedArea:   num.NumberInt(3),
		BillingArea:  num.NumberInt(4),
		Flats:        num.NumberInt(2),
		Floors:       num.NumberInt(2),
		StairCases:   num.NumberInt(2),
		BuildingType: models.BuildingType318,
		HeatingType:  models.HeatingDistrict,
		Cadastre:     comp.Ein(),
		Valid:        models.ValidationStatusValid,
		Country:      "Latvia",
	}

	d, err := st.Create(ctx, &a)
	if err != nil {
		return nil, err
	}

	return d.Data.(*models.Asset), nil
}

func org(ctx context.Context, st stores.Store, uid uuid.UUID) (*models.Organization, error) {
	comp := faker.Company()
	addr := faker.Address()
	ph := faker.PhoneNumber()
	www := faker.Internet()
	date := faker.Date()

	o := models.Organization{
		Name:       comp.Name(),
		VAT:        "GB" + comp.Ein(),
		Address:    addr.String(),
		Telephone:  ph.CellPhone(),
		Website:    www.Url(),
		LegalForm:  models.LegalFormPublicOrganization,
		Registered: date.Backward(200 * time.Hour),
		Email:      www.Email(),
		Valid:      models.ValidationStatusValid,
		Country:    "Latvia",
	}

	o.Roles = models.OrgRoles{LEAR: uid}
	o.OrganizationRoles = []models.OrganizationRole{
		{
			UserID:         uid,
			OrganizationID: o.ID,
			Position:       "lear",
		},
	}

	d, err := st.Create(ctx, &o)
	if err != nil {
		return nil, err
	}

	return d.Data.(*models.Organization), nil
}

func user(ctx context.Context, s stores.Store, isAdmin bool) (*models.User, string, error) {
	name := faker.Name().Name()
	i := faker.Internet()
	pass := "112233p"

	u := models.User{
		Name:      name,
		Email:     i.Email(),
		Password:  pass,
		SuperUser: isAdmin,
		Country:   "Latvia",
		IsActive:  true,
		Valid:     models.ValidationStatusValid,
	}

	d, err := s.Create(ctx, &u)
	if err != nil {
		return nil, "", err
	}

	return d.Data.(*models.User), pass, nil
}
