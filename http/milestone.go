package http

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type workphase struct {
	s sessions.Store
	c *controller.WorkPhase
}

func newWP(e *services.Env) *workphase {
	return &workphase{
		s: e.SessionStore,
		c: controller.NewWorkPhase(e),
	}
}

// Get existing workphase.
func (wp *workphase) getWP(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := wp.c.GetWP(r.Context(), mustExtractUUID(r))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	encode(w, doc, deps, err)
}

func (wp *workphase) uploadWP(w http.ResponseWriter, r *http.Request) {
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

	if err := wp.c.UploadWP(r.Context(), id, form); err != nil {
		writeError(w, r, err)
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // removes '/upload' from the end
	location.Path = path.Join(base, fheader.Filename)
	w.Header().Set("Location", location.String())
}

func (wp *workphase) getFileWP(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	att, f, err := wp.c.GetFileWP(r.Context(), id, fname)
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

func (wp *workphase) delFileWP(w http.ResponseWriter, r *http.Request) {
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := wp.c.DeleteFileWP(r.Context(), mustExtractUUID(r), fname); err != nil {
		writeError(w, r, err)
		return
	}
}

type monitoringphase struct {
	s sessions.Store
	c *controller.MonitoringPhase
}

func newMP(e *services.Env) *monitoringphase {
	return &monitoringphase{
		s: e.SessionStore,
		c: controller.NewMonitoringPhase(e),
	}
}

// Get existing monitoringphase.
func (mp *monitoringphase) getMP(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := mp.c.GetMP(r.Context(), mustExtractUUID(r))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	encode(w, doc, deps, err)
}

func (mp *monitoringphase) uploadMP(w http.ResponseWriter, r *http.Request) {
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

	if err := mp.c.UploadMP(r.Context(), id, form); err != nil {
		writeError(w, r, err)
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // removes '/upload' from the end
	location.Path = path.Join(base, fheader.Filename)
	w.Header().Set("Location", location.String())
}

func (mp *monitoringphase) getFileMP(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	att, f, err := mp.c.GetFileMP(r.Context(), id, fname)
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

func (mp *monitoringphase) delFileMP(w http.ResponseWriter, r *http.Request) {
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := mp.c.DeleteFileMP(r.Context(), mustExtractUUID(r), fname); err != nil {
		writeError(w, r, err)
		return
	}
}
