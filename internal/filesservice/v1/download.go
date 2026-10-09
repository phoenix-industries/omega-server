package filesservice

import (
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/phoenix-industries/omega-server/pkg/database/models"
)

func (s *Service) HandleDownload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)

		return
	}

	ctx := r.Context()
	media, err := models.MediaGetByID(ctx, s.db.Pool(), id)
	if err != nil {
		s.logger.Error("failed to get media", "error", err)
		http.Error(w, "failed to get media", http.StatusInternalServerError)
		return
	}
	if media == nil {
		http.Error(w, "media not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(filepath.Join(s.config.UploadsDir, media.ID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		s.logger.Error("failed to open file", "error", err)
		http.Error(w, "failed to open file", http.StatusInternalServerError)
		return

	}
	defer file.Close()

	w.Header().Set("Content-Type", media.Type)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": media.Name,
	}))
	w.Header().Set("Cache-Control", "private, max-age=86400")

	var modtime time.Time
	if media.UpdatedAt != nil {
		modtime = *media.UpdatedAt
	} else if media.CreatedAt != nil {
		modtime = *media.CreatedAt
	}

	http.ServeContent(w, r, media.Name, modtime, file)
}
