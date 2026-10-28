package httpapi

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const Version = "0.1.0"

// NewRouter builds the HTTP handler: plain chi routes plus the huma API under /api
func NewRouter(logger *slog.Logger, db *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP) // trust X-Forwarded-For; fine because we only listen behind cloudflared
	r.Use(requestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Liveness check for systemd and uptime monitors. Not part of the public API
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	// Keep everything huma serves under /api so the React app can own every other path.
	cfg := huma.DefaultConfig("Automorph API", Version)
	cfg.OpenAPIPath = "/api/openapi"
	cfg.DocsPath = "/api/docs"
	cfg.SchemasPath = "/api/schemas"
	api := humachi.New(r, cfg)

	registerHealth(api, db)

	return r
}
