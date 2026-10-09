package authservice

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database/models"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
)

type refreshData struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) HandleRefresh(w http.ResponseWriter, r *http.Request) *httputil.Response {
	var data refreshData
	if err := httputil.BodyJSON(w, r, &data); err != nil {
		return httputil.ErrInvalidBody.Response()
	}
	if data.RefreshToken == "" {
		return httputil.ErrBadRequest.Response()
	}

	res := AuthResponse{}
	ctx := r.Context()
	err := s.db.InTx(ctx, func(tx pgx.Tx) error {
		session, err := models.UserSessionGetByToken(ctx, tx, data.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		if session == nil {
			return httputil.ErrUnauthorized
		}
		if time.Now().After(session.ExpiresAt.Add(auth.DefaultJWTDuration)) {
			return httputil.ErrUnauthorized
		}
		user, err := models.UserGetByID(ctx, tx, session.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		client := httputil.Client(r)
		accessToken, err := s.auth.GenerateJWT(user.ID, client, user.Role)
		if err != nil {
			return httputil.NewStatusError(err, "failed to generate jwt", http.StatusInternalServerError)
		}

		res.UserID = user.ID
		res.TokenType = auth.TokenType
		res.AccessToken = accessToken
		res.RefreshToken = session.Token
		res.ExpiresAt = session.ExpiresAt.Unix()

		return nil
	})
	if err != nil {
		return httputil.ResponseFromError(err)
	}

	w.Header().Add("Authorization", auth.TokenPrefix+res.AccessToken)
	return httputil.NewResponseOK(http.StatusCreated, res)
}
