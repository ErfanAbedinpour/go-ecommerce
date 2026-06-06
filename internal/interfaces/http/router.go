package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"app/internal/config"
	"app/internal/di"
	appmiddleware "app/internal/interfaces/http/middleware"
)

// NewRouter creates the HTTP router with all middleware and routes.
func NewRouter(c *di.Container) http.Handler {
	r := chi.NewRouter()

	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Recovery(c.Log))
	r.Use(appmiddleware.Logging(c.Log))
	r.Use(appmiddleware.Metrics)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Compress(5))
	r.Use(corsMiddleware(c.Config))

	// Observability routes (no auth)
	r.Get("/healthz", c.Health.Liveness)
	r.Get("/readyz", c.Health.Readiness)
	r.Handle("/metrics", promhttp.Handler())

	// API v1 admin routes
	r.Route("/api/v1/admin", func(r chi.Router) {
		registerAuthRoutes(r, c)
	})

	return r
}

func registerAuthRoutes(r chi.Router, c *di.Container) {
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ecommerce admin API v1"}`))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", c.Auth.Login)
		r.Post("/refresh", c.Auth.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Authenticate(c.JWT))
			r.Post("/logout", c.Auth.Logout)
			r.Get("/me", c.Auth.Me)
		})
	})
}

func corsMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	})
}
