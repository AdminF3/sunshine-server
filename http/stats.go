package http

import (
	"encoding/json"
	"net/http"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"
)

type stats struct {
	store stores.Store
}

func newStats(env *services.Env) *stats {
	return &stats{store: env.UserStore}
}

func (s *stats) getCountryStats(w http.ResponseWriter, r *http.Request) {
	var c models.Country
	if country := r.URL.Query().Get("country"); len(country) > 0 {
		c = models.Country(country)
	}

	stats, err := stores.CountryStats(r.Context(), s.store, c)
	sentry.Report(err, sentry.CaptureRequest(r))
	json.NewEncoder(w).Encode(stats)
}
