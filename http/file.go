// TODO: Remove all of this after migrating meetings

package http

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func uploadFile(w http.ResponseWriter, r *http.Request, sess sessions.Store, st stores.Store, uploadPath string) {
	id := extractUUID(r)

	doc, err := st.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	file, fhandler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fhandler.Filename = generateFilename(fhandler.Filename, doc.Attachments)

	if !controller.CheckFilenameRe.MatchString(fhandler.Filename) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ut := r.FormValue("upload-type")
	att, err := WriteFile(uploadPath, file, fhandler, id, ut)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		sentry.Report(err, "Failed to write uploaded file", sentry.CaptureRequest(r))
		return
	}

	if err := st.PutAttachment(r.Context(), doc, att); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		sentry.Report(err, "Failed to put attachment", sentry.CaptureRequest(r))
		return
	}

	if err := updateUploadFields(r, st, doc, fhandler.Filename, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		sentry.Report(err, "Failed to update related field after upload", sentry.CaptureRequest(r))
		return
	}

	location := *r.URL
	base, _ := path.Split(location.Path) // remove '/upload' from the end
	location.Path = path.Join(base, fhandler.Filename)
	w.Header().Set("Location", location.String())
}

// generateFilename takes the name of the file and tries to make it unique
// by first appending a number suffix ranging 1..20 and if that does not
// work proceeds to append randomly encoded hex string after the filename
// to ensure the uniqueness.
func generateFilename(basename string, atts map[string]models.Attachment) string {
	const (
		numberTo  = 20
		randTries = 3
	)
	filename := basename

	for i := 1; i < numberTo+randTries; i++ {
		if _, ok := atts[filename]; !ok {
			break
		}

		if i >= numberTo {
			filename = randSuffix(basename)
		} else {
			filename = numberSuffix(basename, i)
		}

	}
	return filename
}

func numberSuffix(filename string, n int) string {
	ext := filepath.Ext(filename)
	return fmt.Sprintf("%s (%d)%s", strings.TrimSuffix(filename, ext), n, ext)
}

func randSuffix(filename string) string {
	suffix := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, suffix); err != nil {
		sentry.Report(err)
		// this read should never fail so OK to panic if it actually does.
		panic(err)
	}

	ext := filepath.Ext(filename)
	return fmt.Sprintf("%s_%s%s",
		strings.TrimSuffix(filename, ext),
		hex.EncodeToString(suffix), ext)
}

func WriteFile(path string, file multipart.File, fh *multipart.FileHeader, id uuid.UUID, uploadType string) (*models.Attachment, error) {
	var att = &models.Attachment{
		Value:       models.Value{ID: uuid.New()},
		Name:        fh.Filename,
		Owner:       id,
		ContentType: fh.Header.Get("Content-Type"),
		UploadType:  models.UploadType(uploadType),
		Size:        fh.Size,
	}

	f, err := os.Create(filepath.Join(path, fmt.Sprintf("%s-%s", id, att.Value.ID)))
	if err != nil {
		return nil, err
	}

	defer f.Close()
	defer file.Seek(0, io.SeekStart)
	_, err = io.Copy(f, file)
	return att, err
}

func updateUploadFields(r *http.Request, st stores.Store, doc *models.Document, filename string, id uuid.UUID) error {
	var (
		dirty models.Entity
		kind  = r.FormValue("kind")
		url   = fileURL(doc.Kind, filename, doc.ID)
	)
	switch e := doc.Data.(type) {
	case *models.User:
		if kind == "avatar" {
			e.Avatar = url
			dirty = e
		} else if kind == "identity" {
			e.Identity = url
			dirty = e
		}
	case *models.Organization:
		if kind == "logo" {
			e.Logo = url
			dirty = e
		}
	}
	if dirty != nil {
		doc.Data = dirty
		if _, err := st.Update(r.Context(), doc); err != nil {
			return err
		}
	}
	return nil
}

// fileURL returns path to a static file with given name attached to document
// with given kind and id.
func fileURL(kind, name string, id uuid.UUID) string {
	return path.Join("/", kind, id.String(), url.PathEscape(name))
}

func getFile(w http.ResponseWriter, r *http.Request, store stores.Store, uploadPath string) {
	var (
		fname, err = url.PathUnescape(mux.Vars(r)["filename"])
		id         = extractUUID(r)
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	doc, derr := store.Get(r.Context(), id)
	if derr != nil {
		http.Error(w, derr.Error(), http.StatusNotFound)
		return
	}

	att, aerr := store.GetAttachment(r.Context(), doc, fname)
	if aerr != nil {
		http.NotFound(w, r)
		return
	}

	// postgresql
	filename := fmt.Sprintf("%s-%s", id, att.Value.ID)
	f, err := os.Open(filepath.Join(uploadPath, filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		sentry.Report(err, "Missing file of written attachment", sentry.CaptureRequest(r))
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

func delFile(w http.ResponseWriter, r *http.Request, store stores.Store) {
	var (
		doc *models.Document
		err error
	)

	id := extractUUID(r)
	fname, err := url.PathUnescape(mux.Vars(r)["filename"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	doc, err = store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = store.DeleteAttachment(r.Context(), doc, fname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
