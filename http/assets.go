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

type asset struct {
	s sessions.Store
	c *controller.Asset
}

func newAsset(env *services.Env) *asset {
	return &asset{
		s: env.SessionStore,
		c: controller.NewAsset(env),
	}
}

// Create new asset.
func (a *asset) create(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := a.c.Create(r.Context(), r.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Add("Location", path.Join(r.URL.String(), doc.ID.String()))
	encode(w, doc, deps, err)
}

// Get existing asset.
func (a *asset) get(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := a.c.Get(r.Context(), mustExtractUUID(r))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	encode(w, doc, deps, err)
}

// Update existing asset.
func (a *asset) update(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := a.c.Update(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
	}

	encode(w, doc, deps, err)
}

// List all asset.
func (a *asset) list(w http.ResponseWriter, r *http.Request) {
	docs, deps, n, err := a.c.List(r.Context(), extractUUID(r), ParseFilter(r.URL.Query()))

	w.Header().Set(countHeader, strconv.Itoa(n))
	encode(w, docs, deps, err)
}

func (a *asset) upload(w http.ResponseWriter, r *http.Request) {
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

	if err := a.c.UploadFile(r.Context(), id, form); err != nil {
		writeError(w, r, err)
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // removes '/upload' from the end
	location.Path = path.Join(base, fheader.Filename)
	w.Header().Set("Location", location.String())
}

func (a *asset) getFile(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	att, f, err := a.c.GetFile(r.Context(), id, fname)
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

func (a *asset) delFile(w http.ResponseWriter, r *http.Request) {
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := a.c.DeleteFile(r.Context(), mustExtractUUID(r), fname); err != nil {
		writeError(w, r, err)
		return
	}
}
