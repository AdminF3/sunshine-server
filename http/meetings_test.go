package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

func TestMeetingsUpload(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	usr := stores.NewTestUser(t, e.UserStore)
	org := stores.NewTestOrg(t, e.OrganizationStore, usr.ID)
	otherUsr := stores.NewTestUser(t, e.UserStore)

	cases := []struct {
		name   string
		mtng   uuid.UUID
		user   *models.Document
		status int
		count  int
	}{
		{
			name:   "ok",
			user:   usr,
			status: http.StatusOK,
			count:  1,
		},
		{
			name:   "not logged-in",
			status: http.StatusUnauthorized,
			count:  0,
		},
		{
			name:   "unauthorized",
			user:   otherUsr,
			status: http.StatusUnauthorized,
			count:  0,
		},
		{
			name:   "random user",
			user:   stores.NewTestUser(t, e.UserStore),
			status: http.StatusUnauthorized,
			count:  0,
		},
		{
			name:   "not found",
			mtng:   uuid.New(),
			user:   usr,
			status: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mtng := stores.NewTestMeeting(t, e.MeetingsStore, stores.TMeetingWithOrg(org.ID))
			mtngID := mtng.ID
			if c.mtng != uuid.Nil {
				mtngID = c.mtng
			}
			router := New(e)

			// upload
			r := loginAs(t, e, c.user,
				createTestFileRequest(t, "meeting", mtngID, nil))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, c.status, w.Code, w.Body.String())

			mdoc, err := e.MeetingsStore.Get(context.Background(), mtngID)
			mdoc = tNilOrNotFound(t, err, mdoc, c.status)
			if len(mdoc.Attachments) != c.count {
				t.Errorf("Expected %d attachments got: %v", c.count, mdoc.Attachments)
			}

			// on success we expect Location header wth URL pointing to the uploaded file.
			location := w.Header().Get("Location")
			expLocation := path.Join("meeting", mtngID.String(), "gg.jpg")
			if w.Code == http.StatusOK && !strings.HasSuffix(location, expLocation) {
				t.Errorf("Expected Location header from response, got: %q", location)
			}

			// head to make sure we can access the file even without downloading it
			r = loginAs(t, e, c.user,
				httptest.NewRequest("HEAD", "/meeting/"+mtngID.String()+"/gg.jpg", nil))
			w = httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, c.status, w.Code, w.Body.String())

			// delete
			r = loginAs(t, e, c.user,
				httptest.NewRequest("DELETE", "/meeting/"+mtngID.String()+"/gg.jpg", nil))
			w = httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, c.status, w.Code, w.Body.String())

			mdoc, err = e.MeetingsStore.Get(context.Background(), mtngID)
			mdoc = tNilOrNotFound(t, err, mdoc, c.status)
			if len(mdoc.Attachments) != 0 {
				t.Errorf("Expected 0 attachments after delete got: %v", mdoc.Attachments)
			}
		})
	}
}

func tNilOrNotFound(t *testing.T, err error, doc *models.Document, status int) *models.Document {
	if err == nil {
		return doc
	}

	if errors.Is(err, gorm.ErrRecordNotFound) && status != http.StatusOK {
		return &models.Document{Data: new(models.Meeting)}
	}

	t.Fatal(err)
	return doc
}
