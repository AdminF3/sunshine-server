package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

type testContractTable struct {
	name     string
	project  uuid.UUID
	table    string
	status   int
	download bool
	user     *models.Document
}

var (
	contractFields = `
{
      "date-of-meeting": "33-02-1990",
      "address-of-building": "dragan tzankov 33",
      "meeting-opened-by": "mincho slivata",
      "chair-of-meeting": "Johny walker",
      "meeting-recorded-by": "Juliano recordioni",
      "measurement-implementer": "mincho slivata",
      "building-administrator": "bache kiko",
      "tab1-for-n": "18",
      "tab1-against-n": "3",
      "tab2-for-n": "18",
      "tab3-for-n": "18"
}`
	buildingAdmin = "bache kiko"

	agreementFields = `
{
      "assignor-address": "dragan tsankov 33",
      "assignor-bank-account-iban": "1234 1234 1234 1234",
      "assignor-bank-account-swift": "123",
      "assignor-name": "ass name",
      "assignor-bank-address": "tintyava 31"
}`
	bankIban = "1234 1234 1234 1234"

	maintenanceFields = `
{
	"test1": "da",
	"test2": "123"
}`
)

func TestGetTable(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)

		pm    = stores.NewTestUser(t, e.UserStore)
		p     = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		admin = stores.NewTestAdmin(t, e.UserStore)
		prj   = p.Data.(*models.Project)

		tt = []testContractTable{
			{"good pm", p.ID, "renovation_overall_budget", http.StatusOK, true, pm},
			{"good admin", p.ID, "renovation_overall_budget", http.StatusOK, true, admin},
			{"monitoring table", p.ID, "monitoring_phase_table", http.StatusOK, true, admin},
			{"bad project", uuid.New(), "renovation_overall_budget", http.StatusNotFound, false, pm},
			{"bad table", p.ID, "foo", http.StatusBadRequest, false, pm},
		}
	)
	defer del()

	for _, tc := range tt {
		r := httptest.NewRequest(
			"GET",
			fmt.Sprintf("/project/"+tc.project.String()+"/annex1/%s", tc.table),
			nil,
		)

		r = loginAs(t, e, tc.user, r)
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())
			if w.Code == http.StatusOK {
				var table contract.Table
				if err := json.Unmarshal(w.Body.Bytes(), &table); err != nil {
					t.Errorf("Can't unmarshal table: %q", err)
				}
			}

			if tc.download {
				t.Run("download", func(t *testing.T) {
					testDownload(t, e, tc, router, prj, "download/native")
					testDownload(t, e, tc, router, prj, "tex/native")
				})
			}

		})
	}
}

func TestGetEmptyFields(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)
		w      = httptest.NewRecorder()

		pm       = stores.NewTestUser(t, e.UserStore)
		prj      = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		contr, _ = stores.NewTestContract(t, e.ContractStore, prj)
		doc      = models.Document{Data: &contract.Contract{
			Fields: make(map[string]string),
		}}
	)

	defer del()

	r := httptest.NewRequest("GET", fmt.Sprintf("/project/%s/fields", contr.Data.(*contract.Contract).Project), nil)

	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc.Data.(*contract.Contract).Fields); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}
}

func TestGetAgreementFields(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)
		w      = httptest.NewRecorder()

		pm       = stores.NewTestUser(t, e.UserStore)
		prj      = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		contr, _ = stores.NewTestContract(t, e.ContractStore, prj)
		doc      = models.Document{Data: &contract.Contract{
			Agreement: make(map[string]string),
		}}
	)

	defer del()

	r := httptest.NewRequest("GET", fmt.Sprintf("/project/%s/agreement/fields", contr.Data.(*contract.Contract).Project), nil)

	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc.Data.(*contract.Contract).Agreement); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	t.Run("download", func(t *testing.T) {
		tc := testContractTable{
			project: prj.ID,
			user:    pm,
			status:  http.StatusOK,
		}
		project := prj.Data.(*models.Project)
		testDownload(t, e, tc, router, project, "agreement/download/native")
		testDownload(t, e, tc, router, project, "agreement/tex/english")
	})
}

func TestPutAndDownloadTable(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)

		pm  = stores.NewTestUser(t, e.UserStore)
		p   = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		prj = p.Data.(*models.Project)
		c   = contract.New(p.ID)

		tt = []testContractTable{
			{"good", p.ID, "renovation_overall_budget", http.StatusOK, true, pm},
			{"bad project", uuid.New(), "renovation_overall_budget", http.StatusNotFound, false, pm},
			{"bad table", p.ID, "foo", http.StatusBadRequest, false, pm},
		}
	)
	defer del()
	b, err := json.Marshal(c.Tables["renovation_overall_budget"])
	if err != nil {
		t.Fatalf("table marshal: %s", err)
	}

	for _, tc := range tt {
		r := httptest.NewRequest(
			"PUT",
			fmt.Sprintf("/project/"+tc.project.String()+"/annex1/%s", tc.table),
			bytes.NewReader(b),
		)

		r = loginAs(t, e, pm, r)
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())

			if w.Code == http.StatusOK {
				var table contract.Table
				if err := json.Unmarshal(w.Body.Bytes(), &table); err != nil {
					t.Errorf("Can't unmarshal table: %q", err)
				}

				if !reflect.DeepEqual(table, c.Tables["renovation_overall_budget"]) {
					t.Errorf("Got different tables.\nPre:\t%#v\nPost:\t%#v",
						table, c.Tables["renovation_overall_budget"])
				}
			}

			if tc.download {
				t.Run("download/pdf", func(t *testing.T) {
					testDownload(t, e, tc, router, prj, "download/english")
				})
				t.Run("download/tex", func(t *testing.T) {
					testDownload(t, e, tc, router, prj, "tex/english")
				})
			}
		})
	}
}

func TestUpdateFields(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)
		w      = httptest.NewRecorder()

		pm       = stores.NewTestUser(t, e.UserStore)
		prj      = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		contr, _ = stores.NewTestContract(t, e.ContractStore, prj)
		doc      = models.Document{Data: &contract.Contract{
			Fields: make(map[string]string),
		}}
	)

	defer del()

	r := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/project/%s/fields", contr.Data.(*contract.Contract).Project),
		strings.NewReader(contractFields),
	)

	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)

	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	if admin := doc.Data.(*contract.Contract).Fields["building-administrator"]; admin != buildingAdmin {
		t.Fatalf(
			"expected building administrator to be updated: expected %s,  got %s",
			buildingAdmin,
			admin,
		)
	}
}

func TestUpdateAgreementFields(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)
		w      = httptest.NewRecorder()

		pm       = stores.NewTestUser(t, e.UserStore)
		prj      = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		contr, _ = stores.NewTestContract(t, e.ContractStore, prj)
		doc      = models.Document{Data: &contract.Contract{
			Agreement: make(map[string]string),
		}}
	)

	defer del()

	r := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/project/%s/agreement/fields", contr.Data.(*contract.Contract).Project),
		strings.NewReader(agreementFields),
	)

	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}
	if iban := doc.Data.(*contract.Contract).Agreement["assignor-bank-account-iban"]; iban != bankIban {
		t.Fatalf(
			"expected bank iban to be updated: expected %s,  got %s",
			bankIban,
			iban,
		)
	}
}

func TestUpdateMaintenanceFields(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()

	pm := stores.NewTestUser(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	contr, _ := stores.NewTestContract(t, e.ContractStore, prj)
	doc := models.Document{Data: &contract.Contract{
		Maintenance: make(map[string]string),
	}}

	r := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/project/%s/maintenance/fields", contr.Data.(*contract.Contract).Project),
		strings.NewReader(maintenanceFields),
	)

	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}
	if t1 := doc.Data.(*contract.Contract).Maintenance["test1"]; t1 != "da" {
		t.Fatalf(
			"expected maintenance fields to be updated: expected %s,  got %s",
			"da",
			t1,
		)
	}
}

func TestGetMaintenanceFields(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()

	pm := stores.NewTestUser(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	contr, _ := stores.NewTestContract(t, e.ContractStore, prj)
	doc := models.Document{Data: &contract.Contract{
		Maintenance: map[string]string{"test1": "da"},
	}}

	r := httptest.NewRequest("GET", fmt.Sprintf("/project/%s/maintenance/fields", contr.Data.(*contract.Contract).Project), nil)

	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc.Data.(*contract.Contract).Maintenance); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}
}

func testDownload(t *testing.T, e *services.Env, tc testContractTable,
	router http.Handler, p *models.Project, suffix string) {

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/project/%s/%s", tc.project, suffix), nil)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusUnauthorized, w.Code, "<binary>")

	r = loginAs(t, e, tc.user, r)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	compareRespCode(t, tc.status, w.Code, w.Body.String())

	if w.Code == http.StatusOK {
		// If TeX is requested we don't care whether the content-type
		// is text/plain, application/x-tex, application/x-latex or
		// even application/octet-stream.
		//
		// The result from mime.TypeByExtension vary literally from
		// machine to machine as it reads /etc(/apache2?)?/mime.types
		if sp := strings.Split(suffix, "/"); sp[0] != "tex" && sp[len(sp)-2] != "tex" {
			exContextType := "application/pdf"
			contentType := w.Header().Get("Content-Type")
			if contentType != exContextType {
				t.Fatalf("Expected Content-Type to be %q got %q",
					exContextType, contentType)
			}
		}

		contentLength, _ := strconv.Atoi(w.Header().Get("Content-Length"))
		if contentLength == 0 {
			t.Fatal("Got zero Content-Length")
		}
	}
}

func TestMarkdown(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()
	router := New(e)

	prjDoc := stores.NewTestProject(t, e.ProjectStore)
	admin := stores.NewTestAdmin(t, e.UserStore)

	tc := []struct {
		name   string
		input  string
		output string
		user   *models.Document
		status int
	}{
		{
			name:   "ok",
			input:  "# foobar",
			output: "# foobar",
			user:   admin,
			status: http.StatusOK,
		},
		{
			name:   "unauthorized",
			input:  "# foobar",
			output: "",
			user:   stores.NewTestUser(t, e.UserStore),
			status: http.StatusUnauthorized,
		},
		{
			name:   "xss",
			input:  "<script>alert('pwn&d')</script>",
			output: "&lt;script&gt;alert(&#39;pwn&amp;d&#39;)&lt;/script&gt;",
			user:   admin,
			status: http.StatusOK,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			r := loginAs(t, e, c.user,
				httptest.NewRequest("PUT",
					"/project/"+prjDoc.ID.String()+"/markdown",
					bytes.NewBufferString(c.input)))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			putBody := strings.TrimSpace(w.Body.String())
			compareRespCode(t, c.status, w.Code, putBody)
			if putBody != c.output {
				t.Fatalf("On PUT got:\n%q\nGot:\n%q", c.output, putBody)
			}

			if w.Code == http.StatusOK {
				ct := w.Header().Get("content-type")
				if ct != "text/markdown" {
					t.Errorf("Expected Content-Type to be text/markdown got %q", ct)
				}
			}

			t.Run("get", func(t *testing.T) {
				r := loginAs(t, e, c.user,
					httptest.NewRequest("GET",
						"/project/"+prjDoc.ID.String()+"/markdown", nil))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, r)
				getBody := strings.TrimSpace(w.Body.String())
				compareRespCode(t, c.status, w.Code, getBody)
				if putBody != getBody {
					t.Fatalf("On PUT got:\n%q\nOn GET got:\n%q", putBody, getBody)
				}

				if w.Code == http.StatusOK {
					ct := w.Header().Get("content-type")
					if ct != "text/markdown" {
						t.Errorf("Expected Content-Type to be text/markdown got %q", ct)
					}
				}
			})
		})
	}
}

func TestDownloadAgreement(t *testing.T) {
	env, cleanup := newTestEnv(t)
	defer cleanup()

	admin := stores.NewTestAdmin(t, env.UserStore)
	prj := stores.NewTestProject(t, env.ProjectStore)

	router := New(env)

	cases := []struct {
		name   string
		target string
	}{
		{
			name:   "download/native/pdf",
			target: fmt.Sprintf("/project/%v/agreement/download/native", prj.ID),
		},
		{
			name:   "download/english/pdf",
			target: fmt.Sprintf("/project/%v/agreement/download/english", prj.ID),
		},
		{
			name:   "download/native/tex",
			target: fmt.Sprintf("/project/%v/agreement/tex/native", prj.ID),
		},
		{
			name:   "download/english/tex",
			target: fmt.Sprintf("/project/%v/agreement/tex/english", prj.ID),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := loginAs(
				t,
				env,
				admin,
				httptest.NewRequest("GET", c.target, nil),
			)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
		})
	}
}
