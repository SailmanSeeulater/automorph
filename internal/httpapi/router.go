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

	"github.com/SailmanSeeulater/automorph/internal/auth"
	"github.com/SailmanSeeulater/automorph/internal/media"
	"github.com/SailmanSeeulater/automorph/internal/store"
)

const Version = "0.1.0"

// Deps is everything the HTTP layer needs. New services get added here.
type Deps struct {
	Logger        *slog.Logger
	DB            *sql.DB
	Queries       *store.Queries
	Auth          *auth.Service
	Media         *media.Store
	SecureCookies bool
}

// NewRouter builds the HTTP handler: plain chi routes plus the huma API under /api.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP) // trusts X-Forwarded-For; fine because we only listen behind cloudflared
	r.Use(requestLogger(d.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(http.NewCrossOriginProtection().Handler) // rejects cross-site POST/PUT/DELETE (CSRF)
	r.Use(d.Auth.LoadUser)
	r.Use(adminOnly)

	// Liveness check for systemd and uptime monitors. Not part of the public API.
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	// Uploaded images. Outside /api because they're plain files, not JSON.
	r.Handle("/media/*", d.Media.Handler())

	// Keep everything huma serves under /api so the React app can own every other path.
	cfg := huma.DefaultConfig("Automorph API", Version)
	cfg.OpenAPIPath = "/api/openapi"
	cfg.DocsPath = "/api/docs"
	cfg.SchemasPath = "/api/schemas"
	api := humachi.New(r, cfg)

	registerHealth(api, d.DB)
	registerAuth(api, d)
	registerPosts(api, d)
	registerMedia(api, d)

	return r
}
