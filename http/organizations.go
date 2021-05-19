package http

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type org struct {
	s sessions.Store
	c *controller.Organization
}

func newOrg(env *services.Env) *org {
	return &org{
		s: env.SessionStore,
		c: controller.NewOrganization(env),
	}
}

// Create new organization.
func (o *org) create(w http.ResponseWriter, r *http.Request) {
	doc, err := o.c.Create(r.Context(), r.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Add("Location", path.Join(r.URL.String(), doc.ID.String()))
	json.NewEncoder(w).Encode(doc)
}

// Get existing organization.
func (o *org) get(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := o.c.Get(r.Context(), mustExtractUUID(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	encode(w, doc, deps, err)
}

// Update existing organization.
func (o *org) update(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := o.c.Update(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}

	encode(w, doc, deps, err)
}

// List all organizations.
func (o *org) list(w http.ResponseWriter, r *http.Request) {
	docs, deps, n, err := o.c.List(r.Context(), ParseFilter(r.URL.Query()), extractUUID(r))
	w.Header().Set(countHeader, strconv.Itoa(n))
	encode(w, docs, deps, err)
}

func (o *org) upload(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)

	file, fheader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ut := r.FormValue("upload-type")
	kind := r.FormValue("kind")

	form := controller.RequestForm{
		FileHeader: fheader,
		File:       file,
		UploadType: ut,
		Kind:       kind,
	}

	if err := o.c.UploadFile(r.Context(), id, form); err != nil {
		writeError(w, r, err)
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // removes '/upload' from the end
	location.Path = path.Join(base, fheader.Filename)
	w.Header().Set("Location", location.String())
}

func (o *org) getFile(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	att, f, err := o.c.GetFile(r.Context(), id, fname)
	if err != nil {
		writeError(w, r, err)
		return
	}
	defer f.Close()

	if !strings.HasPrefix(mime.TypeByExtension(filepath.Ext(fname)), "image/") {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fname))

	}
	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(att.Size, 10))

	if r.Method != "HEAD" {
		io.Copy(w, f)
	}
}

func (o *org) delFile(w http.ResponseWriter, r *http.Request) {
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := o.c.DeleteFile(r.Context(), mustExtractUUID(r), fname); err != nil {
		writeError(w, r, err)
		return
	}
}

func (o *org) addRole(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := o.c.AddRole(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
	}
	encode(w, doc, deps, err)
}

func (o *org) removeRole(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := o.c.RemoveRole(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
	}
	encode(w, doc, deps, err)
}

func (o *org) getMeetings(w http.ResponseWriter, r *http.Request) {
	fpath, err := o.c.ExportMeetings(r.Context(), extractUUID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, err := os.Open(fpath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fstat, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+f.Name())
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Length", strconv.FormatInt(fstat.Size(), 10))

	if r.Method != "HEAD" {
		io.Copy(w, f)
	}
}
