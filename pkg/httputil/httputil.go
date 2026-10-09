// Package httputil contains http utilities.
package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/phoenix-industries/omega-server/pkg/auth"
)

func RunServer(ctx context.Context, srv *http.Server, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		serverErr <- nil
	}()
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutting down http server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		shutdownError := srv.Shutdown(shutdownCtx)
		if shutdownError != nil {
			logger.Error("server shutdown failed", "error", shutdownError)
		}
		if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return shutdownError
	}
}

func UserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

func IP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func Client(r *http.Request) string {
	return r.Header.Get("X-Client")
}

func GetAccessToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		header = r.URL.Query().Get("token")
		if header == "" {
			return "", ErrUnauthorized
		}
	}
	if !strings.HasPrefix(header, auth.TokenPrefix) {
		return "", ErrUnauthorized
	}
	token := header[len(auth.TokenPrefix):]
	if token == "" {
		return "", ErrUnauthorized
	}
	return token, nil
}

func GetUserID(auther *auth.Auth, r *http.Request) (string, error) {
	token, err := GetAccessToken(r)
	if err != nil {
		return "", err
	}
	jwt, err := auther.ParseJWT(token)
	if err != nil {
		return "", err
	}
	userID, err := jwt.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	return userID, nil
}

func GetUserRole(auther *auth.Auth, r *http.Request) (auth.Role, error) {
	token, err := GetAccessToken(r)
	if err != nil {
		return "", err
	}
	jwt, err := auther.ParseJWT(token)
	if err != nil {
		return "", err
	}
	claims, err := auther.GetJWTClaims(jwt)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

func CheckAuth(auther *auth.Auth, r *http.Request) error {
	token, err := GetAccessToken(r)
	if err != nil {
		return ErrUnauthorized
	}
	_, err = auther.ParseJWT(token)
	if err != nil {
		return ErrUnauthorized
	}
	return nil
}

func BodyJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return NewStatusError(
				nil,
				fmt.Sprintf("request body too large (max %d bytes)", maxErr.Limit),
				http.StatusRequestEntityTooLarge,
			)
		}
		return err
	}
	return nil
}
