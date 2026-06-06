package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"app/docs/swagger"
	"app/internal/config"
	"app/internal/di"
	"app/internal/domain/user"
	"app/internal/interfaces/http/handler"
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

	// ── Public routes (no authentication) ──────────────────────────
	r.Get("/healthz", c.Health.Liveness)
	r.Get("/readyz", c.Health.Readiness)
	r.Handle("/metrics", promhttp.Handler())

	// Swagger UI and OpenAPI spec
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.InstanceName(swagger.SwaggerInfo.InstanceName()),
	))

	r.Route("/api/v1", func(r chi.Router) {
		registerPublicRoutes(r, c)
		registerAuthenticatedRoutes(r, c)
		registerAdminRoutes(r, c)
	})

	return r
}

// registerPublicRoutes — accessible without a token.
func registerPublicRoutes(r chi.Router, c *di.Container) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", c.Auth.Login)
		r.Post("/refresh", c.Auth.Refresh)
	})
}

// registerAuthenticatedRoutes — any authenticated user (admin or customer).
func registerAuthenticatedRoutes(r chi.Router, c *di.Container) {
	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.Authenticate(c.JWT))
		r.Post("/auth/logout", c.Auth.Logout)
		r.Get("/auth/me", c.Auth.Me)
	})
}

// registerAdminRoutes — admin role required; enforced at router layer.
func registerAdminRoutes(r chi.Router, c *di.Container) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(appmiddleware.Authenticate(c.JWT))
		r.Use(appmiddleware.RequireRole(user.RoleAdmin))

		r.Get("/", handler.AdminIndex)

		registerProductRoutes(r, c)
	})
}

func registerProductRoutes(r chi.Router, c *di.Container) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/search", c.Product.Search)
		r.Get("/", c.Product.List)
		r.Post("/", c.Product.Create)
		r.Get("/{id}", c.Product.Get)
		r.Put("/{id}", c.Product.Update)
		r.Delete("/{id}", c.Product.Delete)
		r.Patch("/{id}/inventory", c.Product.UpdateInventory)
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
