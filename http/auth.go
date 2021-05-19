package http

import (
	"encoding/json"
	"net/http"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

type login struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required"`
}

type chpasswd struct {
	Old string `json:"old" validate:"required"`
	New string `json:"new" validate:"required"`
}

type Auth struct {
	ss sessions.Store
	us stores.Store
	ts stores.TokenStore

	validate *validator.Validate
}

func NewAuth(env *services.Env) *Auth {
	return &Auth{
		us: env.UserStore,
		ss: env.SessionStore,
		ts: env.TokenStore,

		validate: env.Validator,
	}
}

func comparePasswords(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
	var l login
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.validate.Struct(l); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	doc, err := a.us.GetByIndex(r.Context(), l.Email)
	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	user := doc.Data.(*models.User)

	if !user.IsActive {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if !comparePasswords(user.Password, l.Password) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	token, err := a.ts.Create(r.Context(), models.SessionToken, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		sentry.Report(err, "Failed to create session token")
		return
	}

	// Remove any left-over cookie.
	w.Header().Del("Set-Cookie")

	// Encode and write session.
	s := services.Session(a.ss, r)
	s.Values["id"] = token.ID
	s.Values["uuid"] = doc.ID
	services.SaveSession(s, r, w)

	json.NewEncoder(w).Encode(doc)
}

func (a *Auth) changePassword(w http.ResponseWriter, r *http.Request) {
	cv := services.FromContext(r.Context())
	if !cv.Authorized() {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	var cp chpasswd
	if err := json.NewDecoder(r.Body).Decode(&cp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.validate.Struct(cp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !comparePasswords(cv.User.Password, cp.Old) {
		http.Error(w, "wrong password", http.StatusBadRequest)
		return
	}

	cv.User.SetPassword(cp.New)
	if _, err := a.us.Update(r.Context(), models.Wrap(cv.User)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		sentry.Report(err, "Failed to change password", sentry.CaptureRequest(r))
	}
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	cv := services.FromContext(r.Context())
	a.ts.Invalidate(r.Context(), models.SessionToken, cv.ID)

	s := services.Session(a.ss, r)
	for k := range s.Values {
		delete(s.Values, k)
	}
	s.Options.MaxAge = -1
	services.SaveSession(s, r, w)
}

func (a *Auth) confirm(w http.ResponseWriter, r *http.Request) {
	var tid = mustExtractUUID(r)

	token, err := a.ts.Get(r.Context(), models.CreateToken, tid)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	udoc, err := a.us.Get(r.Context(), token.UserID)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	udoc.Data.(*models.User).IsActive = true

	if _, uerr := a.us.Update(r.Context(), udoc); uerr != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
}
