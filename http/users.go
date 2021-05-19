package http

import (
	"encoding/json"
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

type user struct {
	ss sessions.Store
	c  *controller.User
}

func newUser(env *services.Env) *user {
	uc := controller.NewUser(env)

	return &user{
		ss: env.SessionStore,
		c:  uc,
	}
}

// Update existing user.
func (h *user) update(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := h.c.Update(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}

	encode(w, doc, deps, err)
}

// Create new user.
func (h *user) create(w http.ResponseWriter, r *http.Request) {
	doc, err := h.c.Create(r.Context(), r.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Add("Location", path.Join(r.URL.String(), doc.ID.String()))

	json.NewEncoder(w).Encode(doc)
}

// Get existing user.
func (h *user) get(w http.ResponseWriter, r *http.Request) {
	doc, deps, err := h.c.Get(r.Context(), mustExtractUUID(r))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	encode(w, doc, deps, err)
}

// List all users.
func (h *user) list(w http.ResponseWriter, r *http.Request) {
	docs, deps, n, err := h.c.List(r.Context(), ParseFilter(r.URL.Query()))

	w.Header().Set(countHeader, strconv.Itoa(n))
	encode(w, docs, deps, err)
}

func (h *user) upload(w http.ResponseWriter, r *http.Request) {
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

	if err := h.c.UploadFile(r.Context(), mustExtractUUID(r), form); err != nil {
		writeError(w, r, err)
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // removes '/upload' from the end
	location.Path = path.Join(base, fheader.Filename)
	w.Header().Set("Location", location.String())
}

func (h *user) getFile(w http.ResponseWriter, r *http.Request) {
	filename, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}
	id := mustExtractUUID(r)

	att, f, err := h.c.GetFile(r.Context(), id, filename)
	if err != nil {
		writeError(w, r, err)
		return
	}
	defer f.Close()

	if !strings.HasPrefix(mime.TypeByExtension(filepath.Ext(filename)), "image/") {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	}
	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(att.Size, 10))

	if r.Method != "HEAD" {
		io.Copy(w, f)
	}
}

func (h *user) delFile(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	filename, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := h.c.DeleteFile(r.Context(), id, filename); err != nil {
		writeError(w, r, err)
		return
	}
}
