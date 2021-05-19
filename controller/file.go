package controller

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/sentry"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func uploadFile(ctx context.Context, st stores.Store, n stores.Notifier, form RequestForm, target *models.Document, uploadPath string) error {
	// target represents the document type (user, asset or project)

	form.FileHeader.Filename = generateFilename(form.FileHeader.Filename, target.Attachments)

	if !CheckFilenameRe.MatchString(form.FileHeader.Filename) {
		return ErrBadInput
	}

	err := notify(ctx, n, form, target)
	if err != nil {
		return fmt.Errorf("%w: %v", err, ErrBadInput)
	}

	att, err := writeFile(uploadPath, form.File, form.FileHeader, target.ID, form.UploadType)
	if err != nil {
		return fmt.Errorf("%w: %v", err, ErrFatal)
	}
	att.Comment = form.Comment

	if err := st.PutAttachment(ctx, target, att); err != nil {
		return fmt.Errorf("%w: %v", err, ErrFatal)
	}

	if err := updateUploadFields(ctx, st, target, form.FileHeader.Filename, form.Kind); err != nil {
		return fmt.Errorf("%w: %v", err, ErrFatal)
	}

	return nil
}

func notify(ctx context.Context, n stores.Notifier, form RequestForm, target *models.Document) error {
	user := services.FromContext(ctx).User

	if form.Kind == "learApply" {
		org, err := form.RequestCommentOrganization()
		if err != nil {
			return err
		}
		doc, err := n.GetDocument(ctx, org.ID, models.OrganizationT)
		if err != nil {
			return err
		}

		o := *doc.Data.(*models.Organization)
		not := models.Notification{
			RecipientID: o.Roles.LEAR,
			UserID:      user.ID,
			Action:      models.UserActionLEARApply,
			UserKey:     user.Name,
			TargetID:    org.ID,
			TargetKey:   o.Name,
			TargetType:  models.EntityType(models.OrganizationT),
			New:         form.FileHeader.Filename,
			Old:         "",
			Country:     o.Country,
		}
		go n.Notify(ctx, &not)
		return nil
	}

	if form.Kind == "claimResidency" {
		adoc, err := form.RequestCommentAsset()
		if err != nil {
			return err
		}
		asset := adoc.Data.(*models.Asset)
		doc, err := n.GetDocument(ctx, asset.Owner, models.OrganizationT)
		if err != nil {
			return err
		}
		o := *doc.Data.(*models.Organization)

		targetKey := asset.Address

		ak := &stores.AssetKey{
			Address: asset.Address,
			ESCO:    asset.ESCO,
		}

		tkj, err := json.Marshal(ak)
		if err == nil {
			targetKey = string(tkj)
		}

		not := models.Notification{
			RecipientID: o.Roles.LEAR,
			UserID:      user.ID,
			Action:      models.UserActionClaimResidency,
			UserKey:     user.Name,
			TargetID:    asset.ID,
			TargetKey:   targetKey,
			TargetType:  models.EntityType(models.AssetT),
			New:         form.FileHeader.Filename,
			Old:         "",
			Country:     asset.Country,
		}

		go n.Notify(ctx, &not)
		return nil
	}

	go n.Broadcast(ctx, models.UserActionUpload, *user, *target, "", form.FileHeader.Filename, user.ID, nil)
	return nil

}

func getFile(ctx context.Context, st stores.Store, id uuid.UUID, filename, upath string) (*models.Attachment, *os.File, error) {
	doc, err := st.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	att, err := st.GetAttachment(ctx, doc, filename)
	if err != nil {
		return nil, nil, err
	}

	fname := fmt.Sprintf("%s-%s", id, att.Value.ID)
	f, err := os.Open(filepath.Join(upath, fname))
	if err != nil {
		return nil, nil, err
	}
	return att, f, nil
}

func writeFile(path string, file multipart.File, fh *multipart.FileHeader, id uuid.UUID, uploadType string) (*models.Attachment, error) {
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

func updateUploadFields(ctx context.Context, st stores.Store, doc *models.Document, filename, kind string) error {
	var dirty models.Entity
	url := fileURL(doc.Kind, filename, doc.ID)

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
		if _, err := st.Update(ctx, doc); err != nil {
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
		if _, ok := atts[url.PathEscape(filename)]; !ok {
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

type Upload struct {
	File        io.Reader
	Filename    string
	Size        int64
	ContentType string
}

func uploadGQLFiles(st stores.Store, uploads []Upload, target uuid.UUID, path string) (err error) {
	g := new(errgroup.Group)

	for _, u := range uploads {
		u := u
		g.Go(func() error {
			return uploadGQLfile(st, u, target, path)
		})
	}
	return g.Wait()
}

func uploadGQLfile(st stores.Store, u Upload, target uuid.UUID, path string) error {
	var att = &models.Attachment{
		Value:       models.Value{ID: uuid.New()},
		Name:        u.Filename,
		Owner:       target,
		ContentType: u.ContentType,
		Size:        u.Size,
	}

	f, err := os.Create(filepath.Join(path, fmt.Sprintf("%s-%s", target, att.Value.ID)))
	if err != nil {
		return err
	}

	defer f.Close()
	if _, err := io.Copy(f, u.File); err != nil {
		return err
	}

	return st.DB().Save(att).Error
}
