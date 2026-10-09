package authservice

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/phoenix-industries/omega-server/pkg/database/models"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
)

type logoutData struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) HandleLogout(w http.ResponseWriter, r *http.Request) *httputil.Response {
	var data logoutData
	if err := httputil.BodyJSON(w, r, &data); err != nil {
		return httputil.NewStatusError(nil, "invalid request body", http.StatusBadRequest).Response()
	}
	if data.RefreshToken == "" {
		return httputil.ErrUnauthorized.Response()
	}

	ctx := r.Context()
	err := s.db.InTx(ctx, func(tx pgx.Tx) error {
		session, err := models.UserSessionGetByToken(ctx, tx, data.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		if session == nil {
			return httputil.ErrUnauthorized
		}
		if err := models.UserSessionDeleteByID(ctx, tx, session.ID); err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}
		return nil
	})
	if err != nil {
		return httputil.ResponseFromError(err)
	}

	return httputil.NewResponseOK(http.StatusNoContent, nil)
}
