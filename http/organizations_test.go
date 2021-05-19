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
	fmtValidNewOrg = `{
		"name": "Goo Corporation",
		"vat": "%s",
		"address": "End of the world",
		"telephone": "+359888123456",
		"legal_form": 4,
		"registered": "2017-10-02T12:30:10Z",
		"valid": 3,
		"roles": {
			"lear":"%s",
			"lsigns":null,
			"leaas":null,
                        "members": null
		},
		"email": "orgmail@goo.com",
		"country": "Latvia",
		"registration_number": "test_reg_number"
	}`
	fmtValidUpdatedOrg = `{
		"name": "Goo Evil Corporation",
		"vat": "6388047V",
		"address": "End of the world",
		"telephone": "+359888123456",
		"legal_form": 4,
		"registered": "2017-10-02T12:30:10Z",
		"valid": 1,
		"roles": {
			"lear":"%s",
			"lsigns":null,
			"leaas":["%s"],
			"members": null
		},
		"email": "orgmail@goo.com",
		"registration_number": "updated_test_reg_number"
	}`

	fmtHiddenColumnUpdatedOrg = `{
		"name": "Goo Evil Corporation",
		"vat": "6388047V",
		"address": "End of the world",
		"telephone": "+359888123456",
		"legal_form": 4,
		"registered": "2017-10-02T12:30:10Z",
		"valid": 1,
		"roles": {
			"lear":"%s",
			"lsigns":null,
			"leaas":null,
                        "members": null
		},
		"email": "orgmail@goo.com",
		"CreatedAt": %q
	}`
	orgNewName = "Goo Evil Corporation"
)

func TestOrganization(t *testing.T) {
	t.Run("create", testOrganizationCreate)
	t.Run("get", testOrganizationGet)
	t.Run("update", testOrganizationUpdate)
	t.Run("update/hiddenColumn", testOrganizationUpdateHiddenColumn)
	t.Run("list", testOrganizationList)
	t.Run("listByMember", testOrganizationListByUser)
	t.Run("upload", testOrganizationUpload)
	t.Run("upload/delete", testOrganizationDeleteAttachment)
	t.Run("roles", testOrganizationAddRemoveRole)
	t.Run("upload/escapeASCII", testOrgUploadEscapeSymbols)
	t.Run("upload/delete/escapeASCII", testOrgDeleteEscapeSymbols)
}

func testOrgUploadEscapeSymbols(t *testing.T) {
	e, cleanup := newTestEnv(t)
	router := New(e)
	w := httptest.NewRecorder()
	u := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, u.ID)
	data := make(url.Values)

	defer cleanup()

	data.Add("kind", "logo")
	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "organization", org.ID, data, `g g.jpg`)
	r = loginAs(t, e, u, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	gr := httptest.NewRequest("GET", "/organization/"+org.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	gr = loginAs(t, e, u, gr)
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

func testOrgDeleteEscapeSymbols(t *testing.T) {
	e, cleanup := newTestEnv(t)
	router := New(e)
	w := httptest.NewRecorder()
	u := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, u.ID)
	data := make(url.Values)

	defer cleanup()

	data.Add("kind", "logo")
	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "organization", org.ID, data, `g g.jpg`)
	r = loginAs(t, e, u, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	dr := httptest.NewRequest("DELETE", "/organization/"+org.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	dr = loginAs(t, e, u, dr)
	router.ServeHTTP(w, dr)
	compareRespCode(t, http.StatusOK, w.Code, "<binary>")

	gr := httptest.NewRequest("GET", "/organization/"+org.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	gr = loginAs(t, e, u, gr)
	router.ServeHTTP(w, gr)
	compareRespCode(t, http.StatusNotFound, w.Code, "<binary>")
}

func testOrganizationCreate(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	authUser := stores.NewTestUser(t, e.UserStore)
	validOrg := fmt.Sprintf(fmtValidNewOrg, authUser.ID, uuid.New())

	tt := []struct {
		name   string
		body   string
		user   *models.Document // empty means logged out
		valid  models.ValidationStatus
		status int
	}{
		{
			name:   "valid",
			body:   fmt.Sprintf(fmtValidNewOrg, uuid.New(), authUser.ID),
			user:   authUser,
			status: http.StatusOK,
			valid:  models.ValidationStatusValid,
		},
		{
			name:   "invalid",
			body:   fmt.Sprintf(fmtValidNewOrg, uuid.New(), authUser.ID),
			status: http.StatusUnauthorized,
			valid:  models.ValidationStatusRegistered,
		},
		{
			name:   "truncated payload",
			body:   validOrg[:len(validOrg)-1],
			user:   authUser,
			status: http.StatusBadRequest,
			valid:  models.ValidationStatusValid,
		},
		{
			name: "missing registered",
			body: fmt.Sprintf(`{
				"name": "Goo Corporation",
				"address": "End of the world",
				"telephone": "+359888123456",
				"website": "https://goocorp.example",
				"legal_form": 5,
				"valid": 1,
				"roles": {
					"lear":"%s",
					"lsigns":null,
					"leaas":null
				},
				"country": "Latvia",
				"vat": "%s"
			}`, authUser.ID, uuid.New()),
			user:   authUser,
			status: http.StatusBadRequest,
		},
		{
			name: "Residents comunity",
			body: fmt.Sprintf(`{
				"name": "Residents community",
				"address": "End of the world",
				"telephone": "+359888123456",
				"legal_form": 4,
				"valid": 1,
				"roles": {
					"lear":"%s",
					"lsigns":null,
					"leaas":null
				},
				"country": "Latvia"
			}`, authUser.ID),
			user:   authUser,
			status: http.StatusOK,
		},
		{
			name:   "valid but logged out",
			body:   fmt.Sprintf(fmtValidNewOrg, uuid.New(), authUser.ID),
			status: http.StatusUnauthorized,
			valid:  models.ValidationStatusValid,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			doc := models.Document{Data: &models.Organization{}}
			r := httptest.NewRequest("POST", "/organization",
				strings.NewReader(tc.body))

			if tc.user != nil {
				tc.user.Data.(*models.User).Valid = tc.valid
				if _, err := e.UserStore.Update(ctx, tc.user); err != nil {
					t.Errorf("Error updating the user to store: %v", err)
				}
				r = loginAs(t, e, tc.user, r)
			}

			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())

			if w.Code != http.StatusOK {
				return
			}

			if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
				t.Fatalf("can't decode success response: %s", err)
			}

			org := doc.Data.(*models.Organization)
			if org.Valid != models.ValidationStatusRegistered {
				t.Errorf("should be 'Registered' by default; got %q", org.Valid)
			}
			location := w.Header().Get("Location")
			if !strings.HasSuffix(location, "/organization/"+doc.ID.String()) {
				t.Errorf("Expected proper Location header; got %q",
					location)
			}
		})
	}
}

func testOrganizationGet(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)

		o   = stores.NewTestOrg(t, e.OrganizationStore)
		u   = stores.NewTestUser(t, e.UserStore)
		r   = httptest.NewRequest("GET", "/organization/"+o.ID.String(), nil)
		w   = httptest.NewRecorder()
		doc = models.Document{Data: &models.Organization{}}
	)
	defer del()

	r = loginAs(t, e, u, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	if err := e.OrganizationStore.Delete(context.Background(), &doc); err != nil {
		t.Fatalf("can't delete: %s", err)
	}

	// Reply same request once the organization is already deleted
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {

		t.Errorf("expected status code 404 after deletion; got %d", w.Code)
	}
}

func testOrganizationUpdate(t *testing.T) {

	e, del := newTestEnv(t)
	defer del()
	router := New(e)

	u := stores.NewTestUser(t, e.UserStore)
	leaa := stores.NewTestUser(t, e.UserStore)
	randomGuy := stores.NewTestUser(t, e.UserStore)
	o := stores.NewTestOrg(t, e.OrganizationStore, u.ID, leaa.ID)
	validUpdOrg := fmt.Sprintf(fmtValidUpdatedOrg, u.ID, leaa.ID)

	cases := []struct {
		id       uuid.UUID
		name     string
		user     *models.Document
		responce int
	}{
		{

			name:     "lear",
			id:       o.ID,
			user:     u,
			responce: http.StatusOK,
		},
		{
			name:     "leaa",
			id:       o.ID,
			user:     leaa,
			responce: http.StatusOK,
		},
		{
			name:     "randomGuy",
			id:       o.ID,
			user:     randomGuy,
			responce: http.StatusUnauthorized,
		},
		{
			name:     "notLogged",
			id:       o.ID,
			user:     nil,
			responce: http.StatusUnauthorized,
		},
	}

	doc := models.Document{Data: &models.Organization{}}
	for _, c := range cases {
		w := httptest.NewRecorder()
		router.ServeHTTP(w,
			loginAs(t, e, c.user,
				httptest.NewRequest("PUT", "/organization/"+c.id.String(),
					strings.NewReader(validUpdOrg)),
			),
		)
		compareRespCode(t, c.responce, w.Code, w.Body.String())
		if c.responce == http.StatusOK {
			if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
				t.Fatalf("can't decode success response: %s", err)
			}

			if newName := doc.Data.(*models.Organization).Name; newName != orgNewName {
				t.Fatalf("expected name to be updated; got %s", newName)
			}
		}

	}
	// Reply same request once the organization is already deleted
	w := httptest.NewRecorder()
	if err := e.OrganizationStore.Delete(context.Background(), &doc); err != nil {
		t.Fatalf("can't delete: %s", err)
	}

	router.ServeHTTP(w,
		loginAs(t, e, u,
			httptest.NewRequest("PUT", "/organization/"+o.ID.String(),
				strings.NewReader(validUpdOrg)),
		),
	)
	compareRespCode(t, http.StatusNotFound, w.Code, w.Body.String())
}

func testOrganizationUpdateHiddenColumn(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	testTime := time.Date(2000, time.January, 15, 10, 20, 30, 0, time.UTC)
	u := stores.NewTestUser(t, e.UserStore)
	o := stores.NewTestOrg(t, e.OrganizationStore, u.ID)
	validUpdOrg := fmt.Sprintf(fmtHiddenColumnUpdatedOrg, u.ID, testTime.Format(time.RFC3339))

	w := httptest.NewRecorder()
	doc := models.Document{Data: &models.Organization{}}

	router.ServeHTTP(w,
		loginAs(t, e, u,
			httptest.NewRequest("PUT", "/organization/"+o.ID.String(),
				strings.NewReader(validUpdOrg)),
		),
	)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	createdAt := doc.Data.(*models.Organization).Value.CreatedAt
	if testTime.Equal(createdAt) {
		t.Fatalf("The user shouldn't be able to change CreatedAt column ; got %v", createdAt)
	}

}

func testOrganizationList(t *testing.T) {
	var (
		docs   dummyList
		w      = httptest.NewRecorder()
		e, del = newTestEnv(t)
		router = New(e)
		u      = stores.NewTestUser(t, e.UserStore)
	)
	defer del()
	listr := httptest.NewRequest("GET", "/organization", nil)
	listr = loginAs(t, e, u, listr)

	router.ServeHTTP(w, listr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

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

}

func testOrganizationListByUser(t *testing.T) {
	var (
		docs   dummyList
		w      = httptest.NewRecorder()
		e, del = newTestEnv(t)
		router = New(e)
		user   = stores.NewTestUser(t, e.UserStore)
	)
	defer del()

	// Create couple of organizations with different users
	stores.NewTestOrg(t, e.OrganizationStore, user.ID)
	stores.NewTestOrg(t, e.OrganizationStore, user.ID)
	stores.NewTestOrg(t, e.OrganizationStore, user.ID)
	stores.NewTestOrg(t, e.OrganizationStore)
	stores.NewTestOrg(t, e.OrganizationStore)

	// Create one and replace the user from it with a new one. This makes
	// sure a removed member no longer receives that organization in member
	// query. See #174.
	newLEAR := stores.NewTestUser(t, e.UserStore)
	toDel := stores.NewTestOrg(t, e.OrganizationStore, newLEAR.ID, user.ID)
	toDelOrg := toDel.Data.(*models.Organization)
	for i, r := range toDelOrg.OrganizationRoles {
		if r.Position != "leaa" {
			continue
		}
		toDelOrg.Roles.LEAAs = nil
		toDelOrg.OrganizationRoles = append(
			toDelOrg.OrganizationRoles[:i],
			toDelOrg.OrganizationRoles[i+1:]...)
		if err := stores.AtomicDelete(e.OrganizationStore, r); err != nil {
			t.Fatal(err)
		}
	}
	toDel.Data = toDelOrg
	if _, err := e.OrganizationStore.Update(context.Background(), toDel); err != nil {
		t.Fatalf("Remove user from organization: %v", err)
	}

	router.ServeHTTP(w, httptest.NewRequest("GET", fmt.Sprintf("/user/%s/organizations?limit=4", user.ID), nil))
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&docs); err != nil {
		t.Errorf("decoding failed %s", err)
	}

	count, err := strconv.Atoi(w.Header().Get(countHeader))
	if err != nil {
		t.Errorf("%q header is not set: %s", countHeader, err)
	}

	if len(docs.Documents) != 3 {
		t.Errorf("%q header reports %d records, got %d",
			countHeader, count, len(docs.Documents))
	}
}

func TestNotAllowedOrganizationMethods(t *testing.T) {
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
				httptest.NewRequest(method, "/organization", nil),
			)
			compareRespCode(t, http.StatusMethodNotAllowed,
				w.Code, w.Body.String(),
			)
		})
	}
}

func testOrganizationUpload(t *testing.T) {
	var tt = []struct {
		name  string
		kind  string
		utype string
		code  int
	}{
		{
			name:  "bad",
			kind:  "",
			utype: "nan",
			code:  http.StatusOK,
		},
		{
			name:  "logo",
			kind:  "logo",
			utype: "general leaflet",
			code:  http.StatusOK,
		},
		{
			name:  "no-kind",
			kind:  "",
			utype: "general leaflet",
			code:  http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			e, del := newTestEnv(t)
			router := New(e)
			lear := stores.NewTestUser(t, e.UserStore)
			leaa := stores.NewTestUser(t, e.UserStore)
			randomGuy := stores.NewTestUser(t, e.UserStore)
			o := stores.NewTestOrg(t, e.OrganizationStore, lear.ID, leaa.ID)
			getr := httptest.NewRequest("GET", "/organization/"+o.ID.String(), nil)
			doc := models.Document{Data: &models.Organization{}}
			data := make(url.Values)

			defer del()

			if tc.name != "bad" {
				data.Add("kind", tc.kind)
				data.Add("upload-type", tc.utype)
			}

			// Test wit lear
			w := httptest.NewRecorder()
			r := createTestFileRequest(t, "organization", o.ID, data)
			r = loginAs(t, e, lear, r)
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.code, w.Code, w.Body.String())
			if tc.code != http.StatusOK {
				return
			}
			if tc.utype == "survey" {
				// Test wit leaa
				w = httptest.NewRecorder()
				r = createTestFileRequest(t, "organization", o.ID, data)
				r = loginAs(t, e, leaa, r)
				router.ServeHTTP(w, r)
				compareRespCode(t, tc.code, w.Code, w.Body.String())
				if tc.code != http.StatusOK {
					return
				}

				// Test with randomGuy
				w = httptest.NewRecorder()
				r = createTestFileRequest(t, "organization", o.ID, data)
				r = loginAs(t, e, randomGuy, r)
				router.ServeHTTP(w, r)
				compareRespCode(t, http.StatusUnauthorized, w.Code, w.Body.String())
				if tc.code != http.StatusOK {
					return
				}

			}

			w = httptest.NewRecorder()
			getr = loginAs(t, e, lear, getr)
			router.ServeHTTP(w, getr)
			compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

			if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
				t.Errorf("decoding failed: %s", err)
			}

			if len(doc.Attachments) < 1 {
				t.Errorf("no attachments in document")
			}

			if tc.name == "logo" {
				logo := doc.Data.(*models.Organization).Logo
				testSpecialUpload(e, t, "organization", tc.name, logo, lear, doc)

				t.Run("file", func(t *testing.T) {
					compareFile(t, e, router, "testdata/gg.jpg", logo, nil)
				})
			}

		})
	}

}

func testOrganizationAddRemoveRole(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	user := stores.NewTestUser(t, e.UserStore)
	lear := stores.NewTestUser(t, e.UserStore)
	leaa := stores.NewTestUser(t, e.UserStore)
	leaa2 := stores.NewTestUser(t, e.UserStore)

	org := stores.NewTestOrg(t, e.OrganizationStore,
		lear.ID, leaa.ID, leaa2.ID)
	before := len(org.Data.(*models.Organization).Roles.LEAAs)

	cases := []struct {
		name         string
		user         *models.Document
		learID       uuid.UUID
		statusAdd    int
		statusRemove int
		alter        string
	}{
		{
			name:         "valid",
			user:         lear,
			learID:       org.Data.Dependencies()[0].ID,
			statusAdd:    http.StatusOK,
			statusRemove: http.StatusOK,
			alter:        fmt.Sprintf(`{"position": "leaa", "user": %q}`, user.ID),
		},
		{
			name:         "leaa",
			user:         leaa,
			learID:       lear.ID,
			statusAdd:    http.StatusOK,
			statusRemove: http.StatusOK,
			alter:        fmt.Sprintf(`{"position": "leaa", "user": %q}`, user.ID),
		},
		{
			name:         "unauthorized",
			learID:       stores.NewTestUser(t, e.UserStore).ID,
			statusAdd:    http.StatusUnauthorized,
			statusRemove: http.StatusUnauthorized,
			alter:        fmt.Sprintf(`{"position": "leaa", "user": %q}`, user.ID),
		},
		{
			name:         "random role",
			user:         lear,
			learID:       org.Data.Dependencies()[0].ID,
			statusAdd:    http.StatusBadRequest,
			statusRemove: http.StatusBadRequest,
			alter:        fmt.Sprintf(`{"position": "gOd", "user": %q}`, user.ID),
		},
		{
			name:         "remove lear",
			user:         lear,
			learID:       user.ID,
			statusAdd:    http.StatusOK,
			statusRemove: http.StatusUnauthorized,
			alter:        fmt.Sprintf(`{"position": "lear", "user": %q}`, user.ID),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// add role
			w := httptest.NewRecorder()
			router.ServeHTTP(w, loginAs(t, e, c.user,
				httptest.NewRequest("POST", "/organization/"+org.ID.String()+"/roles", bytes.NewBufferString(c.alter))))

			compareRespCode(t, c.statusAdd, w.Code, w.Body.String())

			if c.statusAdd == http.StatusOK {
				doc := models.Document{Data: new(models.Organization)}
				if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
					t.Errorf("decoding failed: %s", err)
				}
				after := len(doc.Data.(*models.Organization).Roles.LEAAs)

				if (before+1 != after && c.statusRemove == http.StatusOK) || doc.Data.(*models.Organization).Roles.LEAR != c.learID {
					t.Errorf("Expected one new member to have been added; got: %v , expected: %v", after, before+1)
				}
			}

			// remove role
			w = httptest.NewRecorder()
			router.ServeHTTP(w, loginAs(t, e, c.user,
				httptest.NewRequest("DELETE", "/organization/"+org.ID.String()+"/roles", bytes.NewBufferString(c.alter))))

			compareRespCode(t, c.statusRemove, w.Code, w.Body.String())

			if c.statusRemove == http.StatusOK {
				doc := models.Document{Data: new(models.Organization)}
				if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
					t.Errorf("decoding failed: %s", err)
				}

				if len(doc.Data.(*models.Organization).Roles.LEAAs) != 2 {
					t.Errorf("Expected to have only 2 remaining LEAAs after delete; got: %v", doc.Data.(*models.Organization).Roles.LEAAs)
				}

			}
		})
	}
}

func testOrganizationDeleteAttachment(t *testing.T) {
	e, cleanup := newTestEnv(t)
	defer cleanup()
	router := New(e)
	w := httptest.NewRecorder()
	lear := stores.NewTestUser(t, e.UserStore)
	leaa := stores.NewTestUser(t, e.UserStore)
	randomGuy := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID, leaa.ID)
	getr := httptest.NewRequest("GET", "/organization/"+org.ID.String(), nil)
	doc := models.Document{Data: &models.Organization{}}
	data := make(url.Values)

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "organization", org.ID, data)
	r = loginAs(t, e, lear, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	// Test LEAR
	delr := httptest.NewRequest("DELETE", "/organization/"+org.ID.String()+"/gg.jpg", nil)
	w = httptest.NewRecorder()
	delr = loginAs(t, e, lear, delr)
	router.ServeHTTP(w, delr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	// Upload test file
	data.Add("upload-type", "general leaflet")
	r = createTestFileRequest(t, "organization", org.ID, data)
	r = loginAs(t, e, lear, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	// Test random guy
	delr = httptest.NewRequest("DELETE", "/organization/"+org.ID.String()+"/gg.jpg", nil)
	w = httptest.NewRecorder()
	delr = loginAs(t, e, randomGuy, delr)
	router.ServeHTTP(w, delr)
	compareRespCode(t, http.StatusUnauthorized, w.Code, w.Body.String())

	// Test LEAA
	delr = httptest.NewRequest("DELETE", "/organization/"+org.ID.String()+"/gg.jpg", nil)
	w = httptest.NewRecorder()
	delr = loginAs(t, e, leaa, delr)
	router.ServeHTTP(w, delr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()
	getr = loginAs(t, e, lear, getr)
	router.ServeHTTP(w, getr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Errorf("decoding failed:%s", err)
	}

	if len(doc.Attachments) > 0 {
		t.Errorf("Expected file to be deleted; attachments: %v", doc.Attachments)
	}
}

func TestExportMeetingsCSV(t *testing.T) {
	e, cleanup := newTestEnv(t)
	defer cleanup()

	router := New(e)
	w := httptest.NewRecorder()

	admin := stores.NewTestAdmin(t, e.UserStore)

	org := stores.NewTestOrg(t, e.OrganizationStore)
	prj := stores.NewTestProject(t, e.ProjectStore)

	r := httptest.NewRequest("GET", "/organization/"+org.ID.String()+"/meetings", nil)

	// some meetings
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithOrg(org.ID))
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithOrg(org.ID))
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithOrg(org.ID))
	stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithOrg(org.ID), stores.TMeetingWithPrj(prj.ID))

	r = loginAs(t, e, admin, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if w.Body.Len() <= 1 {
		t.Errorf("expected csv content but got none")
	}
}
