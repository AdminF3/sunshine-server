package controller

import (
	"context"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

type Meeting struct {
	st         stores.Store
	notifier   stores.Notifier
	uploadPath string
}

func NewMeeting(e *services.Env) *Meeting {
	return &Meeting{
		st:         e.MeetingsStore,
		notifier:   e.Notifier,
		uploadPath: e.Paths.Uploads,
	}
}

func (m *Meeting) Create(ctx context.Context, meet *models.Meeting) (*models.Document, error) {
	if !m.can(ctx, m.st, &meet.Host) && !m.can(ctx, m.st, meet.Project) {
		return nil, ErrUnauthorized
	}
	return m.st.Create(ctx, meet)
}

func (m *Meeting) Get(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return nil, ErrUnauthorized
	}

	meeting, err := m.st.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	ids := []uuid.UUID{}
	if meeting.Data.(*models.Meeting).Project != nil {
		// ensure that the meeting does have project
		prjdoc, err := m.st.FromKind("project").Get(ctx, *meeting.Data.(*models.Meeting).Project)
		if err != nil {
			return nil, err
		}

		prj := prjdoc.Data.(*models.Project)

		for _, id := range prj.ConsortiumOrgs {
			ids = append(ids, uuid.MustParse(id))
		}
	}

	if !m.can(ctx, m.st, &id, ids...) {
		return nil, ErrUnauthorized
	}
	return meeting, nil
}

func (m *Meeting) Delete(ctx context.Context, id uuid.UUID) error {
	if !m.can(ctx, m.st, &id) {
		return ErrUnauthorized
	}
	meet, err := m.st.Get(ctx, id)
	if err != nil {
		return err
	}
	return m.st.Delete(ctx, meet)
}

func (m *Meeting) Update(ctx context.Context, upd models.Meeting) (*models.Document, error) {
	if !m.can(ctx, m.st, &upd.ID) {
		return nil, ErrUnauthorized
	}
	old, err := m.st.Get(ctx, upd.ID)
	if err != nil {
		return nil, err
	}
	o := old.Data.(*models.Meeting)
	newm := update(*o, upd)

	if len(upd.Guests) > 0 {
		err := m.st.DB().Unscoped().
			Where("meeting_id = ?", upd.ID).
			Delete(&models.MeetingGuest{}).Error

		if err != nil {
			return nil, err
		}

		newm.Guests = upd.Guests
	}
	res, err := m.st.Update(ctx, models.Wrap(&newm))

	return res, err
}

func (m *Meeting) List(ctx context.Context, id *uuid.UUID) ([]models.Document, error) {
	ids := []uuid.UUID{}
	if id != nil {
		prjdoc, err := m.st.FromKind("project").Get(ctx, *id)
		if err == nil {
			// this is project ID and not org ID
			prj := prjdoc.Data.(*models.Project)
			for _, id := range prj.ConsortiumOrgs {
				ids = append(ids, uuid.MustParse(id))
			}
		}
	}
	if !m.can(ctx, m.st, id, ids...) {
		return nil, ErrUnauthorized
	}
	var oid uuid.UUID
	if id != nil {
		oid = *id
	}
	docs, _, _, err := m.st.ListByMember(ctx, stores.Filter{}, oid)
	return docs, err
}

func update(old, new models.Meeting) models.Meeting {
	if new.Name != "" {
		old.Name = new.Name
	}

	if new.Host != uuid.Nil {
		old.Host = new.Host
	}

	if new.Location != "" {
		old.Location = new.Location
	}

	if !new.Date.IsZero() {
		old.Date = new.Date
	}

	if new.NextContact == nil {
		old.NextContact = nil
	} else if !new.NextContact.IsZero() {
		old.NextContact = new.NextContact
	}

	if new.Objective != "" {
		old.Objective = new.Objective
	}

	if new.Stakeholder != 0 {
		old.Stakeholder = new.Stakeholder
	}

	if new.Stage != "" {
		old.Stage = new.Stage
	}

	if new.ActionsTaken != "" {
		old.ActionsTaken = new.ActionsTaken
	}

	if new.Notes != "" {
		old.Notes = new.Notes
	}

	if new.Topic != nil {
		old.Topic = new.Topic
	}

	if new.InternalProject != "" {
		old.InternalProject = new.InternalProject
	}

	return old
}

func (m *Meeting) can(ctx context.Context, ms stores.Store, target *uuid.UUID, ids ...uuid.UUID) bool {
	cv := services.FromContext(ctx)
	if !cv.Authorized() {
		return false
	}
	u := cv.User

	if u.SuperUser || u.PlatformManager || u.AdminNwManager {
		return true
	}

	for _, role := range u.CountryRoles {
		// this is a bit ugly but is quick around to give all
		// country admins access to all meetings. Probably if
		// the `country admin` does not have right to see that
		// organization or project he will not be able to get
		// to `meeting screen`.
		if role.Role == models.CountryAdminRole {
			return true
		}
	}

	if target == nil {
		return false
	}

	orgs := append(ids, *target)
	for _, r := range u.OrganizationRoles {
		for _, org := range orgs {
			if r.OrganizationID == org {
				return true
			}
		}
	}

	for _, pr := range u.ProjectRoles {
		if pr.ProjectID == *target {
			return true
		}
	}
	// if target is not null and is not among the organization
	// roles then it tries to search a meeting with that ID.
	mdoc, err := ms.Get(ctx, *target)
	if err != nil {
		return false
	}

	for _, r := range u.OrganizationRoles {
		if r.OrganizationID == mdoc.Data.(*models.Meeting).Host {
			return true
		}
	}

	for _, pr := range u.ProjectRoles {
		if pr.ProjectID == *mdoc.Data.(*models.Meeting).Project {
			return true
		}
	}

	return false
}
