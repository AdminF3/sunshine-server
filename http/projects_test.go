package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

const (
	validNewProjectTemplate = `{
    	"name": "%s",
    	"owner": "%s",
    	"asset": "%s",
	"roles": {
		"pm": [%q]
        },
    	"status": 1,
	"airtemp": 20,
	"watertemp": 40,
	"savings": 51.16,
	"portfolio_director": %q,
	"country": "Latvia"
    }`

	validProjectTemplate = `{
        "name": "Valid project - %s",
	"owner": "%s",
        "asset": "%s",
	"roles": {
		"pm": [%q]
        },
        "status": 1,
	"airtemp": 20,
	"watertemp": 40,
	"savings": 51.16,
	"portfolio_director": %q,
	"country": "Latvia"
	}`
	validProjectTemplateFM = `{
        "name": "Valid project FM - %s",
	"owner": "%s",
        "asset": "%s",
	"roles": {
		"pm": [%q]
        },
        "status": 1,
	"airtemp": 20,
	"watertemp": 40,
	"savings": 51.16,
	"portfolio_director": %q,
	"fund_manager": %q,
	"country": "Latvia"
	}`
	fmtUpdateHiddenColumnTemplate = `{
        "name": "GO PROJECT ColumnTPL",
	"owner": "%s",
        "asset": "%s",
	"roles": {
		"pm": [%q]
        },
        "status": 1,
	"airtemp": 20,
	"watertemp": 40,
	"savings": 51.16,
	"portfolio_director": %q,
	"country": "Latvia",
	"CreatedAt": %q
    }`
)

func TestProject(t *testing.T) {
	t.Run("create", testProjectCreate)
	t.Run("create/invalidOrg", testInvalidOrg)
	t.Run("create/invalidAsset", testInvalidAsset)
	t.Run("get", testProjectGet)
	t.Run("update", testProjectUpdate)
	t.Run("update/hiddenColumn", testProjectUpdateHiddenColumn)
	t.Run("update/PD", testProjectUpdatePD)
	t.Run("update/fundManager", testProjectUpdateFundManager)
	t.Run("list", testProjectList)
	t.Run("list/user", testProjectListByUser)
	t.Run("upload", testProjectUpload)
	t.Run("upload/delete", testProjectDeleteAttachment)
	t.Run("upload/delete/escapeASCII", testPrjDeleteEscapeSymbols)
	t.Run("roles", testProjectRoles)
	t.Run("upload/escapeASCII", testPrjUploadEscapeSymbols)
	t.Run("exportCSV", testExportMeetingsCSV)
}

func testExportMeetingsCSV(t *testing.T) {
	e, cleanup := newTestEnv(t)
	defer cleanup()

	router := New(e)
	w := httptest.NewRecorder()

	admin := stores.NewTestAdmin(t, e.UserStore)

	prj := stores.NewTestProject(t, e.ProjectStore)

	r := httptest.NewRequest("GET", "/project/"+prj.ID.String()+"/meetings", nil)

	// some meetings
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithPrj(prj.ID))
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithPrj(prj.ID))
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithPrj(prj.ID))
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithPrj(prj.ID))

	r = loginAs(t, e, admin, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if w.Body.Len() <= 1 {
		t.Errorf("expected csv content but got none")
	}
}

func testPrjUploadEscapeSymbols(t *testing.T) {
	e, cleanup := newTestEnv(t)
	router := New(e)
	w := httptest.NewRecorder()
	pm := stores.NewTestUser(t, e.UserStore)
	proj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	data := make(url.Values)

	defer cleanup()

	data.Add("kind", "logo")
	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "project", proj.ID, data, `g g.jpg`)
	r = loginAs(t, e, pm, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	gr := httptest.NewRequest("GET", "/project/"+proj.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	gr = loginAs(t, e, pm, gr)
	router.ServeHTTP(w, gr)
	compareRespCode(t, http.StatusOK, w.Code, "<binary>")

	f, err := ioutil.ReadFile(`testdata/g g.jpg`)
	if err != nil {
		t.Fatal(err)
	}

	if cl := w.Header().Get("Content-Length"); cl != strconv.Itoa(len(f)) {
		t.Errorf("Content-Length expected to be %d; got %s", len(f), cl)
	}
	if !bytes.Equal(f, w.Body.Bytes()) {
		t.Errorf("Files don't match. Expected %d bytes; got %d", len(f), len(w.Body.Bytes()))
	}
}

func testPrjDeleteEscapeSymbols(t *testing.T) {
	e, cleanup := newTestEnv(t)
	router := New(e)
	w := httptest.NewRecorder()
	pm := stores.NewTestUser(t, e.UserStore)
	proj := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	data := make(url.Values)

	defer cleanup()

	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "project", proj.ID, data, `g g.jpg`)
	r = loginAs(t, e, pm, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	dr := httptest.NewRequest("DELETE", "/project/"+proj.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	dr = loginAs(t, e, pm, dr)
	router.ServeHTTP(w, dr)
	compareRespCode(t, http.StatusOK, w.Code, "<binary>")

	gr := httptest.NewRequest("GET", "/project/"+proj.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	gr = loginAs(t, e, pm, gr)
	router.ServeHTTP(w, gr)
	compareRespCode(t, http.StatusNotFound, w.Code, "<binary>")
}

func testProjectCreate(t *testing.T) {
	var (
		e, del      = newTestEnv(t)
		router      = New(e)
		pd          = stores.NewTestUser(t, e.UserStore)
		user        = stores.NewTestUser(t, e.UserStore)
		invalid     = stores.NewTestUser(t, e.UserStore)
		org         = stores.NewTestOrg(t, e.OrganizationStore, user.ID, invalid.ID)
		asset       = stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))
		validNewPrj = fmt.Sprintf(validNewProjectTemplate, "valid", org.ID, asset.ID, user.ID, user.ID)

		tt = []struct {
			name      string
			body      string
			logged    *models.Document
			status    int
			valid     models.ValidationStatus
			superuser bool
		}{

			{
				name:   "valid but logged out",
				body:   fmt.Sprintf(validNewProjectTemplate, "valid but logged out", org.ID, asset.ID, user.ID, user.ID),
				status: http.StatusUnauthorized,
			},
			{
				name:      "valid",
				body:      validNewPrj,
				logged:    user,
				superuser: true,
				valid:     models.ValidationStatusValid,
				status:    http.StatusOK,
			},
			{
				name:      "valid",
				body:      validNewPrj,
				logged:    user,
				superuser: true,
				valid:     models.ValidationStatusValid,
				status:    http.StatusInternalServerError,
			},
			{
				name:      "invalid",
				body:      fmt.Sprintf(validNewProjectTemplate, "invalid", org.ID, asset.ID, user.ID, user.ID),
				logged:    invalid,
				superuser: false,
				status:    http.StatusOK,
				valid:     models.ValidationStatusRegistered,
			},
			{
				name:      "truncated payload",
				body:      validNewPrj[:len(validNewPrj)-1],
				superuser: true,
				valid:     models.ValidationStatusValid,
				logged:    stores.NewTestUser(t, e.UserStore),
				status:    http.StatusBadRequest,
			},
			{
				name:      "valid but not org member",
				body:      fmt.Sprintf(validNewProjectTemplate, "valid but not org member", org.ID, asset.ID, user.ID, user.ID),
				valid:     models.ValidationStatusValid,
				superuser: false,
				logged:    stores.NewTestUser(t, e.UserStore),
				status:    http.StatusUnauthorized,
			},
		}
	)
	defer del()
	pdst := stores.NewPortfolioStore(e.UserStore.DB())
	err := pdst.Put(ctx, pd.ID, models.CountryLatvia, models.PortfolioDirectorRole)
	if err != nil {
		t.Fatalf("Can't set PD: %v", err)
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				w   = httptest.NewRecorder()
				doc = models.Document{Data: &models.Project{}}
				r   = httptest.NewRequest("POST", "/project",
					strings.NewReader(tc.body))
			)

			if tc.logged != nil {
				ctx := context.Background()
				tc.logged.Data.(*models.User).Valid = tc.valid
				tc.logged.Data.(*models.User).SuperUser = tc.superuser

				if _, err := e.UserStore.Update(ctx, tc.logged); err != nil {
					t.Errorf("Error updating the user store: %v", err)
				}
				r = loginAs(t, e, tc.logged, r)
			}

			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())

			if w.Code != http.StatusOK {
				return
			}

			if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
				t.Fatalf("cant decode success response: %s", err)
			}

			location := w.Header().Get("Location")
			if !strings.HasSuffix(location, "/project/"+doc.ID.String()) {
				t.Errorf("Expected proper Location header; got %q", location)
			}

			_, err := e.ContractStore.GetByIndex(context.Background(), doc.ID.String())
			if err != nil {
				t.Errorf("Fetch contract failed: %v", err)
			}
		})
	}
}

func testInvalidOrg(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()
	router := New(e)
	user := stores.NewTestAdmin(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, user.ID)
	asset := stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))
	validNewPrj := fmt.Sprintf(validNewProjectTemplate, "invalid org", org.ID, asset.ID, user.ID, user.ID)

	org.Data.(*models.Organization).Valid = models.ValidationStatusRegistered
	_, err := e.OrganizationStore.Update(context.Background(), org)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/project",
		strings.NewReader(validNewPrj))

	r = loginAs(t, e, user, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusBadRequest, w.Code, w.Body.String())

	if msg := strings.TrimSpace(w.Body.String()); !strings.Contains(msg, "organization is not validated") {
		t.Errorf("Got %q expected %q", msg, "organization is not validated")
	}
}

func testInvalidAsset(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()
	router := New(e)
	user := stores.NewTestAdmin(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, user.ID)

	asset := stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))
	validNewPrj := fmt.Sprintf(validNewProjectTemplate, "invalid asset", org.ID, asset.ID, user.ID, user.ID)

	asset.Data.(*models.Asset).Valid = models.ValidationStatusDeclined
	_, err := e.AssetStore.Update(context.Background(), asset)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/project",
		strings.NewReader(validNewPrj))
	r = loginAs(t, e, user, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusBadRequest, w.Code, w.Body.String())

	if msg := strings.TrimSpace(w.Body.String()); !strings.Contains(msg, "asset is not validated") {
		t.Errorf("Got %q expected %q", msg, "asset is not validated")
	}

}
func testProjectGet(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()
	router := New(e)

	u := stores.NewTestUser(t, e.UserStore)
	p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(u.ID))
	stores.NewTestFA(t, e.FAStore, stores.TFAWithProject(p.ID))
	r := httptest.NewRequest("GET", "/project/"+p.ID.String(), nil)
	w := httptest.NewRecorder()
	doc := models.Document{Data: &models.Project{}}

	r = loginAs(t, e, u, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	prj := doc.Data.(*models.Project)
	if prj.ForfaitingApplication.ID == uuid.Nil {
		t.Fatal("fa was not found")
	}

	if prj.IsFAApproved == false {
		t.Fatal("fa should be approved")
	}

	if err := e.ProjectStore.Delete(context.Background(), &doc); err != nil {
		t.Fatalf("can't delete: %s", err)
	}

	// Reply same request once the asset is already deleted
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code 404 after deletion; got %d", w.Code)
	}
}

func testProjectUpdate(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	var (
		admin = stores.NewTestAdmin(t, e.UserStore)
		pm    = stores.NewTestUser(t, e.UserStore)
		teme  = stores.NewTestUser(t, e.UserStore)

		cases = []struct {
			name     string
			user     uuid.UUID
			prjRoles map[string][]uuid.UUID
			code     int
			request  func(*http.Request) *http.Request
			prjName  func(prj *models.Document) string
		}{
			{
				name:     "admin",
				user:     admin.ID,
				code:     http.StatusOK,
				prjRoles: map[string][]uuid.UUID{"pm": []uuid.UUID{pm.ID}},
				request:  func(r *http.Request) *http.Request { return loginAs(t, e, admin, r) },
				prjName:  func(_ *models.Document) string { return "Valid project - admin" },
			},
			{
				name:     "pm",
				user:     pm.ID,
				code:     http.StatusOK,
				prjRoles: map[string][]uuid.UUID{"pm": []uuid.UUID{pm.ID}},
				request:  func(r *http.Request) *http.Request { return loginAs(t, e, pm, r) },
				prjName:  func(_ *models.Document) string { return "Valid project - pm" },
			},
			{
				name:     "teme",
				user:     teme.ID,
				code:     http.StatusUnauthorized,
				prjRoles: map[string][]uuid.UUID{"teme": []uuid.UUID{teme.ID}},
				request:  func(r *http.Request) *http.Request { return loginAs(t, e, teme, r) },
				prjName:  func(p *models.Document) string { return p.Data.(*models.Project).Name },
			},
			{
				name:     "not-logged",
				user:     uuid.New(),
				code:     http.StatusUnauthorized,
				prjRoles: map[string][]uuid.UUID{"pm": []uuid.UUID{pm.ID}},
				request:  func(r *http.Request) *http.Request { return r },
				prjName:  func(p *models.Document) string { return p.Data.(*models.Project).Name },
			},
		}

		router = New(e)
	)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithRole(c.prjRoles))
			validUpdatedPrj := fmt.Sprintf(validProjectTemplate, c.name, p.Data.Dependencies()[0].ID, p.Data.Dependencies()[1].ID, c.user, c.user)
			r := httptest.NewRequest("PUT", "/project/"+p.ID.String(), strings.NewReader(validUpdatedPrj))

			router.ServeHTTP(w, c.request(r))

			compareRespCode(t, c.code, w.Code, w.Body.String())

			if c.code == http.StatusOK {
				doc := models.Document{Data: &models.Project{}}
				if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
					t.Fatalf("can't decode success response: %s", err)
				}

				name := doc.Data.(*models.Project).Name
				if name != c.prjName(p) {
					t.Fatalf("expected name to be updated, got: %s", name)
				}
			}

			nDoc, nerr := e.ProjectStore.Get(context.Background(), p.ID)
			if nerr != nil {
				t.Fatalf("can't get newly updated document: %v", nerr)
			}

			nName := nDoc.Data.(*models.Project).Name
			if nName != c.prjName(p) {
				t.Fatalf("expected name to be %s; got %s", c.prjName(p), nName)
			}
		})
	}
}

func testProjectUpdateFundManager(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	admin := stores.NewTestAdmin(t, e.UserStore)
	pm := stores.NewTestUser(t, e.UserStore)
	pd := stores.NewTestPortfolioRole(t, e.UserStore, models.PortfolioDirectorRole, "Latvia")
	fm := stores.NewTestUser(t, e.UserStore)
	prjRoles := map[string][]uuid.UUID{"pm": []uuid.UUID{pm.ID}}

	cases := []struct {
		name string
		user *models.Document
		code int
	}{
		{
			name: "admin",
			user: admin,
			code: http.StatusOK,
		},
		{
			name: "pd",
			user: pd,
			code: http.StatusOK,
		},
		{
			name: "pm",
			user: pm,
			code: http.StatusUnauthorized,
		},
	}

	router := New(e)

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithRole(prjRoles))
			u := c.user.ID
			validUpdatedPrj := fmt.Sprintf(validProjectTemplateFM, c.name, p.Data.Dependencies()[0].ID, p.Data.Dependencies()[1].ID, u, u, fm.ID)

			r := httptest.NewRequest("PUT", "/project/"+p.ID.String(), strings.NewReader(validUpdatedPrj))
			r = loginAs(t, e, c.user, r)
			router.ServeHTTP(w, r)
			compareRespCode(t, c.code, w.Code, w.Body.String())

			if w.Code == http.StatusOK {
				doc := models.Document{Data: &models.Project{}}
				if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
					t.Fatalf("can't decode success response: %s", err)
				}

				nFM := doc.Data.(*models.Project).FundManager
				if *nFM != uuid.UUID(fm.ID) {
					t.Fatalf("expected fund manager to be %v; got %v", fm.ID, nFM)
				}

			}
			if c.code != http.StatusOK {

				nDoc, nerr := e.ProjectStore.Get(context.Background(), p.ID)

				if nerr != nil {
					t.Fatalf("can't get newly updated document: %v", nerr)
				}

				nFM := nDoc.Data.(*models.Project).FundManager
				if nFM != nil && *nFM == fm.ID {
					t.Fatalf("expected fund manager not be changed")
				}
			}

		})
	}
}

func testProjectUpdateHiddenColumn(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	admin := stores.NewTestAdmin(t, e.UserStore)
	pm := stores.NewTestUser(t, e.UserStore)

	name := "hiddenColumn"
	user := admin.ID
	code := http.StatusOK
	testTime := time.Date(2000, time.January, 15, 10, 20, 30, 0, time.UTC)
	router := New(e)

	t.Run(name, func(t *testing.T) {
		w := httptest.NewRecorder()
		p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		validUpdatedPrj := fmt.Sprintf(fmtUpdateHiddenColumnTemplate, p.Data.Dependencies()[0].ID, p.Data.Dependencies()[1].ID, user, user, testTime.Format(time.RFC3339))
		r := httptest.NewRequest("PUT", "/project/"+p.ID.String(), strings.NewReader(validUpdatedPrj))
		r = loginAs(t, e, admin, r)
		router.ServeHTTP(w, r)

		compareRespCode(t, code, w.Code, w.Body.String())
		doc := models.Document{Data: &models.Project{}}
		if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
			t.Fatalf("can't decode success response: %s", err)
		}

		createdAt := doc.Data.(*models.Project).Value.CreatedAt

		if testTime.Equal(createdAt) {
			t.Fatalf("The user shouldn't be able to change CreatedAt column ; got %v", createdAt)
		}

	})
}

func testProjectUpdatePD(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	var (
		admin = stores.NewTestAdmin(t, e.UserStore)
		pm    = stores.NewTestUser(t, e.UserStore)
		teme  = stores.NewTestUser(t, e.UserStore)
		dir   = stores.NewTestUser(t, e.UserStore)

		// The PD must be valid member of any org
		_ = stores.NewTestOrg(t, e.OrganizationStore, dir.ID)

		roles = map[string][]uuid.UUID{
			"pm":   []uuid.UUID{pm.ID},
			"teme": []uuid.UUID{teme.ID},
		}

		cases = []struct {
			name     string
			request  func(*http.Request) *http.Request
			code     int
			director func(*models.Document) uuid.UUID
		}{
			{
				name: "admin",
				code: http.StatusOK,
				request: func(r *http.Request) *http.Request {
					return loginAs(t, e, admin, r)
				},
				director: func(_ *models.Document) uuid.UUID {
					return dir.ID
				},
			},
			{
				name: "pm",
				code: http.StatusOK,
				request: func(r *http.Request) *http.Request {
					return loginAs(t, e, pm, r)
				},
				director: func(p *models.Document) uuid.UUID {
					return p.Data.(*models.Project).PortfolioDirector
				},
			},
			{
				name: "teme",
				code: http.StatusUnauthorized,
				request: func(r *http.Request) *http.Request {
					return loginAs(t, e, teme, r)
				},
				director: func(p *models.Document) uuid.UUID {
					return p.Data.(*models.Project).PortfolioDirector
				},
			},
			{
				name: "not logged in",
				code: http.StatusUnauthorized,
				request: func(r *http.Request) *http.Request {
					return r
				},
				director: func(p *models.Document) uuid.UUID {
					return p.Data.(*models.Project).PortfolioDirector
				},
			},
		}

		router = New(e)
	)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithRole(roles))
			validUpdatedPrj := fmt.Sprintf(validProjectTemplate, c.name, p.Data.Dependencies()[0].ID, p.Data.Dependencies()[1].ID, pm.ID, dir.ID)
			r := httptest.NewRequest("PUT", "/project/"+p.ID.String(), strings.NewReader(validUpdatedPrj))

			router.ServeHTTP(w, c.request(r))

			compareRespCode(t, c.code, w.Code, w.Body.String())

			nDoc, nerr := e.ProjectStore.Get(context.Background(), p.ID)
			if nerr != nil {
				t.Fatalf("can't get newly updated document: %v", nerr)
			}

			npd := nDoc.Data.(*models.Project).PortfolioDirector
			if npd != c.director(p) {
				t.Fatalf("expected portfolio director to be %s; got %s", c.director(p), npd)
			}
		})
	}
}

func testProjectList(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	w := httptest.NewRecorder()
	router := New(e)

	u := stores.NewTestAdminNwManager(t, e.UserStore)

	cases := []struct {
		name string
		path string
	}{
		{
			name: "all",
			path: "/project",
		},
		{
			name: "all reports",
			path: "/project/reports",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			router.ServeHTTP(w, loginAs(t, e, u,
				httptest.NewRequest("GET", c.path, nil)))
			compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

			var docs dummyList
			if err := json.NewDecoder(w.Body).Decode(&docs); err != nil {
				t.Errorf("decoding failed: %s", err)
			}

			count, err := strconv.Atoi(w.Header().Get(countHeader))
			if err != nil {
				t.Errorf("%q header is not set: %s", countHeader, err)
			}

			if len(docs.Documents) > count {
				t.Errorf("%q header reports %d records, got %d",
					countHeader, count, len(docs.Documents))
			}
		})
	}
}

func testProjectListByUser(t *testing.T) {
	var docs dummyList

	e, del := newTestEnv(t)
	defer del()

	w := httptest.NewRecorder()
	router := New(e)

	pm := stores.NewTestUser(t, e.UserStore)
	prog := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	user := prog.Data.(*models.Project).Roles.PM[0]
	stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
	stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))

	router.ServeHTTP(w, loginAs(t, e, pm,
		httptest.NewRequest("GET", fmt.Sprintf("/user/%s/projects", user), nil)))

	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&docs); err != nil {
		t.Errorf("decoding failed %s", err)
	}

	count, err := strconv.Atoi(w.Header().Get(countHeader))
	if err != nil {
		t.Errorf("%q header is not set: %s", countHeader, err)
	}

	if len(docs.Documents) != count || len(docs.Documents) != 1 {
		t.Errorf("%q header reports %d records, got %d",
			countHeader, count, len(docs.Documents))
	}
}

func TestNotAllowedProjectMethods(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)
		undef  = []string{"HEAD", "DELETE", "TRACE"}
	)
	defer del()

	for _, method := range undef {
		t.Run(method, func(t *testing.T) {
			var w = httptest.NewRecorder()

			router.ServeHTTP(w,
				httptest.NewRequest(method, "/project", nil),
			)
			compareRespCode(t, http.StatusMethodNotAllowed,
				w.Code, w.Body.String(),
			)
		})
	}
}

func testProjectUpload(t *testing.T) {
	var cases = []struct {
		name  string
		utype string
		code  int
	}{
		{
			name:  "empty",
			utype: "",
			code:  http.StatusOK,
		},
		{
			name:  "not_existing",
			utype: "not_existing",
			code:  http.StatusBadRequest,
		},
		{
			name:  "ok",
			utype: "general leaflet",
			code:  http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			var (
				e, del = newTestEnv(t)
				router = New(e)
				w      = httptest.NewRecorder()
				pm     = stores.NewTestUser(t, e.UserStore)
				p      = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
				getr   = httptest.NewRequest("GET", "/project/"+p.ID.String(), nil)
				doc    = models.Document{Data: &models.Project{}}
				data   = make(url.Values)
			)
			defer del()

			if tc.utype != "" {
				data.Add("upload-type", tc.utype)
			}

			r := createTestFileRequest(t, "project", p.ID, data)
			r = loginAs(t, e, pm, r)

			router.ServeHTTP(w, r)
			compareRespCode(t, tc.code, w.Code, w.Body.String())

			if tc.code != http.StatusOK {
				return
			}

			w = httptest.NewRecorder()
			getr = loginAs(t, e, pm, getr)
			router.ServeHTTP(w, getr)
			compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

			if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
				t.Errorf("decoding failed: %s", err)
			}

			if len(doc.Attachments) < 1 {
				t.Errorf("no attachments in document")
			}
		})
	}
}

func testProjectDeleteAttachment(t *testing.T) {
	var (
		e, cleanup = newTestEnv(t)
		router     = New(e)
		w          = httptest.NewRecorder()
		pm         = stores.NewTestUser(t, e.UserStore)
		proj       = stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
		getr       = httptest.NewRequest("GET", "/project/"+proj.ID.String(), nil)
		doc        = models.Document{Data: &models.Project{}}
		data       = make(url.Values)
	)
	defer cleanup()

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "project", proj.ID, data)
	r = loginAs(t, e, pm, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	delr := httptest.NewRequest("DELETE", "/project/"+proj.ID.String()+"/gg.jpg", nil)
	delr = loginAs(t, e, pm, delr)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, delr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()
	getr = loginAs(t, e, pm, getr)
	router.ServeHTTP(w, getr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Errorf("decoding failed:%s", err)
	}

	if len(doc.Attachments) > 0 {
		t.Errorf("Expected file to be deleted; attachments: %v", doc.Attachments)
	}
}

func testProjectRoles(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	user := stores.NewTestUser(t, e.UserStore)

	pm := stores.NewTestUser(t, e.UserStore)
	err := e.UserStore.Portfolio().Put(ctx, pm.ID, models.CountryLatvia, models.PortfolioDirectorRole)
	if err != nil {
		t.Fatalf("Can't set PD: %v", err)
	}
	proj := stores.NewTestProject(t, e.ProjectStore)

	before := len(proj.Data.(*models.Project).Roles.TaMa)
	before += len(proj.Data.(*models.Project).Roles.PM) - 1 // don't count the default PM

	cases := []struct {
		name         string
		logged       *models.Document
		statusAdd    int
		statusRemove int
		alter        string
	}{
		{
			name:         "ok",
			logged:       pm,
			statusAdd:    http.StatusOK,
			statusRemove: http.StatusOK,
			alter:        fmt.Sprintf(`{"position": "tama", "user": %q}`, user.ID),
		},
		{
			name:         "unauth",
			statusAdd:    http.StatusUnauthorized,
			statusRemove: http.StatusUnauthorized,
			alter:        fmt.Sprintf(`{"position": "tama", "user": %q}`, user.ID),
		},
		{
			name:         "random role",
			logged:       pm,
			statusAdd:    http.StatusBadRequest,
			statusRemove: http.StatusBadRequest,
			alter:        fmt.Sprintf(`{"position": "g0d", "user": %q}`, user.ID),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			router.ServeHTTP(w, loginAs(t, e, c.logged,
				httptest.NewRequest("POST", "/project/"+proj.ID.String()+"/roles",
					bytes.NewBufferString(c.alter))))
			compareRespCode(t, c.statusAdd, w.Code, w.Body.String())

			if c.statusAdd == http.StatusOK {
				doc := models.Document{Data: new(models.Project)}
				if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
					t.Errorf("decoding failed: %s", err)
				}

				roles := doc.Data.(*models.Project).Roles
				after := len(roles.TaMa)
				after += len(roles.PM) - 1 // again, don't count the default PM

				if before+1 != after {
					t.Errorf("Expected one new role to have been added; got: %v, expected: %v", after, before+1)
				}

			}

			router.ServeHTTP(w, loginAs(t, e, c.logged,
				httptest.NewRequest("DELETE", "/project/"+proj.ID.String()+"/roles",
					bytes.NewBufferString(c.alter))))
			compareRespCode(t, c.statusRemove, w.Code, w.Body.String())

			if c.statusRemove == http.StatusOK {
				doc := models.Document{Data: new(models.Project)}
				if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
					t.Errorf("decoding failed: %s", err)
				}

				roles := doc.Data.(*models.Project).Roles
				if len(roles.TaMa) != 0 {
					t.Errorf("Expected no removed role to remain; got: %v", len(roles.TaMa))
				} else if len(roles.PM) != 1 {
					t.Errorf("Expected no removed role to remain; got: %v", len(roles.PM))
				}

			}
		})
	}

}
