package authservice

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database/models"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
	"github.com/phoenix-industries/omega-server/pkg/validate"
)

type resetPasswordData struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

func (s *Service) HandleResetPassword(w http.ResponseWriter, r *http.Request) *httputil.Response {
	var data resetPasswordData
	if err := httputil.BodyJSON(w, r, &data); err != nil {
		return httputil.ErrInvalidBody.Response()
	}
	if data.Password == "" || data.NewPassword == "" || data.NewPassword == data.Password {
		return httputil.ErrBadRequest.Response()
	}

	if err := validate.Password(data.NewPassword); err != nil {
		return httputil.NewStatusError(nil, err.Error(), http.StatusBadRequest).Response()
	}

	userID, err := httputil.GetUserID(s.auth, r)
	if err != nil {
		return httputil.ResponseFromError(err)
	}

	ctx := r.Context()
	res := AuthResponse{}
	err = s.db.InTx(ctx, func(tx pgx.Tx) error {
		user, err := models.UserGetByID(ctx, tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if user == nil {
			return httputil.ErrUnauthorized
		}

		if valid, err := s.auth.VerifyPassword(data.Password, user.Password); err != nil {
			return httputil.NewStatusError(err, "request failed", http.StatusInternalServerError)
		} else if !valid {
			return httputil.ErrUnauthorized
		}

		hash, err := s.auth.HashPassword(data.NewPassword)
		if err != nil {
			return httputil.NewStatusError(err, "failed to hash password", http.StatusInternalServerError)
		}

		if err := models.UserUpdatePassword(ctx, tx, userID, hash); err != nil {
			return httputil.NewStatusError(err, "failed to update user", http.StatusInternalServerError)
		}

		sessionID, err := s.auth.GenerateID()
		if err != nil {
			return httputil.NewStatusError(err, "failed to generate id", http.StatusInternalServerError)
		}

		refreshToken, err := s.auth.GenerateToken()
		if err != nil {
			return httputil.NewStatusError(err, "failed to generate token", http.StatusInternalServerError)
		}

		session := models.UserSession{
			ID:        sessionID,
			UserID:    user.ID,
			Token:     refreshToken,
			IPAddress: httputil.IP(r),
			UserAgent: httputil.UserAgent(r),
		}
		if err := models.UserSessionInsert(ctx, tx, &session); err != nil {
			return httputil.NewStatusError(err, "failed to create session", http.StatusInternalServerError)
		}

		client := httputil.Client(r)
		accessToken, err := s.auth.GenerateJWT(user.ID, client, user.Role)
		if err != nil {
			return httputil.NewStatusError(err, "failed to generate jwt", http.StatusInternalServerError)
		}

		res.UserID = user.ID
		res.TokenType = auth.TokenType
		res.AccessToken = accessToken
		res.RefreshToken = refreshToken
		res.ExpiresAt = session.ExpiresAt.Unix()

		return nil
	})
	if err != nil {
		return httputil.ResponseFromError(err)
	}

	w.Header().Add("Authorization", auth.TokenPrefix+res.AccessToken)
	return httputil.NewResponseOK(http.StatusCreated, res)
}
