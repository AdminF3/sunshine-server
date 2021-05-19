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
	"github.com/jinzhu/gorm"
)

func TestUser(t *testing.T) {
	t.Run("create", testUserCreate)
	t.Run("get", testUserGet)
	t.Run("update", testUserUpdate)
	t.Run("update/hiddenColumn", testUserUpdateHiddenColumns)
	t.Run("list", testUserList)
	t.Run("upload", testUserUpload)
	t.Run("upload/delete", testUserDeleteAttachment)
	t.Run("upload/delete/escapeASCII", testUserDeleteEscapeSymbols)
	t.Run("confirm", testUserConfirm)
	t.Run("upload/escapeASCII", testUserUploadEscapeSymbols)
}

func testUserCreate(t *testing.T) {
	e, del := newTestEnv(t)
	router := New(e)
	defer del()

	tt := []struct {
		name   string
		body   string
		status int
	}{
		{
			name: "normal",
			body: `{
				"name": "John Doe",
				"email": "john_doe@example.com",
				"password": "foo",
				"superuser": false,
				"country": "Latvia"
			}`,
			status: http.StatusOK,
		},
		{
			name: "duplicate_email",
			body: `{
				"name": "John Doe",
				"email": "john_doe@example.com",
				"password": "foo",
				"superuser": false,
				"country": "Latvia"
			}`,
			status: http.StatusBadRequest,
		},
		{
			name: "superuser",
			body: `{
				"name": "John Doe",
				"email": "super_doe@example.com",
				"password": "foo",
				"superuser": true,
				"country": "Latvia"
			}`,
			status: http.StatusOK,
		},
		{
			name:   "empty body",
			body:   ``,
			status: http.StatusBadRequest,
		},
		{
			name: "malformed json",
			body: `{
				"name": "John Doe",
				"email": "john_doe@example.com",`,
			status: http.StatusBadRequest,
		},
		{
			name: "missing name",
			body: `{
				"name": "",
				"email": "john_doe@example.com",
				"password": "foo"
			}`,
			status: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			body: `{
				"name": "John Doe",
				"email": "john_doe_at_example.com",
				"password": "foo"
			}`,
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/user",
				strings.NewReader(tc.body))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())

			if w.Code != http.StatusOK {
				return
			}

			var doc = models.Document{Data: &models.User{}}

			err := json.NewDecoder(w.Body).Decode(&doc)
			if err != nil {
				t.Fatalf("can't decode success response: %s", err)
			}

			if password := doc.Data.(*models.User).Password; password != "" {
				t.Errorf("password hash shouldn't be sent; got %q", password)
			}

			if doc.Data.(*models.User).SuperUser {
				t.Errorf("user can not be superuser upon signup")
			}

			location := w.Header().Get("Location")
			if !strings.HasSuffix(location, "/user/"+doc.ID.String()) {
				t.Errorf("Expected proper Location header; got %q",
					location)
			}
		})
	}
}

func testUserGet(t *testing.T) {
	e, del := newTestEnv(t)
	router := New(e)
	defer del()
	d := stores.NewTestUser(t, e.UserStore)

	r := httptest.NewRequest("GET", "/user/"+d.ID.String(), nil)
	r = loginAs(t, e, d, r)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	var doc = models.Document{Data: &models.User{}}
	err := json.NewDecoder(w.Body).Decode(&doc)
	if err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	if password := doc.Data.(*models.User).Password; password != "" {
		t.Errorf("password hash shouldn't be sent; got %q", password)
	}

	if err = e.UserStore.Delete(context.Background(), &doc); err != nil {
		t.Fatalf("can't delete: %s", err)
	}

	// Reply same request once the user is already deleted
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code 404 after deletion; got %d", w.Code)
	}
}

func testUserUpdate(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()
	router := New(e)

	d := stores.NewTestUser(t, e.UserStore)
	updateBody := strings.NewReader(fmt.Sprintf(
		`{"name":"John Doe II","email":"%s", "superuser": true, "valid":3}`,
		d.Data.(*models.User).Email))

	user := d.Data.(*models.User)
	updateReq := loginWith(t, router, user.Email, "foo",
		httptest.NewRequest("PUT", "/user/"+d.ID.String(), updateBody))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, updateReq)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	doc := models.Document{Data: &models.User{}}
	err := json.NewDecoder(w.Body).Decode(&doc)
	if err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	user = doc.Data.(*models.User)
	if newName := user.Name; newName != "John Doe II" {
		t.Fatalf("expected name to be updated; got %s", newName)
	}

	if password := user.Password; password != "" {
		t.Errorf("password hash shouldn't be sent; got %q", password)
	}

	if supusr := user.SuperUser; supusr {
		t.Errorf("users should not be able to elevate their admin rights; got %t", supusr)
	}

	if validUser := user.Valid; validUser == models.ValidationStatusValid {
		t.Errorf("Users should not be able to change their valid rights; got %d", validUser)
	}

	if user.CreatedAt.Year() != time.Now().Year() {
		t.Errorf("CreatedAt has a strange value: %v", user.CreatedAt)
	}

	if user.UpdatedAt.Year() != time.Now().Year() {
		t.Errorf("UpdatedAt has a strange value: %v", user.UpdatedAt)
	}

	t.Run("unauthorized", func(t *testing.T) {
		outUpdateReq := httptest.NewRequest("PUT", "/user/"+d.ID.String(), updateBody)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, outUpdateReq)
		compareRespCode(t, http.StatusUnauthorized, w.Code, w.Body.String())
	})

	t.Run("deleted", func(t *testing.T) {
		w = httptest.NewRecorder()
		deletedReq := loginWith(t, router, user.Email, "foo",
			httptest.NewRequest("PUT", "/user/"+d.ID.String(), updateBody))

		if err = e.UserStore.Delete(context.Background(), &doc); err != nil {
			t.Fatalf("can't delete: %s", err)
		}

		router.ServeHTTP(w, deletedReq)
		compareRespCode(t, http.StatusNotFound, w.Code, w.Body.String())
	})
}

func testUserUpdateHiddenColumns(t *testing.T) {
	e, del := newTestEnv(t)
	router := New(e)
	defer del()
	d := stores.NewTestUser(t, e.UserStore)
	testTime := time.Date(2000, time.January, 15, 10, 20, 30, 0, time.UTC)
	updated := fmt.Sprintf(
		`{"name":"John Doe II","email":"%s", "password": "foo", "superuser": true, "valid":3,"CreatedAt": %q}`,
		d.Data.(*models.User).Email, testTime.Format(time.RFC3339),
	)

	updateReq := httptest.NewRequest("PUT", "/user/"+d.ID.String(), strings.NewReader(updated))
	updateReq = loginAs(t, e, d, updateReq)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, updateReq)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	var doc = models.Document{Data: &models.User{}}

	err := json.NewDecoder(w.Body).Decode(&doc)
	if err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	user := doc.Data.(*models.User)
	if newName := user.Name; newName != "John Doe II" {
		t.Fatalf("expected name to be updated; got %s", newName)
	}
	createdAt := user.Value.CreatedAt
	if createdAt.UTC().Equal(testTime) {
		t.Fatalf("The user shouldn't be able to change created_At column ; got %v", createdAt)
	}

}

func testUserList(t *testing.T) {
	e, del := newTestEnv(t)
	router := New(e)
	defer del()

	u := stores.NewTestUser(t, e.UserStore)
	usrs, pdos := listu(e.DB)

	cases := []struct {
		name  string
		url   string
		count int
	}{
		{
			name:  "default",
			url:   "/user",
			count: usrs,
		},
		{
			name:  "filter",
			url:   "/user?country_roles=data_protection_officer",
			count: pdos,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var docs dummyList

			r := httptest.NewRequest("GET", c.url, nil)
			r = loginAs(t, e, u, r)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

			if err := json.NewDecoder(w.Body).Decode(&docs); err != nil {
				t.Errorf("decoding failed: %s", err)
			}

			count, err := strconv.Atoi(w.Header().Get(countHeader))
			if err != nil {
				t.Errorf("%q header is not set: %s", countHeader, err)
			}

			if len(docs.Documents) != c.count || count != c.count {
				t.Errorf("%q header reports %d records, got %d, but expected %d",
					countHeader, count, len(docs.Documents), c.count)
			}

			for i, doc := range docs.Documents {
				// dummyList.[]Documents.Data is map[string]interface{}.
				// If the value is missing (as we hope) the type assert will
				// fail and ok will be false.
				if password, ok := doc.Data["password"].(string); ok && password != "" {
					t.Errorf("password hash shouldn't be sent; got %q on %d",
						password, i)
				}
			}
		})

	}
}

func listu(db *gorm.DB) (users, dpos int) {
	db.Table("users").Count(&users)
	db.Model(&models.User{}).
		Joins("inner join country_roles on users.id = country_roles.user_id::UUID").
		Where("country_roles.role = ?", "data_protection_officer").
		Count(&dpos)
	return
}

func testUserUpload(t *testing.T) {
	var tt = []struct {
		name  string
		kind  string
		utype string
		code  int
	}{
		{
			name:  "bad",
			kind:  "",
			utype: "Nan",
			code:  http.StatusBadRequest,
		},
		{
			name:  "no-kind",
			kind:  "",
			utype: "general leaflet",
			code:  http.StatusOK,
		},
		{
			name:  "avatar",
			kind:  "avatar",
			utype: "general leaflet",
			code:  http.StatusOK,
		},
		{
			name:  "identity",
			kind:  "identity",
			utype: "general leaflet",
			code:  http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				e, del = newTestEnv(t)
				router = New(e)
				w      = httptest.NewRecorder()
				d      = stores.NewTestUser(t, e.UserStore)
				getr   = httptest.NewRequest("GET", "/user/"+d.ID.String(), nil)
				doc    = models.Document{Data: &models.User{}}
				data   = make(url.Values)
			)
			defer del()

			if tc.kind != "bad" {
				data.Add("kind", tc.kind)
				data.Add("upload-type", tc.utype)
			}

			r := createTestFileRequest(t, "user", d.ID, data)
			r = loginAs(t, e, d, r)

			router.ServeHTTP(w, r)
			compareRespCode(t, tc.code, w.Code, w.Body.String())

			if tc.code != http.StatusOK {
				return
			}

			w = httptest.NewRecorder()
			getr = loginAs(t, e, d, getr)
			router.ServeHTTP(w, getr)
			compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

			if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
				t.Fatalf("decoding failed: %s", err)
			}

			if tc.name == "avatar" {
				avatar := doc.Data.(*models.User).Avatar
				testSpecialUpload(e, t, "user", tc.name, avatar, d, doc)

				t.Run("file", func(t *testing.T) {
					compareFile(t, e, router, "testdata/gg.jpg", avatar, nil)
				})
			} else if tc.name == "identity" {
				id := doc.Data.(*models.User).Identity
				testSpecialUpload(e, t, "user", tc.name, id, d, doc)
				t.Run("file", func(t *testing.T) {
					compareFile(t, e, router, "testdata/gg.jpg", id, d)
				})
			}

			if len(doc.Attachments) < 1 {
				t.Errorf("no attachments in document")
			}

		})
	}
}

func testUserDeleteAttachment(t *testing.T) {
	var (
		e, cleanup = newTestEnv(t)
		router     = New(e)
		w          = httptest.NewRecorder()
		user       = stores.NewTestUser(t, e.UserStore)
		getr       = httptest.NewRequest("GET", "/user/"+user.ID.String(), nil)
		doc        = models.Document{Data: &models.User{}}
		data       = make(url.Values)
	)
	defer cleanup()

	data.Add("kind", "avatar")
	data.Add("upload-type", "general leaflet")

	r := createTestFileRequest(t, "user", user.ID, data)
	r = loginAs(t, e, user, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	delr := httptest.NewRequest("DELETE", "/user/"+user.ID.String()+"/gg.jpg", nil)

	w = httptest.NewRecorder()
	delr = loginAs(t, e, user, delr)
	router.ServeHTTP(w, delr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()
	getr = loginAs(t, e, user, getr)
	router.ServeHTTP(w, getr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Errorf("decoding failed:%s", err)
	}

	if len(doc.Attachments) > 0 || doc.Data.(*models.User).Avatar != "" {
		t.Errorf("Expected file to be deleted; attachments: %v, avatar: %s",
			doc.Attachments, doc.Data.(*models.User).Avatar)
	}
}

func testUserConfirm(t *testing.T) {
	e, cleanup := newTestEnv(t)
	defer cleanup()
	var (
		router = New(e)
		w      = httptest.NewRecorder()

		// is_active should be ignored
		cjson = `
{
	"name": "ivan",
	"email": "ivan@Dimitrov.com",
	"password": "1111",
	"is_active": true,
	"country": "Bulgaria"
}`

		ljson = `
{
	"email":"ivan@Dimitrov.com",
	"password":"1111"
}`
	)

	// create user, should be inactive
	creater := httptest.NewRequest("POST", "/user", strings.NewReader(cjson))
	router.ServeHTTP(w, creater)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	u := isactive(t, e.UserStore, false)

	// login should fail
	w = httptest.NewRecorder()
	lr := httptest.NewRequest("POST", "/auth/login", strings.NewReader(ljson))
	router.ServeHTTP(w, lr)
	compareRespCode(t, http.StatusUnauthorized, w.Code, w.Body.String())

	// confirm user via email, mock the token
	token := newTestToken(t, e.UserStore, u.ID)
	w = httptest.NewRecorder()
	cr := httptest.NewRequest("POST", "/confirm_user/"+token, nil)
	router.ServeHTTP(w, cr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	isactive(t, e.UserStore, true)

	// again login; this time should succeed.
	w = httptest.NewRecorder()
	lr = httptest.NewRequest("POST", "/auth/login", strings.NewReader(ljson))
	router.ServeHTTP(w, lr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
}

func testUserUploadEscapeSymbols(t *testing.T) {
	e, cleanup := newTestEnv(t)
	router := New(e)
	w := httptest.NewRecorder()
	user := stores.NewTestUser(t, e.UserStore)
	data := make(url.Values)
	defer cleanup()

	data.Add("kind", "avatar")
	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "user", user.ID, data, `g g.jpg`)
	r = loginAs(t, e, user, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	gr := httptest.NewRequest("GET", "/user/"+user.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	gr = loginAs(t, e, user, gr)
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

func testUserDeleteEscapeSymbols(t *testing.T) {
	e, cleanup := newTestEnv(t)
	router := New(e)
	w := httptest.NewRecorder()
	user := stores.NewTestUser(t, e.UserStore)
	data := make(url.Values)
	defer cleanup()

	data.Add("kind", "avatar")
	data.Add("upload-type", "general leaflet")

	r := createTestFilenameRequest(t, "user", user.ID, data, `g g.jpg`)
	r = loginAs(t, e, user, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	dr := httptest.NewRequest("DELETE", "/user/"+user.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	dr = loginAs(t, e, user, dr)
	router.ServeHTTP(w, dr)
	compareRespCode(t, http.StatusOK, w.Code, "<binary>")

	gr := httptest.NewRequest("GET", "/user/"+user.ID.String()+`/g%20g.jpg`, nil)

	w = httptest.NewRecorder()
	gr = loginAs(t, e, user, gr)
	router.ServeHTTP(w, gr)
	compareRespCode(t, http.StatusNotFound, w.Code, "<binary>")
}

func isactive(t *testing.T, st stores.Store, active bool) *models.User {
	udoc, err := st.GetByIndex(context.Background(), "ivan@Dimitrov.com")
	if err != nil {
		t.Errorf("getting user failed:%s", err)
	}

	user := udoc.Data.(*models.User)
	if user.IsActive != active {
		t.Fatalf("newly created user should be %v but is %v", user.IsActive, active)
	}

	return user
}

func newTestToken(t *testing.T, store stores.Store, user uuid.UUID) string {
	var token models.Token
	err := store.DB().Where("user_id = ? AND purpose = ?", user, models.CreateToken).First(&token).Error
	if err != nil {
		t.Fatalf("get token error: %s", err)
	}

	return token.ID.String()
}
