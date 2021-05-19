package http

import (
	"context"
	"net/http"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type meeting struct {
	store   stores.Store
	session sessions.Store

	uploadPath string
}

func newMeeting(env *services.Env) *meeting {
	return &meeting{
		session: env.SessionStore,
		store:   env.MeetingsStore,

		uploadPath: env.Paths.Uploads,
	}
}

func (m *meeting) can(w http.ResponseWriter, r *http.Request) bool {
	meetingID := mustExtractUUID(r)
	meeting, err := m.store.Get(r.Context(), meetingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return false
	}

	md := meeting.Data.(*models.Meeting)

	if md.Project != nil {
		pdoc, err := m.store.FromKind("project").Get(r.Context(), *md.Project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return false
		}
		proj := pdoc.Data.(*models.Project)
		if controller.Can(r.Context(), controller.UploadProject, proj.ID, proj.Country) {
			return true
		}
	}

	host := md.Host
	can := m.authorized(r.Context(), services.Session(m.session, r), host)
	if !can {
		http.Error(w, "", http.StatusUnauthorized)
	}
	return can
}

func (m *meeting) authorized(ctx context.Context, s *sessions.Session, target uuid.UUID) bool {
	id, logged := s.Values["uuid"].(uuid.UUID)
	if !logged {
		return false
	}

	user, err := m.store.FromKind("user").Get(ctx, id)
	if err != nil {
		return false
	}

	for _, org := range user.Data.(*models.User).OrganizationRoles {
		if org.OrganizationID == target {
			return true
		}
	}
	return false
}

func (m *meeting) upload(w http.ResponseWriter, r *http.Request) {
	if m.can(w, r) {
		uploadFile(w, r, m.session, m.store, m.uploadPath)
	}
}

func (m *meeting) getFile(w http.ResponseWriter, r *http.Request) {
	if m.can(w, r) {
		getFile(w, r, m.store, m.uploadPath)
	}
}

func (m *meeting) delFile(w http.ResponseWriter, r *http.Request) {
	if m.can(w, r) {
		delFile(w, r, m.store)
	}
}
