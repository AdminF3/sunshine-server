package contract

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
)

var (
	testdata = filepath.Join(config.Path, "contract", "testdata")
	dumptex  = flag.String("dumptex", "", "Dump PDF file to from test")

	ctx = context.Background()
)

func TestTeX(t *testing.T) {
	t.Parallel()

	c := New(uuid.New())
	glob := filepath.Join(testdata, "..", "tex", "latvia")

	pdf, err := NewDocumentFromPath(TemplateContext{
		Contract: *c,
		Project:  models.Project{Country: "Latvia"},
	}, glob)
	if err != nil {
		t.Fatalf("Did not expect error from NewPDF; got %q", err)
	}
	if err != nil {
		return
	}

	texFile, err := pdf.GenerateTeX(ctx, "contract.tex")
	if err == nil && texFile != nil {
		defer texFile.Close()
	}
	if err != nil {
		t.Fatalf("Did not expect error, got %v", err)
	}
}

func TestPDF(t *testing.T) {
	t.Parallel()
	var tt = []struct {
		name   string
		errGen bool
		errNew bool
	}{
		{name: "good"},
		{name: "bad_template", errGen: true}, // latex exec
		{name: "bad_tex", errGen: true},      // pdflatex
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				c    = New(uuid.New())
				glob = filepath.Join(testdata, tc.name)
			)

			if tc.name == "good" {
				glob = filepath.Join(testdata, "..", "tex", "latvia")

				markdown, err := ioutil.ReadFile("testdata/pandoc/input.md")
				if err != nil {
					t.Fatal(err)
				}
				c.Markdown = markdown
			}

			pdf, err := NewDocumentFromPath(TemplateContext{
				Contract: *c,
				Project:  models.Project{Country: "Latvia"},
			}, glob)
			if (err != nil) != tc.errNew {
				s := ""
				if !tc.errNew {
					s = "not"
				}
				t.Fatalf("Did %s expect error from NewPDF; got %q", s, err)
			}

			file, err := pdf.GeneratePDF(ctx, "contract.tex")
			if err == nil && file != nil {
				defer file.Close()
			}
			if (err != nil) != tc.errGen {
				s := ""
				if !tc.errGen {
					s = "not"
				}

				t.Fatalf("Did %s expect error, got %v", s, err)
			}

			if !tc.errGen {
				stat, err := file.Stat()
				if err != nil {
					t.Fatal(err)
				}
				if stat.Size() == 0 {
					t.Fatal("Generated empty PDF")
				}

				b, _ := ioutil.ReadAll(file)
				ctype := http.DetectContentType(b)
				if ctype != "application/pdf" {
					t.Fatalf("Content type: %q", ctype)
				}

				if len(*dumptex) > 0 {
					backup, err := os.Create(*dumptex)
					if err != nil {
						t.Fatal(err)
					}
					defer backup.Close()
					if _, err := backup.Write(b); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}

func TestPerCountry(t *testing.T) {
	t.Parallel()
	for _, c := range supportedCountries {
		if c.IsConsortium() {
			t.Run(c.String(), func(t *testing.T) {
				tctx := TemplateContext{
					Contract: *New(uuid.New()),
					Project:  models.Project{Country: c},
				}
				pdf, err := NewDocument(tctx)
				if err != nil {
					t.Fatal(err)
				}

				file, err := pdf.GeneratePDF(ctx, "contract.tex")
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
			})
		} else {
			t.Run(c.String(), func(t *testing.T) {
				tctx := TemplateContext{
					Contract: *New(uuid.New()),
					Project:  models.Project{Country: c},
				}
				pdf, err := NewDocumentFromLanguage(tctx, "latvia")
				if err != nil {
					t.Fatal(err)
				}

				file, err := pdf.GeneratePDF(ctx, "contract.tex")
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
			})
		}
	}
}

func TestMarkdownToLaTeX(t *testing.T) {
	t.Parallel()
	input, err := ioutil.ReadFile("testdata/pandoc/input.md")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := md2tex(ctx, &buf, input); err != nil {
		t.Fatal(err)
	}

	if buf.Len() == 0 {
		t.Error("md2tex produced 0 bytes")
	}
}
