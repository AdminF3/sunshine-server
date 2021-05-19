package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/shopspring/decimal"
)

// supportedCountries declares projects in which country this module is
// capable of building PDF contract for.
var supportedCountries = []models.Country{models.CountryAustria, models.CountryBulgaria, models.CountryLatvia, models.CountryPoland, models.CountryRomania,
	models.CountryAlbania, models.CountryAndorra, models.CountryArmenia, models.CountryBelarus, models.CountryBelgium, models.CountryBosniaHerzegovina,
	models.CountryCroatia, models.CountryCyprus, models.CountryCzech, models.CountryDenmark, models.CountryEstonia, models.CountryFinland, models.CountryFrance,
	models.CountryGeorgia, models.CountryGermany, models.CountryGreece, models.CountryHungary, models.CountryIreland, models.CountryItaly, models.CountryKosovo,
	models.CountryLiechtenstein, models.CountryLithuania, models.CountryLuxembourg, models.CountryMalta, models.CountryMoldova, models.CountryMonaco,
	models.CountryMontenegro, models.CountryNetherlands, models.CountryMacedonia, models.CountryNorway, models.CountryPortugal, models.CountryRussia,
	models.CountrySanMarino, models.CountrySerbia, models.CountrySlovenia, models.CountrySpain, models.CountrySweden, models.CountrySwitzerland,
	models.CountryTurkey, models.CountryUkraine, models.CountryUK, models.CountryVatican, models.CountrySlovakia}

// baseTemplate is a templated with parsed preamble and english contract and
// should be used as a base template for each contract template.
//
// NOTE: This value should never be used directly. Each use outside of the init
// function shall be invoking the Clone() method and use the result in order
// not to mutate it and create a race condition.
var baseTemplate *template.Template

// TODO: This global state is stupid. Fix it.
var paths config.Paths

// init tries to generate a country specific pdf template for all our
// supported countries. Panics if such template cannot be created.
func init() {
	paths = config.Load().Paths
	// Parse preamble and english contract and use them as a base for pdf
	// value of each country.
	path, err := contractPath("base")
	if err != nil {
		panic(fmt.Sprintf("base contract path: %v", err))
	}
	baseTemplate = template.Must(template.New("").Funcs(fmap).
		ParseGlob(filepath.Join(path, "*.tex")))

	// Make sure we can create pdf value for each supported country.
	for _, c := range supportedCountries {
		if c.IsConsortium() {
			// TODO this is hack because initially we do
			// not have all adapted versions; after adding
			// all, refactor this section and simplify it
			// a bit
			lang := strings.ToLower(string(c)) + "_adp"

			tc := TemplateContext{Project: models.Project{Country: c}}
			if _, err := NewDocument(tc); err != nil {
				panic(fmt.Sprintf("Can not generate contracts for %s: %v", c, err))
			}

			if c.HasAdaptedContract() {
				// TODO add adapted versions here
				if _, err := NewDocumentFromLanguage(tc, lang); err != nil {
					panic(fmt.Sprintf("Can not generate contracts for %s: %v", lang, err))
				}
			}
		} else {
			// Generate the latvian contract for all non-consortium countries.
			tc := TemplateContext{Project: models.Project{Country: c}}
			if _, err := NewDocumentFromLanguage(tc, "latvia"); err != nil {
				panic(fmt.Sprintf("Can not generate contracts for %s: %v", c, err))
			}
		}
	}
}

type TemplateContext struct {
	Context       context.Context
	Asset         models.Asset
	AssetSnapshot models.AssetSnapshot
	Contract      Contract
	Project       models.Project
	Client        models.Organization
	ESCo          models.Organization
	BuildingOwner models.Organization
	LEAR          models.User
	Markdown      io.Reader
	Attachments   map[string]string
	FABankAcc     models.BankAccount
	EUROBOR       float64
	VAT           float64
}

type document struct {
	tex *template.Template
	ctx TemplateContext
}

// NewDocument creates a country specific pdf template determined
// by the Project's country in the TemplateContext.
func NewDocument(ctx TemplateContext) (*document, error) {
	cp, err := contractPath(ctx.Project.Country.String())
	if err != nil {
		return nil, err
	}
	return newDocument(ctx, cp)
}

// NewDocumentFromLanguage create a new tex template for the giving
// language.  Usually the language should be one of the
// `supportedCountries` (e.g. `bulgaria`) or adapted version for that
// country (e.g. `<country>_adp`).
func NewDocumentFromLanguage(ctx TemplateContext, language string) (*document, error) {
	cp, err := contractPath(language)
	if err != nil {
		return nil, err
	}
	return newDocument(ctx, cp)
}

// NewDocumentFromPath creates a document template from a given path, ignoring
// the country set by the context.
func NewDocumentFromPath(ctx TemplateContext, path string) (*document, error) {
	return newDocument(ctx, path)
}

func newDocument(ctx TemplateContext, path string) (*document, error) {
	// template.Must is safe to be called over Clone.
	// text/template.Template.Clone could never return a non-nil error.
	// That result is put there just so the Template type in both
	// html/template and text/template can have the same API.
	templ, err := template.Must(baseTemplate.Clone()).
		ParseGlob(filepath.Join(path, "*.tex"))
	return &document{
		tex: templ,
		ctx: ctx,
	}, err
}

// GeneratePDF produces PDF file of a project contract after performing all the calculations.
func (p *document) GeneratePDF(ctx context.Context, rootFile string) (*FileInTempDir, error) {
	// genPDF runs xetex command from given dir and stdin from given
	// io.Reader with jobname "out" and blocks until it finishes. On
	// successful run the output file is located in 'dir/out.pdf'.
	//
	// ctx is used to kill the process if the context become done before
	// the command completes on its own.
	genPDF := texOp{
		do: func(ctx context.Context, r io.Reader, dir string) error {
			cmd := exec.CommandContext(ctx, "xelatex",
				"-halt-on-error", "-jobname=out")
			cmd.Dir = dir
			cmd.Stdin = r
			// Uncomment this while debugging TeX files.
			// cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
		output: "out.pdf",
	}
	return p.generate(ctx, genPDF, rootFile)
}

// GenerateTeX dumps raw LaTex file of project's contract after performing all the calculations.
func (p *document) GenerateTeX(ctx context.Context, rootFile string) (*FileInTempDir, error) {
	// dumpTeX simply creates a file and copy's whatever is read from r to
	// dir/out.tex.
	dumpTeX := texOp{
		do: func(ctx context.Context, r io.Reader, dir string) error {
			f, err := os.Create(filepath.Join(dir, "out.tex"))
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, r)
			return err
		},
		output: "out.tex",
	}
	return p.generate(ctx, dumpTeX, rootFile)
}

// texOp is an operation on a LaTeX file which produces an output in
// pre-defined place and format.
type texOp struct {
	do     func(ctx context.Context, r io.Reader, dir string) error
	output string
}

func (p *document) generate(ctx context.Context, op texOp, root string) (*FileInTempDir, error) {
	dir, err := ioutil.TempDir("", "sunshine_latex")
	if err != nil {
		return nil, err
	}

	if err = os.Symlink(imagesPath(), path.Join(dir, "images")); err != nil {
		return nil, err
	}

	// errch is channel for errors occurred in the pipeline below. Each
	// step in the flow is being ran in goroutine and is supposed to be
	// piped to the next. Speaking in UNIX shell terms consider the
	// following blocks below as something roughly like:
	//
	//     md2tex | executeTemplate | xelatex
	//
	// If any step fails PDF cannot be generated and this whole function
	// should fail. Thus all but the last steps should write ONLY non-nil
	// errors to errch. The last one writes whatever error it gets and then
	// we receive from that channel EXACTLY ONCE. In other words we only
	// care about the first send value as the real error should be there.
	// All the rest are to be due to incomplete input from the previous
	// (already failed).
	//
	// errch SHOULD be buffered with size N-1 where N is the count of steps
	// that could possibly fail. Otherwise on non-nil error anywhere in the
	// pipeline we don't drain the channel and if not buffered all but the
	// first goroutines in the flow will be blocked forever.
	errch := make(chan error, 2)
	mdr, mdw := io.Pipe()
	p.ctx.Markdown = mdr
	go func() {
		defer mdw.Close()
		if err := md2tex(ctx, mdw, p.ctx.Contract.Markdown); err != nil {
			errch <- fmt.Errorf("pandoc: %w", err)
		}
	}()

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		if err := p.tex.ExecuteTemplate(w, root, p.ctx); err != nil {
			errch <- fmt.Errorf("template: %w", err)
		}
	}()

	go func() {
		defer r.Close()
		err := op.do(ctx, r, dir)
		if err != nil {
			err = fmt.Errorf("latex: %w", err)
		}
		errch <- err
	}()

	if err := <-errch; err != nil {
		os.RemoveAll(dir)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			log.Printf("%s.PDF of %q failed: %q", root, p.ctx.Contract.Project, err)
			return nil, fmt.Errorf("generate pdf: %w", err)
		}
	}

	return OpenFileInTempDir(filepath.Join(dir, op.output))
}

// md2tex writes latex to w converted from Markdown read from r using pandoc.
//
// ctx is used to kill the process if the context become done before the
// command completes on its own.
func md2tex(ctx context.Context, w io.Writer, input []byte) error {
	cmd := exec.CommandContext(ctx, "pandoc", "--read=markdown", "--write=latex")
	cmd.Stdin = bytes.NewBuffer(input)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var fmap = template.FuncMap{
	"column": func(columns []Column) string {
		var names = make([]string, len(columns))
		for i, c := range columns {
			names[i] = c.Name
		}
		return join(names)
	},
	"column_sideways": func(columns []Column) string {
		var names = make([]string, len(columns))
		for i, c := range columns {
			names[i] = `\begin{sideways} ` + c.Name + ` \end{sideways}`
		}
		return join(names)
	},
	"row": func(row Row) string {
		var cells = make([]string, len(row))
		for i, cell := range row {
			cells[i] = cell.String()
		}
		return join(cells)
	},
	"rowf": func(table Table, row Row) string {
		var cells = make([]string, len(row))
		for i, cell := range row {
			if table.Column(i).Kind.IsNumber() {
				f, _ := cell.Decimal().Float64()
				cells[i] = FormatFloat(f).String()
			} else {
				cells[i] = cell.String()
			}
		}
		return join(cells)
	},
	"rowfmt": func(table Table, row Row) string {
		var cells = make([]string, len(row))
		for i, cell := range row {
			cells[i] = cell.Format(table.Column(i).Kind)
		}
		return join(cells)
	},
	"average": func(t Table) string {
		var columns = make([]string, t.ColumnLen())
		for i, c := range t.columns {
			if c.Kind == Name {
				columns[i] = "Average"
			} else if CanAvg(c) {
				if total, err := t.Average(i); err == nil {
					columns[i] = total.String()
				}
			} else {
				columns[i] = "{-}"
			}
		}
		return join(columns)
	},
	"total": func(t Table) string {
		var columns = make([]string, t.ColumnLen())
		for i, c := range t.columns {
			if c.Kind == Name {
				columns[i] = "Total"
			} else if CanSum(c) {
				if total, err := t.Total(i); err == nil {
					columns[i] = total.String()
				}
			} else {
				columns[i] = "{-}"
			}
		}
		return join(columns)
	},
	"total_excl_vat": func(t Table) string {
		return totalFormat(t, `\textbf{Total \emph{(ex VAT)}}`,
			func(d decimal.Decimal) decimal.Decimal {
				return d
			})
	},
	"vat": func(t Table, vat float64) string {
		return totalFormat(t, `\textbf{VAT}`,
			func(d decimal.Decimal) decimal.Decimal {
				return d.Mul(decimal.NewFromFloat(vat))
			})
	},
	"total_incl_vat": func(t Table, vat float64) string {
		return totalFormat(t, `\textbf{Total \emph{including VAT}}`,
			func(d decimal.Decimal) decimal.Decimal {
				return d.Add(d.Mul(decimal.NewFromFloat(vat)))
			})
	},
	"join_tables": func(tables ...Table) Table {
		return Join(tables...)
	},
	"baseyear": func(lang string, n int) string {
		t := unmarshalLang(lang, `{"en": "Year", "pl": "Rok", "ro": "An", "au": "Jahr", "lv":"Gadā", "bg": "Година"}`)

		return fmt.Sprintf("%s %d", t, n+1)
	},
	"date": func(t time.Time) string {
		return fmt.Sprintf("%d/%d/%d", t.Year(), t.Month(), t.Day())
	},
	"div": func(a, b int) int {
		return a / b
	},
	"mul": func(a, b int) int {
		return a * b
	},
	"read": func(r io.Reader) (string, error) {
		b, err := ioutil.ReadAll(r)
		return string(b), err
	},
	"date_diff": func(from, to time.Time) int {
		diff := to.Sub(from)
		return int(diff.Hours() / 24)
	},
	"translate": func(lang string, t Table) Table {
		return translate(lang, t)
	},
	"asset_address": func(addr string) string {
		var aj map[string]string
		err := json.Unmarshal([]byte(addr), &aj)
		if err != nil {
			return ""
		}

		return fmt.Sprintf(
			"\"%s %s %s %s\"",
			aj["postcode"],
			aj["streetAddress"],
			aj["city"],
			aj["country"],
		)
	},
}

// totalFormat returns a summed row with given function to modify the result
// (e.g. vat, total incl. vat, etc.)
func totalFormat(t Table, title string, f func(decimal.Decimal) decimal.Decimal) string {
	var columns = make([]string, t.ColumnLen())
	for i, c := range t.columns {
		if c.Kind == Name {
			columns[i] = title
		} else if CanSum(c) {
			if total, err := t.Total(i); err == nil {
				columns[i] = Cell(f(total).Round(2).String()).Format(c.Kind)
			}

		}
	}
	return join(columns)
}

// join all cells with LaTeX table separator.
func join(cells []string) string {
	return strings.Join(cells, " & ")
}

func contractPath(country string) (string, error) {
	fp := filepath.Join(paths.LaTeX, strings.ToLower(country))
	_, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("dir(%q) for %s country does not exist", fp, country)
	}
	return fp, err
}

func imagesPath() string { return filepath.Join(paths.LaTeX, "images") }

func translate(lang string, t Table) Table {
	// columns
	cols := make([]Column, t.ColumnLen())
	for i, col := range t.Columns() {
		// Column Name
		colName := unmarshalLang(lang, col.Name)
		col.Name = colName

		// Column headers
		headers := make([]string, len(col.Headers))
		for i, h := range col.Headers {
			headers[i] = unmarshalLang(lang, h)
		}
		col.Headers = headers

		cols[i] = col
	}

	// rows
	rows := make([]Row, 0)
	for _, row := range t.Rows() {
		var r = make([]Cell, len(row))
		for i, cell := range row {
			cellValue := unmarshalLang(lang, string(cell))
			r[i] = Cell(cellValue)
		}
		rows = append(rows, r)
	}

	res, _ := NewTable(cols, rows...)
	return res
}

// unmarshalLang the `target` and return the value for that `lang`. If
// the unmarshaler cannot properly do the job, it will return the
// `target` as it is.
func unmarshalLang(lang string, target string) string {
	var keys map[string]string
	err := json.Unmarshal([]byte(target), &keys)
	if err != nil {
		return target
	}
	return keys[lang]
}
