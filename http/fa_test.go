package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestFA(t *testing.T) {
	t.Run("uploadFA", testUploadFAFile)
	t.Run("deleteFA", testDeleteFAFile)
}

func testUploadFAFile(t *testing.T) {
	cases := []struct {
		name  string
		utype string
		code  int
	}{
		{
			name:  "not_existing",
			utype: "not_existing",
			code:  http.StatusBadRequest,
		},
		{
			name:  "ok",
			utype: "fa bank confirmation",
			code:  http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, del := newTestEnv(t)
			defer del()

			router := New(e)
			w := httptest.NewRecorder()
			pm := stores.NewTestUser(t, e.UserStore)
			p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))
			fa := stores.NewTestFA(t, e.FAStore, stores.TFAWithProject(p.ID))
			data := make(url.Values)

			if tc.utype != "" {
				data.Add("upload-type", tc.utype)
			}

			r := createTestFileRequest(t, "forfaitinga", fa.ID, data)
			r = loginAs(t, e, pm, r)

			router.ServeHTTP(w, r)
			compareRespCode(t, tc.code, w.Code, w.Body.String())

			if tc.code != http.StatusOK {
				return
			}

			res, err := e.FAStore.Get(ctx, fa.ID)
			if err != nil {
				t.Fatalf("fail to fetch fa: %v", err)
			}

			if len(res.Attachments) < 1 {
				t.Errorf("no attachments in document")
			}
		})
	}
}

func testDeleteFAFile(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()
	usr := stores.NewTestAdmin(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore)
	fa := stores.NewTestFA(t, e.FAStore, stores.TFAWithProject(prj.ID))
	doc := models.Document{Data: &models.ForfaitingApplication{}}
	data := make(url.Values)

	data.Add("upload-type", "fa bank confirmation")
	r := createTestFileRequest(t, "forfaitinga", fa.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	delr := httptest.NewRequest("DELETE", "/forfaitinga/"+fa.ID.String()+"/gg.jpg", nil)
	delr = loginAs(t, e, usr, delr)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, delr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	res, err := e.FAStore.Get(ctx, fa.ID)
	if err != nil {
		t.Fatalf("fail to fetch fa: %v", err)
	}

	if len(res.Attachments) > 0 {
		t.Errorf("Expected file to be deleted; attachments: %v", doc.Attachments)
	}
}
