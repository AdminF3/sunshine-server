package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"
)

func TestGetCountryStats(t *testing.T) {
	e, cleanup := newTestEnv(t)
	defer cleanup()

	stores.NewTestProject(t, e.ProjectStore)
	stores.NewTestAsset(t, e.AssetStore)
	stores.NewTestAsset(t, e.OrganizationStore)

	router := New(e)
	all := fetchStats(t, router, "")[models.CountryLatvia]
	latvia := fetchStats(t, router, "?country=Latvia")[models.CountryLatvia]
	if latvia.Assets != 3 || latvia.Projects != 1 || latvia.Organizations != 4 {
		t.Errorf("got bad stats from Latvia: %#v", latvia)
	}

	if !reflect.DeepEqual(all, latvia) {
		t.Error("Got different result for the same country based on filter")
		t.Errorf("All:\t%#v", all)
		t.Errorf("Latvia:\t%#v", latvia)
	}
}

func fetchStats(t *testing.T, h http.Handler, urlSuffix string) map[models.Country]stores.Stats {
	w := httptest.NewRecorder()

	h.ServeHTTP(w, httptest.NewRequest("GET", "/country_stats"+urlSuffix, nil))
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	var stats map[models.Country]stores.Stats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Errorf("decoding failed: %s; got: %v", err, stats)
	}

	return stats
}
