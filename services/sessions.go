package services

import (
	"encoding/gob"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"stageai.tech/sunshine/sunshine/sentry"
)

type key int

const (
	// SessKey is custom-typed int in order to avoid collisions when
	// storing in context with value.
	SessKey key = 0

	// SessName is the name of the session key that will be stored in
	// client's cookie jar.
	SessName = "_session"
)

// init registers uuid.UUID as expected value type to be stored in session
// store. This really is just implementation detail github.com/gorilla/session
// and the fact it uses encoding/gob for storing session data.
//
// For more information see the documentation of encoding/gob.
func init() {
	gob.Register(uuid.Nil)
}

// Session gets or creates new session for the given request.
//
// Failure could happen only on bad previous write or a malicious attempt, so
// in that cases send the error to sentry and move on.
func Session(store sessions.Store, r *http.Request) *sessions.Session {
	s, err := store.Get(r, SessName)
	if err != nil {
		sentry.Report(err, "services.Session failed")
	}
	return s
}

// SaveSession writes session.
//
// Failure could happen only on bad write, so in that cases this panics.
func SaveSession(s *sessions.Session, r *http.Request, w http.ResponseWriter) {
	if err := s.Save(r, w); err != nil {
		sentry.Report(err, "services.SaveSession failed")
		panic(err.Error())
	}
}
