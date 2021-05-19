package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestMilestones(t *testing.T) {
	t.Run("upload_WP", testUploadWP)
	t.Run("upload_del_WP", testUploadDelWP)
	t.Run("upload_MP", testUploadMP)
	t.Run("upload_del_MP", testUploadDelMP)
}

func testUploadWP(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()
	usr := stores.NewTestAdmin(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore)
	wp := stores.NewTestWorkPhase(t, e.WPStore, prj.ID)
	getr := httptest.NewRequest("GET", "/workphase/"+wp.ID.String(), nil)
	doc := models.Document{Data: &models.WorkPhase{}}
	data := make(url.Values)

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "workphase", wp.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()

	// create another file to test same name collision and renaming
	r = createTestFileRequest(t, "workphase", wp.ID, data)
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

func testUploadDelWP(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()
	usr := stores.NewTestAdmin(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore)
	wp := stores.NewTestWorkPhase(t, e.WPStore, prj.ID)
	getr := httptest.NewRequest("GET", "/workphase/"+wp.ID.String(), nil)
	doc := models.Document{Data: &models.WorkPhase{}}
	data := make(url.Values)

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "workphase", wp.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	delr := httptest.NewRequest("DELETE", "/workphase/"+wp.ID.String()+"/gg.jpg", nil)
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

func testUploadMP(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()
	usr := stores.NewTestAdmin(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore)
	mp := stores.NewTestMonitoringPhase(t, e.MPStore, prj.ID)
	getr := httptest.NewRequest("GET", "/monitoringphase/"+mp.ID.String(), nil)
	doc := models.Document{Data: &models.MonitoringPhase{}}
	data := make(url.Values)

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "monitoringphase", mp.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()

	// create another file to test same name collision and renaming
	r = createTestFileRequest(t, "monitoringphase", mp.ID, data)
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

func testUploadDelMP(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	w := httptest.NewRecorder()
	usr := stores.NewTestAdmin(t, e.UserStore)
	prj := stores.NewTestProject(t, e.ProjectStore)
	mp := stores.NewTestMonitoringPhase(t, e.MPStore, prj.ID)
	getr := httptest.NewRequest("GET", "/monitoringphase/"+mp.ID.String(), nil)
	doc := models.Document{Data: &models.MonitoringPhase{}}
	data := make(url.Values)

	data.Add("upload-type", "general leaflet")
	r := createTestFileRequest(t, "monitoringphase", mp.ID, data)
	r = loginAs(t, e, usr, r)

	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	delr := httptest.NewRequest("DELETE", "/monitoringphase/"+mp.ID.String()+"/gg.jpg", nil)
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
