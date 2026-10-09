package filesservice

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/zeebo/blake3"

	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database/models"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
)

func (s *Service) HandleUpload(w http.ResponseWriter, r *http.Request) *httputil.Response {
	if err := httputil.CheckAuth(s.auth, r); err != nil {
		return httputil.ErrUnauthorized.Response()
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxSize)
	if err := r.ParseMultipartForm(s.config.MaxSize); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return httputil.NewStatusError(err, "file too large", http.StatusRequestEntityTooLarge).Response()
		}
		return httputil.NewStatusError(err, "failed to parse multipart form", http.StatusBadRequest).Response()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return httputil.NewStatusError(err, "failed to read file", http.StatusBadRequest).Response()
	}
	defer file.Close()
	if header.Size <= 0 {
		return httputil.NewStatusError(nil, "file is empty", http.StatusBadRequest).Response()
	}
	if header.Size > s.config.MaxSize {
		return httputil.NewStatusError(nil, "file too large", http.StatusRequestEntityTooLarge).Response()
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return httputil.NewStatusError(err, "failed to read file", http.StatusInternalServerError).Response()
	}
	fileType := http.DetectContentType(buf[:n])
	if !slices.Contains(s.config.AllowedMimeTypes, fileType) {
		return httputil.NewStatusError(nil, "invalid file type", http.StatusUnsupportedMediaType).Response()
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return httputil.NewStatusError(err, "failed to seek file", http.StatusInternalServerError).Response()
	}

	ctx := r.Context()
	renamed := false
	tmp, err := os.CreateTemp(s.config.UploadsDir, "upload-*")
	if err != nil {
		return httputil.NewStatusError(err, "failed to create temp file", http.StatusInternalServerError).Response()
	}
	defer func() {
		if !renamed {
			if err := os.Remove(tmp.Name()); err != nil {
				s.logger.ErrorContext(ctx, "failed to remove temp file", "error", err)
			}
		}
	}()

	hasher := blake3.New()
	if _, err := io.Copy(io.MultiWriter(tmp, hasher), file); err != nil {
		return httputil.NewStatusError(err, "failed to write file", http.StatusInternalServerError).Response()
	}
	if err := tmp.Close(); err != nil {
		return httputil.NewStatusError(err, "failed to close temp file", http.StatusInternalServerError).Response()
	}
	hash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	var id string
	if err := s.db.InTx(ctx, func(tx pgx.Tx) error {
		media, err := models.MediaGetByHash(ctx, tx, hash)
		if err != nil {
			return httputil.NewStatusError(err, "failed to get media", http.StatusInternalServerError)
		}
		if media != nil {
			id = media.ID
			return nil
		}
		id, err = auth.GenerateID()
		if err != nil {
			return httputil.NewStatusError(err, "failed to generate id", http.StatusInternalServerError)
		}
		media = &models.Media{
			Model: models.Model{
				ID: id,
			},
			Name: header.Filename,
			Type: fileType,
			Hash: hash,
		}
		if err := models.MediaInsert(ctx, tx, media); err != nil {
			return httputil.NewStatusError(err, "failed to create media", http.StatusInternalServerError)
		}
		if err := os.Rename(tmp.Name(), filepath.Join(s.config.UploadsDir, id)); err != nil {
			return httputil.NewStatusError(err, "failed to persist file", http.StatusInternalServerError)
		}
		renamed = true
		return nil
	}); err != nil {
		if renamed {
			if rerr := os.Rename(filepath.Join(s.config.UploadsDir, id), tmp.Name()); rerr != nil {
				s.logger.ErrorContext(ctx, "orphaned file", "path", filepath.Join(s.config.UploadsDir, id), "error", rerr)
			}
			renamed = false
		}
		return httputil.ResponseFromError(err)
	}

	return httputil.NewResponseOK(http.StatusCreated, map[string]string{"id": id})
}
