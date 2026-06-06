package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
)

// Recovery recovers from panics and returns a 500 error response.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
						slog.String("request_id", GetRequestID(r.Context())),
						slog.String("path", r.URL.Path),
					)
					response.Error(w, r, log, apperror.Internal("an unexpected error occurred"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
