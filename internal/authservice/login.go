package authservice

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database/models"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
	"github.com/phoenix-industries/omega-server/pkg/validate"
)

const (
	errInvalidCredentials = "invalid credentials"
)

type loginData struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (s *Service) HandleLogin(w http.ResponseWriter, r *http.Request) *httputil.Response {
	var data loginData
	if err := httputil.BodyJSON(w, r, &data); err != nil {
		return httputil.NewStatusError(nil, "invalid request body", http.StatusBadRequest).Response()
	}

	ctx := r.Context()
	res := AuthResponse{}
	err := s.db.InTx(ctx, func(tx pgx.Tx) error {
		var user *models.User
		if validate.IsEmail(data.Identifier) {
			u, err := models.UserGetByEmail(ctx, s.db.Pool(), data.Identifier)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return httputil.NewStatusError(nil, errInvalidCredentials, http.StatusUnauthorized)
				}
				return httputil.NewStatusError(err, "request failed", http.StatusInternalServerError)
			}
			user = u
		} else if validate.IsPhoneNumber(data.Identifier) {
			u, err := models.UserGetByPhone(ctx, s.db.Pool(), data.Identifier)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return httputil.NewStatusError(nil, errInvalidCredentials, http.StatusUnauthorized)
				}
				return httputil.NewStatusError(err, "request failed", http.StatusInternalServerError)
			}
			user = u
		} else {
			return httputil.NewStatusError(nil, errInvalidCredentials, http.StatusUnauthorized)
		}
		if user == nil {
			return httputil.NewStatusError(nil, errInvalidCredentials, http.StatusUnauthorized)
		}

		if valid, err := s.auth.VerifyPassword(data.Password, user.Password); err != nil {
			return httputil.NewStatusError(err, "request failed", http.StatusInternalServerError)
		} else if !valid {
			return httputil.NewStatusError(nil, errInvalidCredentials, http.StatusUnauthorized)
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
