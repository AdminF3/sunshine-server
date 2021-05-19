package http

import (
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

type proj struct {
	// TODO: remove that session store
	s sessions.Store
	c *controller.Project
}

func newProject(env *services.Env) *proj {
	return &proj{
		s: env.SessionStore,
		c: controller.NewProject(env),
	}
}

//create new project
func (p *proj) create(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := p.c.Create(r.Context(), r.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Add("Location", path.Join(r.URL.String(), doc.ID.String()))
	encode(w, doc, deps, err)
}

//get existing project
func (p *proj) get(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := p.c.Get(r.Context(), mustExtractUUID(r))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	encode(w, doc, deps, err)
}

//update exisiting project
func (p *proj) update(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := p.c.Update(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
	}

	encode(w, doc, deps, err)
}

func (p *proj) reports(w http.ResponseWriter, r *http.Request) {
	docs, deps, n, err := p.c.Reports(r.Context(), ParseFilter(r.URL.Query()))
	if err != nil {
		writeError(w, r, err)
	}

	w.Header().Set(countHeader, strconv.Itoa(n))
	encode(w, docs, deps, err)
}

func (p *proj) list(w http.ResponseWriter, r *http.Request) {
	docs, deps, n, err := p.c.List(r.Context(), ParseFilter(r.URL.Query()), extractUUID(r))
	if err != nil {
		writeError(w, r, err)
	}

	w.Header().Set(countHeader, strconv.Itoa(n))
	encode(w, docs, deps, err)
}

func (p *proj) upload(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)

	file, fheader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ut := r.FormValue("upload-type")
	kind := r.FormValue("kind")
	comment := r.FormValue("comment")

	form := controller.RequestForm{
		FileHeader: fheader,
		File:       file,
		UploadType: ut,
		Kind:       kind,
		Comment:    comment,
	}

	if err := p.c.UploadFile(r.Context(), id, form); err != nil {
		writeError(w, r, err)
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // removes '/upload' from the end
	location.Path = path.Join(base, fheader.Filename)
	w.Header().Set("Location", location.String())
}

func (p *proj) getFile(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	att, f, err := p.c.GetFile(r.Context(), id, fname)
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

func (p *proj) delFile(w http.ResponseWriter, r *http.Request) {
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := p.c.DeleteFile(r.Context(), mustExtractUUID(r), fname); err != nil {
		writeError(w, r, err)
		return
	}
}

func (p *proj) addRole(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := p.c.AddRole(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
	}
	encode(w, doc, deps, err)
}

func (p *proj) removeRole(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := p.c.RemoveRole(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
	}
	encode(w, doc, deps, err)
}

func (p *proj) getMeetings(w http.ResponseWriter, r *http.Request) {
	fpath, err := p.c.ExportMeetings(r.Context(), extractUUID(r))
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
