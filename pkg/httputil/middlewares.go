package httputil

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/phoenix-industries/omega-server/internal/buildinfo"
	"github.com/phoenix-industries/omega-server/pkg/auth"
)

type Middleware func(http.Handler) http.Handler

type RequestID string

const RequestIDKey RequestID = "request_id"

func ChainMiddlewares(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			requestID, err := auth.GenerateID()
			if err != nil {
				logger.Error("failed to generate id", "error", err)
				http.Error(w, "failed to generate id", http.StatusInternalServerError)
				return
			}
			r = r.WithContext(context.WithValue(ctx, RequestIDKey, requestID))
			w.Header().Set("X-Request-ID", requestID)
			if buildinfo.DevMode() {
				start := time.Now()
				next.ServeHTTP(w, r)
				duration := time.Since(start)
				logger.Info("request handled", "request_id", requestID, "duration", duration, "client", Client(r), "user_agent", UserAgent(r), "ip", IP(r))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					ctx := r.Context()
					requestID := ctx.Value(RequestIDKey)
					logger.ErrorContext(
						ctx,
						"request panic",
						"request_id", requestID,
						"error", err,
						"stacktrace", debug.Stack(),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func AuthGuardMiddleware(auth *auth.Auth) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := CheckAuth(auth, r); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func AuthRoleMiddleware(auth *auth.Auth, requiredRole auth.Role) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := GetUserRole(auth, r)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !role.Allowed(requiredRole) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CorsMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
