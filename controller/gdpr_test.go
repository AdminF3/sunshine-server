package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var emptyCtx = context.Background()

func TestGDPRRequests(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	gdpr := NewGDPR(e)
	u := NewUser(e)
	su := stores.NewTestAdmin(t, u.st.FromKind("user"))
	ctx := services.NewTestContext(t, e, su)
	gReq := &models.GDPRRequest{

		RequesterName:    "Ivan",
		RequesterPhone:   "0088112233",
		RequesterEmail:   "i.ivaonv@test.com",
		RequesterAddress: "Test address",
		Name:             "Petar",
		Phone:            "0088112234",
		Email:            "p.petrov@test.com",
		Address:          "test address peter",
		Action:           models.GDPRTypeGet,
		Reason:           "da",
		Information:      "da",
	}

	dReq := &models.GDPRRequest{

		RequesterName:    "Emil",
		RequesterPhone:   "0088112233",
		RequesterEmail:   "i.ivaonv@test.com",
		RequesterAddress: "Test address",
		Name:             "Petar",
		Phone:            "0088112234",
		Email:            "p.petrov@test.com",
		Address:          "test address peter",
		Action:           models.GDPRTypeDelete,
		Reason:           "da",
		Information:      "da",
	}

	if err := gdpr.SendRequest(emptyCtx, gReq, nil); err != nil {
		t.Errorf("failed to create a GDPR get request: %v", err)
	}
	if err := gdpr.SendRequest(ctx, dReq, nil); err != nil {
		t.Errorf("failed to create a GDPR delete request: %v", err)
	}

	reqs, total, err := gdpr.List(ctx, 0, 0)
	if err != nil {
		t.Fatalf("could not list requests; %s", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 requests; got: %d", len(reqs))
	}

	for _, req := range reqs {
		r, err := gdpr.Get(ctx, req.ID)
		if err != nil {
			t.Errorf("could not get gdpr request; got: %s", err)
		}
		if r.ID == uuid.Nil {
			t.Errorf("got a non-valid gdpr request; got: %v", r)
		}
	}
}

func TestSingleFileUpload(t *testing.T) {
	e := services.NewTestEnv(t)

	cases := []struct {
		name         string
		path         string
		fname        string
		escapedfname string
	}{
		{
			name:         "ok",
			path:         filepath.Join("..", "http", "testdata", "gg.jpg"),
			fname:        "gg.jpg",
			escapedfname: "gg.jpg",
		},
		{
			name:         "with space",
			path:         filepath.Join("..", "http", "testdata", "g g.jpg"),
			fname:        "g g.jpg",
			escapedfname: "g%20g.jpg",
		},
		{
			name:         "with hashtag",
			path:         filepath.Join("..", "http", "testdata", "g#g.jpg"),
			fname:        "g#g.jpg",
			escapedfname: "g%23g.jpg",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cleanup := func(db *gorm.DB) {
				db.Exec("TRUNCATE attachments")
			}
			defer cleanup(e.DB)

			f, err := os.Open(c.path)
			if err != nil {
				t.Fatalf("fail to open file: %v", err)
			}

			u := Upload{
				File:        f,
				Filename:    c.fname,
				Size:        134,
				ContentType: "image/jpeg",
			}

			if err := uploadGQLfile(e.GDPRStore, u, uuid.New(), e.Paths.Uploads); err != nil {
				t.Fatalf("fail to upload a file: %v", err)
			}

			var att []models.Attachment
			e.DB.Model(&models.Attachment{}).Find(&att)

			if len(att) != 1 {
				t.Fatalf("fail to find attach")
			}

			if att[0].Name != c.escapedfname {
				t.Fatalf("expected filename: %s, got: %s", c.fname, att[0].Name)
			}
		})
	}
}
