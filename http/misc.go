package http

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

// ParseFilter safely parses url.Values and returns a Filter value.
func ParseFilter(v url.Values) stores.Filter {
	var (
		q   stores.Filter
		err error
	)

	q.Ascending, _ = strconv.ParseBool(v.Get("asc"))
	q.Offset, _ = strconv.Atoi(v.Get("offset"))
	q.Search = v.Get("search")
	q.BuildingType, _ = strconv.Atoi(v.Get("building_type"))
	q.Status, _ = strconv.Atoi(v.Get("status"))
	q.LegalForm, _ = strconv.Atoi(v.Get("legal_form"))
	q.Country = models.Country(v.Get("country"))
	q.CountryRoles = v["country_roles"]
	q.Owner, _ = uuid.Parse(v.Get("owner"))
	parseNullableUUID(&q, "esco", v.Get("esco"))
	q.AssetOwner, _ = uuid.Parse(v.Get("asset_owner"))
	q.PlatformRoles = v["platform_roles"]
	q.RelatedOrganizationID, _ = uuid.Parse(v.Get("related_organization_id"))

	q.Limit, err = strconv.Atoi(v.Get("limit"))
	if err != nil {
		q.Limit = 25
	}

	return q
}

func extractUUID(r *http.Request) uuid.UUID {
	uid, _ := uuid.Parse(mux.Vars(r)["id"])
	return uid
}

func mustExtractUUID(r *http.Request) uuid.UUID {
	return uuid.Must(uuid.Parse(mux.Vars(r)["id"]))
}

// writeError writes given error to w via http.Error with status code deduced
// by inspecting err. If deducing fails uses http.StatusInternalServerError as
// status.
//
// If err is nil this function is a noop.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, controller.ErrUnauthorized):
		status = http.StatusUnauthorized
		err = errors.New("")
	case errors.Is(err, controller.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, gorm.ErrRecordNotFound):
		status = http.StatusNotFound
	case errors.Is(err, controller.ErrBadInput):
		status = http.StatusBadRequest
	case errors.Is(err, controller.ErrInvalidTable):
		status = http.StatusBadRequest
	case errors.Is(err, controller.ErrDuplicate):
		status = http.StatusBadRequest

	case gorm.IsRecordNotFoundError(err):
		status = http.StatusNotFound
	case errors.As(err, new(validator.ValidationErrors)):
		status = http.StatusBadRequest
	case errors.As(err, new(*validator.InvalidValidationError)):
		status = http.StatusBadRequest
	}

	if status == http.StatusInternalServerError {
		sentry.Report(err, sentry.CaptureRequest(r))
	}

	http.Error(w, err.Error(), status)
}

type nullableFilter struct {
	fields []string
}

func (nf nullableFilter) nullable(f string) bool {
	for _, item := range nf.fields {
		if item == f {
			return true
		}
	}
	return false
}

var nullableFilters = nullableFilter{fields: []string{"esco"}}

func parseNullableUUID(q *stores.Filter, k, v string) {
	if !nullableFilters.nullable(k) {
		return
	}

	if v == "null" {
		q.NullFields = append(q.NullFields, k)
		return
	}

	switch k {
	case "esco":
		q.ESCO, _ = uuid.Parse(v)
	}

}
