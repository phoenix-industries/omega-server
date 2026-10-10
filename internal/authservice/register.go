package authservice

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database/models"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
	"github.com/phoenix-industries/omega-server/pkg/validate"
)

type registerData struct {
	Name        string            `json:"name"`
	Email       string            `json:"email"`
	Phone       *string           `json:"phone"`
	Password    string            `json:"password"`
	Gender      models.UserGender `json:"gender"`
	Birthdate   time.Time         `json:"birthdate"`
	City        *string           `json:"city"`
	Governorate *string           `json:"governorate"`
	Address     *string           `json:"address"`
}

func (s *Service) HandleRegister(w http.ResponseWriter, r *http.Request) *httputil.Response {
	var data registerData
	if err := httputil.BodyJSON(w, r, &data); err != nil {
		return httputil.ErrInvalidBody.Response()
	}

	if err := validate.Password(data.Password); err != nil {
		return httputil.NewResponseError(http.StatusBadRequest, err)
	}

	user := models.User{
		Name:        data.Name,
		Email:       data.Email,
		Phone:       data.Phone,
		Role:        auth.RoleMember,
		Gender:      data.Gender,
		Birthdate:   data.Birthdate,
		City:        data.City,
		Governorate: data.Governorate,
		Address:     data.Address,
	}
	if err := user.Validate(); err != nil {
		return httputil.NewStatusError(nil, err.Error(), http.StatusBadRequest).Response()
	}

	res := AuthResponse{}
	ctx := r.Context()
	err := s.db.InTx(ctx, func(tx pgx.Tx) error {
		if exists, err := models.UserExistsWithEmail(ctx, tx, user.Email); err != nil {
			return httputil.NewStatusError(err, "failed to get user by email", http.StatusInternalServerError)
		} else if exists {
			return httputil.NewStatusError(nil, "user with this email already exists", http.StatusConflict)
		}

		if user.Phone != nil {
			if exists, err := models.UserExistsWithPhone(ctx, tx, *user.Phone); err != nil {
				return httputil.NewStatusError(err, "failed to get user by phone", http.StatusInternalServerError)
			} else if exists {
				return httputil.NewStatusError(nil, "user with this phone number already exists", http.StatusConflict)
			}
		}

		userID, err := s.auth.GenerateID()
		if err != nil {
			return httputil.NewStatusError(err, "failed to generate id", http.StatusInternalServerError)
		}
		user.ID = userID

		hash, err := s.auth.HashPassword(data.Password)
		if err != nil {
			return httputil.NewStatusError(err, "failed to hash password", http.StatusInternalServerError)
		}
		user.Password = hash

		if err := models.UserInsert(ctx, tx, &user); err != nil {
			return httputil.NewStatusError(err, "failed to create user", http.StatusInternalServerError)
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

		res.User = &user
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
	return httputil.NewResponseOK(http.StatusCreated, &res)
}
