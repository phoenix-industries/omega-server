package httputil

import (
	"log/slog"
	"net/http"
)

func NotFoundHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := NewStatusError(nil, "are you lost bud?", http.StatusNotFound).Response()
		if err := res.WriteJSON(w); err != nil {
			logger.ErrorContext(r.Context(), "encoding json error failed, that's an error for an error, congrats!", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}
