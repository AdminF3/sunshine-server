package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/google/uuid"
)

func compareRespCode(t *testing.T, expStatus, gotStatus int, body string) {
	t.Helper()
	if expStatus != gotStatus {
		t.Errorf("expected status code %d; got %d: %s",
			expStatus, gotStatus, body)
	}
}

func createTestFileRequest(t *testing.T, kind string, uuid uuid.UUID, data url.Values) *http.Request {
	return createTestFilenameRequest(t, kind, uuid, data, "gg.jpg")
}

func createTestFilenameRequest(t *testing.T, kind string, uuid uuid.UUID, data url.Values, filename string) *http.Request {
	var (
		body   bytes.Buffer
		f, err = os.Open("testdata/" + filename)
	)
	defer f.Close()

	if err != nil {
		t.Fatalf("Can't open %s: %s", "testdata/"+filename, err)
	}

	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("writer.CreateFormFile: %s", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		t.Fatalf("io.Copy: %s", err)
	}
	for field := range data {
		if err := writer.WriteField(field, data.Get(field)); err != nil {
			t.Fatalf("writer %q field: %s", field, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("writer close: %s", err)
	}

	r := httptest.NewRequest("POST", "/"+kind+"/"+uuid.String()+"/upload", &body)
	r.Header = http.Header{"Content-Type": {
		fmt.Sprintf(`multipart/form-data; boundary=%s`, writer.Boundary())}}

	return r
}

func testSpecialUpload(e *services.Env, t *testing.T, entityKind, kind, real string, user *models.Document, doc models.Document) {
	var (
		router   = New(e)
		expected string
	)

	if kind != "" {
		expected = fileURL(entityKind, "gg.jpg", doc.ID)
	}
	if real != expected && kind != "" {
		t.Errorf("expected %q to be %q got: %q", kind, expected, real)
	}

	if real != "" {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", real, nil)
		r = loginAs(t, e, user, r)
		router.ServeHTTP(w, r)
		compareRespCode(t, http.StatusOK, httptest.NewRecorder().Code, "<binary>")

		r = httptest.NewRequest("HEAD", real, nil)
		r = loginAs(t, e, user, r)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		compareRespCode(t, http.StatusOK, httptest.NewRecorder().Code, "")
	}
}

func compareFile(t *testing.T, e *services.Env, r http.Handler, file, url string, user *models.Document) {
	t.Helper()
	w := httptest.NewRecorder()
	filer := httptest.NewRequest("GET", url, nil)
	if user != nil {
		filer = loginAs(t, e, user, filer)
	}
	r.ServeHTTP(w, filer)
	compareRespCode(t, http.StatusOK, w.Code, "<binary>")
	f, err := ioutil.ReadFile(file)
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

type dummyDocument struct {
	ID          uuid.UUID                    `json:"_id" validate:"required" gorm:"type:uuid; primary_key"`
	Deleted     bool                         `json:"_deleted,omitempty" gorm:"-"`
	Kind        string                       `json:"type" validate:"required" gorm:"-"`
	Timestamp   time.Time                    `json:"timestamp" validate:"required" gorm:"-"`
	Attachments map[string]models.Attachment `json:"_attachments,omitempty" gorm:"-"`

	// Data can now be unmarshalled from any models.Entity value.
	Data map[string]interface{} `json:"data"`
}

type dummyList struct {
	Documents    []dummyDocument          `json:"documents"`
	Dependencies map[string]dummyDocument `json:"dependencies"`
	Errors       error                    `json:"errors"`
}
