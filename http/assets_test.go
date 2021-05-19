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
)

const (
	fmtValidNewAsset = `{
		"owner":"%s",
		"address": "End of the world",
		"coordinates": {
			"Lat": 9.83012,
			"Lng": 37.75721
		},
		"area": 9000,
		"valid": 3,
		"cadastre": "666666666666",
		"country": "Latvia",
		"building_type": 4,
		"category": "nonresidential_office_buildings"
	}`
	fmtValidUpdatedAsset = `{
		"owner":"%s",
		"address": "End of the world, floor 2",
		"coordinates": {
			"Lat": 9.83012,
			"Lng": 73.75721
		},
		"area": 9001,
		"valid": 1,
		"cadastre": "666666666666"
	}`
	fmtUpdatedHiddenColumn = `{
		"owner":"%s",
		"address": "End of the world, floor 2",
		"coordinates": {
			"Lat": 9.83012,
			"Lng": 73.75721
		},
		"area": 9001,
		"valid": 1,
		"cadastre": "666666666666",
		"CreatedAt": %q
	}`
	fmtMissingAreaAsset = `{
		"owner":"%s",
		"address": "End of the world"
	}`
	newArea = 9001
)

func TestAsset(t *testing.T) {
	t.Run("create", testAssetCreate)
	t.Run("get", testAssetGet)
	t.Run("update", testAssetUpdate)
	t.Run("update/hiddenColumn", testUpdateHiddenColumn)
	t.Run("list", testAssetList)
	t.Run("list/user", testAssetListByUser)
	t.Run("list/user/empty", testAssetListByUserNoOrganization)
	t.Run("upload", testAssetUpload)
	t.Run("upload/delete", testAssetDeleteAttachment)
	t.Run("upload/escapeASCII", testAssetUploadASCII)
	t.Run("upload/delete/escapeASCII", testAssetDeleteASCII)
}

func testAssetDeleteASCII(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()
	usr := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, usr.ID)
	asset := stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))

	data := make(url.Values)
	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "asset", asset.ID, data, `g g.jpg`)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()

	dr := httptest.NewRequest("DELETE", "/asset/"+asset.ID.String()+`/g%20g.jpg`, nil)
	dr = loginAs(t, e, usr, dr)

	router.ServeHTTP(w, dr)
	compareRespCode(t, http.StatusOK, w.Code, "<binary>")

	w = httptest.NewRecorder()

	getr := httptest.NewRequest("GET", "/asset/"+asset.ID.String()+`/g%20g.jpg`, nil)
	getr = loginAs(t, e, usr, getr)

	router.ServeHTTP(w, getr)
	compareRespCode(t, http.StatusNotFound, w.Code, "<binary")
}

func testAssetUploadASCII(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()
	usr := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, usr.ID)
	asset := stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))

	data := make(url.Values)
	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "asset", asset.ID, data, `g g.jpg`)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()

	getr := httptest.NewRequest("GET", "/asset/"+asset.ID.String()+`/g%20g.jpg`, nil)
	getr = loginAs(t, e, usr, getr)

	router.ServeHTTP(w, getr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

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

func testAssetCreate(t *testing.T) {
	var (
		e, del      = newTestEnv(t)
		router      = New(e)
		lear        = stores.NewTestUser(t, e.UserStore)
		org         = stores.NewTestOrg(t, e.OrganizationStore, lear.ID)
		vAsset      = fmt.Sprintf(fmtValidNewAsset, org.ID)
		invalidUser = stores.NewTestUser(t, e.UserStore)
		tt          = []struct {
			name   string
			body   string
			status int
			user   *models.Document
			valid  models.ValidationStatus
		}{
			{
				name:   "valid but logged out",
				body:   vAsset,
				valid:  models.ValidationStatusValid,
				status: http.StatusUnauthorized,
			},
			{
				name:   "valid",
				body:   vAsset,
				valid:  models.ValidationStatusValid,
				status: http.StatusOK,
				user:   lear,
			},
			{
				name:   "invalid",
				body:   vAsset,
				valid:  models.ValidationStatusRegistered,
				status: http.StatusUnauthorized,
				user:   invalidUser,
			},
			{
				name:   "truncated payload",
				body:   vAsset[:len(vAsset)-1],
				user:   stores.NewTestUser(t, e.UserStore),
				status: http.StatusBadRequest,
				valid:  models.ValidationStatusValid,
			},
			{
				name:   "missing area",
				body:   fmt.Sprintf(fmtMissingAreaAsset, org.ID),
				user:   lear,
				status: http.StatusBadRequest,
				valid:  models.ValidationStatusValid,
			},
			{
				name:   "valid but not org member",
				body:   vAsset,
				status: http.StatusUnauthorized,
				user:   stores.NewTestUser(t, e.UserStore),
				valid:  models.ValidationStatusValid,
			},
		}
	)
	defer del()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				w   = httptest.NewRecorder()
				doc = models.Document{Data: &models.Asset{}}
				r   = httptest.NewRequest("POST", "/asset",
					strings.NewReader(tc.body))
			)

			if tc.user != nil {
				ctx := context.Background()
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

			asset := doc.Data.(*models.Asset)
			if asset.Valid != models.ValidationStatusRegistered {
				t.Errorf("should be 'Registered' by default; got %v", asset.Valid)
			}

			location := w.Header().Get("Location")
			if !strings.HasSuffix(location, "/asset/"+doc.ID.String()) {
				t.Errorf("Expected proper Location header; got %q",
					location)
			}
		})
	}
}

func testAssetGet(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)

		lear = stores.NewTestUser(t, e.UserStore)
		o    = stores.NewTestOrg(t, e.OrganizationStore, lear.ID)
		a    = stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(o.ID))
		r    = httptest.NewRequest("GET", "/asset/"+a.ID.String(), nil)
		w    = httptest.NewRecorder()
		doc  = models.Document{Data: &models.Asset{}}
	)
	defer del()
	r = loginAs(t, e, lear, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	if err := e.AssetStore.Delete(context.Background(), &doc); err != nil {
		t.Fatalf("can't delete: %s", err)
	}

	// Reply same request once the asset is already deleted
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code 404 after deletion; got %d", w.Code)
	}
}

func testAssetUpdate(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)

		lear = stores.NewTestUser(t, e.UserStore)
		org  = stores.NewTestOrg(t, e.OrganizationStore, lear.ID)
		a    = stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))
		w    = httptest.NewRecorder()
		doc  = models.Document{Data: &models.Asset{}}
	)
	defer del()

	router.ServeHTTP(w,
		loginAs(t, e, lear,
			httptest.NewRequest("PUT", "/asset/"+a.ID.String(),
				strings.NewReader(fmt.Sprintf(fmtValidUpdatedAsset, a.Data.Dependencies()[0].ID))),
		),
	)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	asset := doc.Data.(*models.Asset)
	if area := asset.Area; area != newArea {
		t.Fatalf("expected name to be updated; got %d", area)
	}

	if asset.CreatedAt.Year() != time.Now().Year() {
		t.Errorf("CreatedAt has a strange value: %v", asset.CreatedAt)
	}

	if asset.UpdatedAt.Year() != time.Now().Year() {
		t.Errorf("UpdatedAt has a strange value: %v", asset.UpdatedAt)
	}

	// Try again without being logged in.
	w = httptest.NewRecorder()
	router.ServeHTTP(w,
		httptest.NewRequest("PUT", "/asset/"+a.ID.String(),
			strings.NewReader(fmt.Sprintf(fmtValidUpdatedAsset, a.Data.Dependencies()[0].ID))),
	)
	compareRespCode(t, http.StatusUnauthorized, w.Code, w.Body.String())

	// Reply same request once the asset is already deleted
	w = httptest.NewRecorder()
	if err := e.AssetStore.Delete(context.Background(), &doc); err != nil {
		t.Fatalf("can't delete: %s", err)
	}

	router.ServeHTTP(w,
		loginAs(t, e, lear,
			httptest.NewRequest("PUT", "/asset/"+a.ID.String(),
				strings.NewReader(fmt.Sprintf(fmtValidUpdatedAsset, a.Data.Dependencies()[0].ID))),
		),
	)
	compareRespCode(t, http.StatusNotFound, w.Code, w.Body.String())
}

func testUpdateHiddenColumn(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	testTime := time.Date(2000, time.January, 15, 10, 20, 30, 0, time.UTC)
	lear := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, lear.ID)
	a := stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))
	w := httptest.NewRecorder()
	doc := models.Document{Data: &models.Asset{}}

	router.ServeHTTP(w,
		loginAs(t, e, lear,
			httptest.NewRequest("PUT", "/asset/"+a.ID.String(),
				strings.NewReader(fmt.Sprintf(fmtUpdatedHiddenColumn, a.Data.Dependencies()[0].ID, testTime.Format(time.RFC3339)))),
		),
	)

	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}
	createdAt := doc.Data.((*models.Asset)).Value.CreatedAt
	if testTime.Equal(createdAt) {
		t.Fatalf("The user shouldn't be able to change CreatedAt column ; got %v", createdAt)
	}
}

func testAssetList(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	owner := stores.NewTestOrg(t, e.OrganizationStore)
	tcases := []struct {
		name  string
		path  string
		count func() int
	}{
		{
			name: "ok",
			path: "/asset",
			count: func() int {
				stores.NewTestAsset(t, e.AssetStore)
				stores.NewTestAsset(t, e.AssetStore)
				stores.NewTestAsset(t, e.AssetStore)
				stores.NewTestAsset(t, e.AssetStore)
				stores.NewTestAsset(t, e.AssetStore)

				return 5
			},
		},
		{
			name: "search",
			path: "/asset?search=tinTyava",
			count: func() int {
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithAddr("Tintyava 15-17"))
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithAddr("tinTyava 15-17"))

				return 2
			},
		},
		{
			name: "building_type",
			path: fmt.Sprintf("/asset?building_type=%d", models.BuildingType104),
			count: func() int {
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithBuildingType(models.BuildingType104))
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithBuildingType(models.BuildingType104))
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithBuildingType(models.BuildingType104))

				return 3
			},
		},
		{
			name: "status",
			path: fmt.Sprintf("/asset?status=%d", models.ValidationStatusDeclined),
			count: func() int {
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithStatus(models.ValidationStatusDeclined))
				return 1
			},
		},
		{
			name: "owner",
			path: fmt.Sprintf("/asset?owner=%s", owner.ID),
			count: func() int {
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(owner.ID))
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(owner.ID))
				stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(owner.ID))
				return 3
			},
		},
		{
			name: "with_all",
			path: fmt.Sprintf("/asset?search=%s&status=%d&building_type=%d", "tintyava", models.ValidationStatusPending, models.BuildingType103),
			count: func() int {
				stores.NewTestAsset(t, e.AssetStore,
					stores.TAWithStatus(models.ValidationStatusPending),
					stores.TAWithBuildingType(models.BuildingType103),
					stores.TAWithAddr("Tintyava 15-17"))
				stores.NewTestAsset(t, e.AssetStore,
					stores.TAWithStatus(models.ValidationStatusPending),
					stores.TAWithBuildingType(models.BuildingType103),
					stores.TAWithAddr("Tintyava 15-17"))
				stores.NewTestAsset(t, e.AssetStore,
					stores.TAWithStatus(models.ValidationStatusPending),
					stores.TAWithBuildingType(models.BuildingType103),
					stores.TAWithAddr("Tintyava 15-17"))
				stores.NewTestAsset(t, e.AssetStore,
					stores.TAWithStatus(models.ValidationStatusPending),
					stores.TAWithBuildingType(models.BuildingType103),
					stores.TAWithAddr("Tintyava 15-17"))

				return 4
			},
		},
	}

	w := httptest.NewRecorder()
	router := New(e)
	u := stores.NewTestUser(t, e.UserStore)

	for _, c := range tcases {
		t.Run(c.name, func(t *testing.T) {
			var (
				expDocs = c.count()
				docs    dummyList
			)

			router.ServeHTTP(w, loginAs(t, e, u,
				httptest.NewRequest("GET", c.path, nil)))
			compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

			if err := json.NewDecoder(w.Body).Decode(&docs); err != nil {
				t.Errorf("decoding failed: %s", err)
			}

			count, err := strconv.Atoi(w.Header().Get(countHeader))
			if err != nil {
				t.Errorf("%q header is not set: %s", countHeader, err)
			}

			if len(docs.Documents) != count || count != expDocs {
				t.Errorf("%q header reports %d records, got %d, but expects %d",
					countHeader, count, len(docs.Documents), expDocs)
			}
		})
	}
}

func testAssetListByUser(t *testing.T) {
	var (
		docs   dummyList
		w      = httptest.NewRecorder()
		e, del = newTestEnv(t)
		router = New(e)
	)
	defer del()

	// createDeps creates couple assets for the purpose of testing listByMember
	//
	//       user       user
	//        /\         |
	//      org org     org
	//      |    |       |
	//      as   as      as

	user := stores.NewTestUser(t, e.UserStore)
	o1 := stores.NewTestOrg(t, e.OrganizationStore, user.ID)
	o2 := stores.NewTestOrg(t, e.OrganizationStore, user.ID)

	stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(o1.ID))
	stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(o2.ID))

	stores.NewTestAsset(t, e.AssetStore)

	router.ServeHTTP(w, loginAs(t, e, user,
		httptest.NewRequest("GET", fmt.Sprintf("/user/%s/assets", user.ID), nil)))
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&docs); err != nil {
		t.Errorf("decoding failed %s", err)
	}

	if len(docs.Documents) != 2 {
		t.Errorf("Get %d assets; expected: 2", len(docs.Documents))
	}
	count, err := strconv.Atoi(w.Header().Get(countHeader))
	if err != nil {
		t.Errorf("%q header is not set: %s", countHeader, err)
	}

	if len(docs.Documents) != count {
		t.Errorf("%q header reports %d records, got %d",
			countHeader, count, len(docs.Documents))
	}
}

func testAssetListByUserNoOrganization(t *testing.T) {
	var (
		docs   dummyList
		w      = httptest.NewRecorder()
		e, del = newTestEnv(t)
		router = New(e)
	)
	defer del()

	user := stores.NewTestUser(t, e.UserStore)
	stores.NewTestAsset(t, e.AssetStore)
	stores.NewTestAsset(t, e.AssetStore)
	stores.NewTestAsset(t, e.AssetStore)

	router.ServeHTTP(w, loginAs(t, e, user,
		httptest.NewRequest("GET", fmt.Sprintf("/user/%s/assets", user.ID), nil)))
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&docs); err != nil {
		t.Errorf("decoding failed %s", err)
	}

	if len(docs.Documents) != 0 {
		t.Errorf("Got %d assets; expected: 0", len(docs.Documents))
	}
	count, err := strconv.Atoi(w.Header().Get(countHeader))
	if err != nil {
		t.Errorf("%q header is not set: %s", countHeader, err)
	}

	if len(docs.Documents) != count {
		t.Errorf("%q header reports %d records, got %d",
			countHeader, count, len(docs.Documents))
	}
}

func TestNotAllowedAssetMethods(t *testing.T) {
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
				httptest.NewRequest(method, "/asset", nil),
			)
			compareRespCode(t, http.StatusMethodNotAllowed,
				w.Code, w.Body.String(),
			)
		})
	}
}

func testAssetUpload(t *testing.T) {
	var (
		e, del = newTestEnv(t)
		router = New(e)
		w      = httptest.NewRecorder()
		usr    = stores.NewTestUser(t, e.UserStore)
		org    = stores.NewTestOrg(t, e.OrganizationStore, usr.ID)
		asset  = stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))
		getr   = httptest.NewRequest("GET", "/asset/"+asset.ID.String(), nil)
		doc    = models.Document{Data: &models.Asset{}}
		data   = make(url.Values)
	)
	defer del()

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "asset", asset.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()

	// create another file to test same name collision and renaming
	r = createTestFileRequest(t, "asset", asset.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()

	getr = loginAs(t, e, usr, getr)

	router.ServeHTTP(w, getr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Errorf("decoding failed: %s", err)
	}

	if len(doc.Attachments) != 2 {
		t.Errorf("expected 2 attachments in document; got %d; \n attachments map: %v",
			len(doc.Attachments), doc.Attachments)
	}
}

func testAssetDeleteAttachment(t *testing.T) {
	var (
		e, cleanup = newTestEnv(t)
		router     = New(e)
		w          = httptest.NewRecorder()
		doc        = models.Document{Data: &models.Asset{}}
		usr        = stores.NewTestUser(t, e.UserStore)
		org        = stores.NewTestOrg(t, e.OrganizationStore, usr.ID)
		asset      = stores.NewTestAsset(t, e.AssetStore, stores.TAWithOrg(org.ID))
		getr       = httptest.NewRequest("GET", "/asset/"+asset.ID.String(), nil)
		data       = make(url.Values)
	)
	defer cleanup()

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "asset", asset.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	delr := httptest.NewRequest("DELETE", "/asset/"+asset.ID.String()+"/gg.jpg", nil)
	delr = loginAs(t, e, usr, delr)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, delr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()

	getr = loginAs(t, e, usr, getr)
	router.ServeHTTP(w, getr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Errorf("decoding failed:%s", err)
	}

	if len(doc.Attachments) > 0 {
		t.Errorf("Expected file to be deleted; attachments: %v", doc.Attachments)
	}
}
