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
)

type fa struct {
	c *controller.ForfaitingAgreement
}

func newForfaitingApplication(env *services.Env) *fa {
	return &fa{c: controller.NewForfaitingAgreement(env)}
}

func (f *fa) upload(w http.ResponseWriter, r *http.Request) {
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

	if err := f.c.UploadFile(r.Context(), id, form); err != nil {
		writeError(w, r, err)
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // removes '/upload' from the end
	location.Path = path.Join(base, fheader.Filename)
	w.Header().Set("Location", location.String())
}

func (f *fa) getFile(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	att, file, err := f.c.GetFile(r.Context(), id, fname)
	if err != nil {
		writeError(w, r, err)
		return
	}
	defer file.Close()

	if !strings.HasPrefix(mime.TypeByExtension(filepath.Ext(fname)), "image/") {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fname))

	}
	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(att.Size, 10))

	if r.Method != "HEAD" {
		io.Copy(w, file)
	}
}

func (f *fa) delFile(w http.ResponseWriter, r *http.Request) {
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := f.c.DeleteFile(r.Context(), mustExtractUUID(r), fname); err != nil {
		writeError(w, r, err)
		return
	}
}
