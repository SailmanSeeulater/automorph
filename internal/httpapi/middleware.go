package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/SailmanSeeulater/automorph/internal/auth"
)

// requestLogger logs one line request with status, size and duration.
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"dur", time.Since(start).Round(time.Microsecond),
				"req_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}

// adminOnly rejects /api/admin/* requests from non-admins before the body is read,
// so anonymous visitors can't make the server parse a 20 MB upload.
// Handlers still call requireAdmin; this is a second lock on the same door.
func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/admin/") {
			user, ok := auth.UserFrom(r.Context())
			if !ok {
				writeProblem(w, http.StatusUnauthorized, "not logged in")
				return
			}
			if user.Role != "admin" {
				writeProblem(w, http.StatusForbidden, "admins only")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// writeProblem sends an error in the same JSON shape huma uses.
func writeProblem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"title":  http.StatusText(status),
		"status": status,
		"detail": detail,
	})
}
