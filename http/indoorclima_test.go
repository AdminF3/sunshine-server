package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"
)

var (
	ValidIndoorClima = `{
	"project":"%s",
	"airex_windows": {
		"baseyear_n": 3.14,
		"baseyear_n_1": 4.14,
		"baseyear_n_2": 5.14
	},
        "attic":{
		"zone1":{
			"area": 1001,
			"uvalue": 233.442
		},
		"zone2":{
			"area": 123.321,
			"uvalue": 443.42
		}
	},
	"external_door": {
			"num1":{
				"zone1": {
					"area": 289.1279,
					"uvalue": 769.0881,
					"outdoor_temp": {
						"baseyear_n": 65.10820,
						"baseyear_n_1": 378.7457,
						"baseyear_n_2": 2345
					},
					"tempdiff": {
						"baseyear_n": 261.6159,
						"baseyear_n_1": 27.14273,
						"baseyear_n_2": 6866
					},
					"heat_loss_coeff": 956.7037
				},
				"zone2": {
					"area": 289.1279,
					"uvalue": 769.0881,
					"outdoor_temp": {
						"baseyear_n": 65.10820,
						"baseyear_n_1": 378.7457,
						"baseyear_n_2": 2345
					},
					"tempdiff": {
						"baseyear_n": 261.6159,
						"baseyear_n_1": 27.14273,
						"baseyear_n_2": 6866
					},
					"heat_loss_coeff": 956.7037
				}
			}
	},
	"basement_pipes": [{
			"quality": 1,
			"installed_length": %f,
			"diameter": 11.11,
			"heat_loss_unit": 9999.99,
			"heat_loss_year": 333.333,
			"indoorclima_id": "%s"
		}],
	"attic_pipes": [{
			"quality": 1,
			"installed_length": %f,
			"diameter": 11.11,
			"heat_loss_unit": 9999.99,
			"heat_loss_year": 333.333,
			"indoorclima_id": "%s"
	}]
	}`
	IndoorClimaHiddenColumn = `{
		"project":"%s",
		"airex_windows": {
			"baseyear_n": 3.14,
			"baseyear_n_1": 4.14,
			"baseyear_n_2": 5.14
		},
			"attic":{
			"zone1":{
				"area": 1001,
				"uvalue": 233.442
			},
			"zone2":{
				"area": 123.321,
				"uvalue": 443.42
			}
		},
		"external_door": {
				"num1":{
					"zone1": {
						"area": 289.1279,
						"uvalue": 769.0881,
						"outdoor_temp": {
							"baseyear_n": 65.10820,
							"baseyear_n_1": 378.7457,
							"baseyear_n_2": 2345
						},
						"tempdiff": {
							"baseyear_n": 261.6159,
							"baseyear_n_1": 27.14273,
							"baseyear_n_2": 6866
						},
						"heat_loss_coeff": 956.7037
					},
					"zone2": {
						"area": 289.1279,
						"uvalue": 769.0881,
						"outdoor_temp": {
							"baseyear_n": 65.10820,
							"baseyear_n_1": 378.7457,
							"baseyear_n_2": 2345
						},
						"tempdiff": {
							"baseyear_n": 261.6159,
							"baseyear_n_1": 27.14273,
							"baseyear_n_2": 6866
						},
						"heat_loss_coeff": 956.7037
					}
				}
		},
		"basement_pipes": [{
				"quality": 1,
				"installed_length": %f,
				"diameter": 11.11,
				"heat_loss_unit": 9999.99,
				"heat_loss_year": 333.333,
				"indoorclima_id": "%s"
			}],
		"attic_pipes": [{
				"quality": 1,
				"installed_length": %f,
				"diameter": 11.11,
				"heat_loss_unit": 9999.99,
				"heat_loss_year": 333.333,
				"indoorclima_id": "%s"
		}],
		"CreatedAt": %q
		}`
)

func TestIndoorClima(t *testing.T) {
	t.Run("get", func(t *testing.T) { testGetIndoorClima(t) })
	t.Run("update", func(t *testing.T) { testUpdateIndoorClima(t) })
	t.Run("update/HiddenColumn", func(t *testing.T) { testUpdateIndoorClimaHiddenColumn(t) })

}

func testGetIndoorClima(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	pm := stores.NewTestUser(t, e.UserStore)
	p := stores.NewTestProject(t, e.ProjectStore, stores.TPrjWithPm(pm.ID))

	r := loginAs(t, e, pm,
		httptest.NewRequest("GET", "/project/"+p.ID.String()+"/indoorclima", nil))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	doc := models.Document{Data: &contract.IndoorClima{}}
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}
}

func testUpdateIndoorClima(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()
	router := New(e)

	icd, p := stores.NewTestInClima(t, e.IndoorClimaStore)
	updatedIC := fmt.Sprintf(ValidIndoorClima, p.ID, 111.111, icd.ID, 1.1, icd.ID)

	pm, err := e.UserStore.Get(ctx, p.Data.(*models.Project).Roles.PM[0])
	if err != nil {
		t.Fatalf("Can't fetch project's PM: %v", err)
	}

	r := loginAs(t, e, pm, httptest.NewRequest(
		"PUT", "/project/"+p.ID.String()+"/indoorclima", strings.NewReader(updatedIC)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	if err := json.NewDecoder(w.Body).Decode(&models.Document{Data: &contract.IndoorClima{}}); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	// Do a GET after updating to ensure everything is legit.
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/project/"+p.ID.String()+"/indoorclima", nil)
	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	doc := models.Document{Data: &contract.IndoorClima{}}
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	ic := doc.Data.(*contract.IndoorClima)
	if airex := ic.AirexWindows; airex.N != 3.14 {
		t.Errorf("AirexWindows.N = %v; expected %v", airex.N, 3.14)
	}
	if len(ic.BasementPipes) != 1 {
		t.Fatalf("expected 1 basement pipe after updating; got %d: %#v", len(ic.BasementPipes), ic.BasementPipes)
	}
	if len(ic.AtticPipes) != 1 {
		t.Fatalf("expected 1 attic pipe after updating; got: %d", len(ic.AtticPipes))
	}

}

func testUpdateIndoorClimaHiddenColumn(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	testTime := time.Date(2000, time.January, 15, 10, 20, 30, 0, time.UTC)
	ic, p := stores.NewTestInClima(t, e.IndoorClimaStore)
	updatedIC := fmt.Sprintf(IndoorClimaHiddenColumn, p.ID, 111.111, ic.ID, 1.1, ic.ID, testTime.Format(time.RFC3339))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		"PUT", "/project/"+p.ID.String()+"/indoorclima", strings.NewReader(updatedIC))

	doc := models.Document{Data: &contract.IndoorClima{}}
	pm, err := e.UserStore.Get(ctx, p.Data.(*models.Project).Roles.PM[0])
	if err != nil {
		t.Fatalf("Can't fetch project's PM: %v", err)
	}
	r = loginAs(t, e, pm, r)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	if err := json.NewDecoder(w.Body).Decode(&doc); err != nil {
		t.Fatalf("can't decode success response: %s", err)
	}

	createdAt := doc.Data.(*contract.IndoorClima).Value.CreatedAt
	if testTime.Equal(createdAt) {
		t.Fatalf("The user shouldn't be able to change Created_At column ; got %v", createdAt)
	}
}
