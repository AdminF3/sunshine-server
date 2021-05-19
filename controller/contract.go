package controller

import (
	"context"
	"fmt"
	"html"
	"io"
	"io/ioutil"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Contract struct {
	pst stores.Store
	cst stores.Store
	url string
}

func NewContract(env *services.Env) *Contract {
	return &Contract{
		pst: env.ProjectStore,
		cst: env.ContractStore,
		url: env.General.URL,
	}
}

type feature func(context.Context, *contractCTX, stores.Store, map[string]string) error

type contractCTX struct {
	id          uuid.UUID
	table       string
	project     *models.Project
	doc         *models.Document
	contract    *contract.Contract
	asset       *models.Asset
	client      *models.Organization
	esco        *models.Organization
	buildOwner  *models.Organization
	ESCoLear    *models.User
	ic          *models.Document
	attachments map[string]models.Attachment
	baccount    *models.BankAccount
	eurobor     decimal.Decimal
	vat         decimal.Decimal
}

func (c *Contract) GetTable(ctx context.Context, id uuid.UUID, vars map[string]string) (*contract.Table, error) {
	ctr, err := c.buildContext(ctx, id, vars, withTable)
	if err != nil {
		return nil, err
	}

	ids := []uuid.UUID{id}
	for _, id := range ctr.project.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}
	if !canGetProject(ctx, GetProjectContractTable, ctr.project.Country, ids...) {
		return nil, ErrUnauthorized
	}

	table, ok := ctr.contract.Tables[ctr.table]
	if !ok {
		return nil, fmt.Errorf("no such table: %w", ErrBadInput)
	}

	return &table, nil
}

func (c *Contract) DownloadContract(ctx context.Context, id uuid.UUID, language, format string) (*contract.FileInTempDir, string, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, "", err
	}

	if !Can(ctx, DownloadProjectContract, id, ctr.project.Country) {
		return nil, "", ErrUnauthorized

	}

	var (
		prjcountry = ctr.project.Country.LegalCountry()
		texFile    string
		op         func(context.Context, string) (*contract.FileInTempDir, error)
	)

	switch language {
	case "native":
		texFile = "contract.tex"
	case "english":
		texFile = "en_contract.tex"

		// This is needed because of the adapted english
		// versions contracts. If new adapted contract is
		// added, it should be added here as well, if not it
		// will be build the `base` version. Eventually when
		// every contract has adapted english version,
		// `texFile` should be equal every time to
		// `contract.tex` and this if will become
		// redundant. Also base version of the contract will
		// be still needed because of sanity check.
		if prjcountry == models.CountryBulgaria.String() ||
			prjcountry == models.CountryRomania.String() {
			texFile = "contract.tex"
			prjcountry = prjcountry + "_adp"
		}
	default:
		return nil, "", fmt.Errorf("bad language: %v", language)
	}

	doc, err := contract.NewDocumentFromLanguage(newTemplateContext(ctr, c.url), prjcountry)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate contract template: %w", err)
	}

	switch format {
	case "pdf":
		op = doc.GeneratePDF
	case "tex":
		op = doc.GenerateTeX
	default:
		return nil, "", fmt.Errorf("bad format: %v", format)
	}

	file, err := op(ctx, texFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate document for contract: %w", err)
	}

	return file, fmt.Sprintf("contract_%s_%s.%s", texFile, ctr.id, format), nil
}

func (c *Contract) DownloadAgreement(ctx context.Context, id uuid.UUID, format, language string) (*contract.FileInTempDir, string, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, "", err
	}

	if !Can(ctx, DownloadProjectAgreement, id, ctr.project.Country) {
		return nil, "", ErrUnauthorized
	}

	doc, err := contract.NewDocument(newTemplateContext(ctr, c.url))
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate contract template: %w", err)
	}

	var op func(context.Context, string) (*contract.FileInTempDir, error)
	switch format {
	case "pdf":
		op = doc.GeneratePDF
	case "tex":
		op = doc.GenerateTeX
	default:
		return nil, "", fmt.Errorf("bad format: %v", format)
	}

	var lng string
	switch language {
	case "native":
		lng = "agreement.tex"
	case "english":
		lng = "en_agreement.tex"
	default:
		return nil, "", fmt.Errorf("bad language: %v", language)
	}

	file, err := op(ctx, lng)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate document for agreement: %w", err)
	}

	return file, fmt.Sprintf("agreement_%s.%s", ctr.id, format), nil
}

func (c *Contract) UpdateTable(ctx context.Context, id uuid.UUID, table contract.Table, vars map[string]string) (*contract.Table, error) {
	ctr, err := c.buildContext(ctx, id, vars, withTable)
	if err != nil {
		return nil, err
	}

	if !Can(ctx, UpdateProjectContractTable, id, ctr.project.Country) {
		return nil, ErrUnauthorized
	}
	// Make sure that all rows are valid and no columns are edited.
	newTable := ctr.contract.Tables[ctr.table].Empty()
	newTable.SetTitle(table.Title())

	for i := 0; i < table.Len(); i++ {
		row := table.Row(i)
		if err := newTable.AddRow(row); err != nil {
			return nil, fmt.Errorf("%w: row %d: %v", ErrBadInput, i, err)
		}
	}
	ctr.contract.Tables[ctr.table] = *newTable

	ctr.doc, err = c.cst.Update(ctx, ctr.doc)
	if err != nil {
		return nil, err
	}

	res := ctr.contract.Tables[ctr.table]
	return &res, nil
}

func (c *Contract) UpdateFields(ctx context.Context, id uuid.UUID, fields contract.JSONMap) (*models.Document, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	ctr.contract.Fields = fields

	if !Can(ctx, UpdateProjectContractFields, id, ctr.project.Country) {
		return nil, ErrUnauthorized
	}

	return c.cst.Update(ctx, ctr.doc)
}

func (c *Contract) GetFields(ctx context.Context, id uuid.UUID) (contract.JSONMap, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	ids := []uuid.UUID{id}
	for _, id := range ctr.project.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}
	if !canGetProject(ctx, GetProjectContractFields, ctr.project.Country, ids...) {
		return nil, ErrUnauthorized
	}

	return ctr.contract.Fields, nil
}

func (c *Contract) GetAgreement(ctx context.Context, id uuid.UUID) (contract.JSONMap, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	ids := []uuid.UUID{id}
	for _, id := range ctr.project.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}
	if !canGetProject(ctx, GetProjectAgreement, ctr.project.Country, ids...) {
		return nil, ErrUnauthorized
	}

	return ctr.contract.Agreement, nil
}

func (c *Contract) UpdateAgreement(ctx context.Context, id uuid.UUID, agreement contract.JSONMap) (*models.Document, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	if !Can(ctx, UpdateProjectAgreement, id, ctr.project.Country) {
		return nil, ErrUnauthorized
	}

	ctr.contract.Agreement = agreement

	return c.cst.Update(ctx, ctr.doc)
}

func (c *Contract) GetMaintenance(ctx context.Context, id uuid.UUID) (contract.JSONMap, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	ids := []uuid.UUID{id}
	for _, id := range ctr.project.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}
	if !canGetProject(ctx, GetProjectMaintenance, ctr.project.Country, ids...) {
		return nil, ErrUnauthorized
	}

	return ctr.contract.Maintenance, nil
}

func (c *Contract) UpdateMaintenance(ctx context.Context, id uuid.UUID, maintenance contract.JSONMap) (*models.Document, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	if !Can(ctx, UpdateProjectMaintenance, id, ctr.project.Country) {
		return nil, ErrUnauthorized
	}

	ctr.contract.Maintenance = maintenance

	return c.cst.Update(ctx, ctr.doc)
}

func (c *Contract) UpdateMarkdown(ctx context.Context, id uuid.UUID, r io.Reader) ([]byte, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	if !Can(ctx, UpdateProjectContractFields, id, ctr.project.Country) {
		return nil, ErrUnauthorized
	}

	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	ctr.contract.Markdown = []byte(html.EscapeString(string(content)))
	if _, err := c.cst.Update(ctx, ctr.doc); err != nil {
		return nil, fmt.Errorf("failed to update contract: %w", err)
	}

	return ctr.contract.Markdown, nil
}

func (c *Contract) GetMarkdown(ctx context.Context, id uuid.UUID) ([]byte, error) {
	ctr, err := c.buildContext(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	ids := []uuid.UUID{id}
	for _, id := range ctr.project.ConsortiumOrgs {
		ids = append(ids, uuid.MustParse(id))
	}
	if !canGetProject(ctx, GetProjectContractFields, ctr.project.Country, ids...) {
		return nil, ErrUnauthorized
	}

	return ctr.contract.Markdown, nil
}

func (c *Contract) buildContext(ctx context.Context, id uuid.UUID, vars map[string]string, funcs ...feature) (contractCTX, error) {
	var ok bool

	ctr := contractCTX{
		id:      id,
		eurobor: contract.GetEurobor(c.cst.DB()),
	}

	project, deps, err := c.pst.Unwrap(ctx, ctr.id)
	if err != nil {
		return ctr, err
	}

	ctr.project, ok = project.Data.(*models.Project)
	if !ok {
		return ctr, fmt.Errorf("not a project: %w", ErrFatal)
	}

	if a, ok := deps[ctr.project.Asset]; ok {
		ctr.asset = a.Data.(*models.Asset)
	}

	gv := contract.GetVat(c.cst.DB(), ctr.project.Country)
	ctr.vat = decimal.NewFromFloat(gv)

	ctr.doc, err = c.cst.GetByIndex(ctx, ctr.id.String())
	if err != nil {
		ctr.doc, err = c.cst.Create(ctx, contract.New(ctr.id))
		if err != nil {
			return ctr, ErrFatal
		}
	}

	ctr.contract, ok = ctr.doc.Data.(*contract.Contract)
	if !ok {
		return ctr, fmt.Errorf("not a contract: %w", ErrFatal)
	}

	for _, f := range funcs {
		if err := f(ctx, &ctr, c.cst, vars); err != nil {
			return ctr, err
		}
	}

	//add attachments
	mdocs, _, _, err := c.cst.FromKind("meeting").
		ListByMember(ctx, stores.Filter{}, ctr.project.ID)
	if err != nil {
		return ctr, err
	}

	ctr.attachments = project.Attachments

	for _, d := range mdocs {
		for k, v := range d.Attachments {
			ctr.attachments[k] = v
		}
	}

	o, err := c.cst.FromKind("organization").Get(ctx, ctr.project.Owner)
	if err != nil {
		return ctr, err
	}

	ctr.esco = o.Data.(*models.Organization)

	lear, err := c.cst.FromKind("user").Get(ctx, ctr.esco.Roles.LEAR)
	if err != nil {
		return ctr, err
	}

	ctr.ESCoLear = lear.Data.(*models.User)

	bo, err := c.cst.FromKind("organization").Get(ctx, ctr.asset.Owner)
	if err != nil {
		return ctr, err
	}

	ctr.buildOwner = bo.Data.(*models.Organization)

	var ba models.BankAccount
	err = c.cst.DB().
		Joins("INNER JOIN forfaiting_applications as fa ON fa.id = bank_accounts.fa_id").
		Joins("INNER JOIN projects as p ON p.id = fa.project_id AND p.id = ?", project.ID).
		First(&ba).Error
	if err != nil && !stores.IsRecordNotFound(err) {
		return ctr, err
	}
	ctr.baccount = &ba

	return ctr, nil
}

// calcContract is a dummy save of contract, so the recalculation can
// be triggered.
func calcContract(db *gorm.DB, proj uuid.UUID) error {
	ctr := &contract.Contract{}
	err := db.Where("project_id = ?", proj).First(ctr).Error
	if err != nil {
		return fmt.Errorf("fail to recalculate contract: %w", err)
	}

	if err = db.Save(ctr).Error; err != nil {
		return fmt.Errorf("fail to recalculate contract: %w", err)
	}
	return nil
}

func withTable(ctx context.Context, ctr *contractCTX, _ stores.Store, vars map[string]string) error {
	var (
		ok  bool
		err error
	)
	ctr.table, ok = vars["table"]
	if !ok {
		err = ErrInvalidTable
	}
	return err
}

func newTemplateContext(ctx contractCTX, url string) contract.TemplateContext {
	ctr := contract.TemplateContext{}
	ctr.Attachments = make(map[string]string)
	if ctx.contract != nil {
		ctr.Contract = *ctx.contract
	}

	if ctx.project != nil {
		ctr.Project = *ctx.project
		ctr.AssetSnapshot = ctx.project.AssetSnapshot
	}

	if ctx.asset != nil {
		ctr.Asset = *ctx.asset
	}

	if ctx.client != nil {
		ctr.Client = *ctx.client
	}

	if ctx.esco != nil {
		ctr.ESCo = *ctx.esco
	}

	if ctx.ESCoLear != nil {
		ctr.LEAR = *ctx.ESCoLear
	}

	if ctx.buildOwner != nil {
		ctr.BuildingOwner = *ctx.buildOwner
	}

	u := fmt.Sprintf("%s/project/%s/", url, ctr.Project.Value.ID.String())
	for k, v := range ctx.attachments {
		switch v.UploadType {
		case "aquisition protocol meeting":
			ctr.Attachments["acquisition_meeting"] = u + k
		case "commitment protocol meeting":
			ctr.Attachments["commitment_meeting"] = u + k
		case "kickoff protocol meeting":
			ctr.Attachments["kickoff_meeting"] = u + k
		case "energy audit report":
			ctr.Attachments["energy_audit_report"] = u + k
		case "technical inspection report":
			ctr.Attachments["technical_inspection_report"] = u + k
		default:
			u = fmt.Sprintf("%s/meeting/%v/", url, v.Owner)
			ctr.Attachments["others"] = u + k
		}
	}

	if ctx.baccount != nil {
		ctr.FABankAcc = *ctx.baccount
	}
	ctr.EUROBOR, _ = ctx.eurobor.Float64()
	ctr.VAT, _ = ctx.vat.Float64()

	return ctr
}
