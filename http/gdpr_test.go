package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/uuid"
	"stageai.tech/sunshine/sunshine/graphql"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestGDPRUpload(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	gd := stores.NewTestGDPRRequest(t, e.GDPRStore.DB(), models.GDPRTypeGet)

	cases := []struct {
		name   string
		gdID   uuid.UUID
		status int
		count  int
	}{
		{
			name:   "ok",
			gdID:   gd.ID,
			status: http.StatusOK,
			count:  1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			router := New(e)
			// upload
			r := loginAs(t, e, nil,
				createTestFileRequest(t, "gdpr", c.gdID, nil))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)
			compareRespCode(t, c.status, w.Code, w.Body.String())

			gdoc, err := e.GDPRStore.Get(context.Background(), c.gdID)
			gdoc = tNilOrNotFound(t, err, gdoc, c.status)
			if len(gdoc.Attachments) != c.count {
				t.Errorf("Expected %d attachments got: %v", c.count, gdoc.Attachments)
			}

			// on success we expect Location header with URL pointing to the uploaded file.
			location := w.Header().Get("Location")
			expLocation := path.Join("gdpr", c.gdID.String(), "gg.jpg")
			if w.Code == http.StatusOK && !strings.HasSuffix(location, expLocation) {
				t.Errorf("Expected Location header from response, got: %q", location)
			}

			// head to make sure we can access the file even without downloading it
			r = loginAs(t, e, nil,
				httptest.NewRequest("HEAD", "/gdpr/"+c.gdID.String()+"/gg.jpg", nil))
			w = httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, c.status, w.Code, w.Body.String())

			// delete
			r = loginAs(t, e, nil,
				httptest.NewRequest("DELETE", "/gdpr/"+c.gdID.String()+"/gg.jpg", nil))
			w = httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, c.status, w.Code, w.Body.String())

			gdoc, err = e.GDPRStore.Get(context.Background(), c.gdID)
			gdoc = tNilOrNotFound(t, err, gdoc, c.status)
			if len(gdoc.Attachments) != 0 {
				t.Errorf("Expected 0 attachments after delete got: %v", gdoc.Attachments)
			}
		})
	}
}

func TestGDPRUploads(t *testing.T) {
	e := services.NewTestEnv(t)

	router := New(e)
	w := httptest.NewRecorder()

	o := graphql.LoadGQLTestFile(t, "mutation_gdpr_upload_files.json")
	m := `{ "0": ["variables.gdpr.files.0"], "1": ["variables.gdpr.files.1"] }`
	files := []string{"gg.jpg", "g g.jpg"}

	r := createUploadRequest(t, o, m, files)

	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected success but got status code: %v, body %+v", w.Code, w.Body.String())
	}
}

func createUploadRequest(t *testing.T, operations, mapData string, files []string) *http.Request {
	t.Helper()

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	if err := bodyWriter.WriteField("operations", operations); err != nil {
		t.Fatalf("write field operations failed: %v", err)
	}

	if err := bodyWriter.WriteField("map", mapData); err != nil {
		t.Fatalf("write field map failed: %v", err)
	}

	for i := range files {
		f, err := os.Open("testdata/" + files[i])
		if err != nil {
			t.Fatalf("fail to open file: %v", err)
		}

		part, err := bodyWriter.CreateFormFile(fmt.Sprintf("%d", i), files[i])
		if err != nil {
			t.Fatalf("writer.CreateFormFile: %s", err)
		}

		if _, err := io.Copy(part, f); err != nil {
			t.Fatalf("io.Copy: %s", err)
		}

		f.Close()
	}

	if err := bodyWriter.Close(); err != nil {
		t.Fatalf("writer close: %s", err)
	}

	req := httptest.NewRequest("POST", "/query", bodyBuf)
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	return req
}
