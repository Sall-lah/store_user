package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Sall-lah/store_user/internal/config"
	"github.com/Sall-lah/store_user/internal/handler"
	"github.com/Sall-lah/store_user/internal/middleware"
	"github.com/Sall-lah/store_user/internal/ratelimit"
)

// NewRouter configures global middleware, health endpoints, identity guards, and rate-limited API routes.
// Why: Centralizes all HTTP ingress routing and middleware orchestration for store_user.
func NewRouter(cfg *config.Config, profileHandler *handler.ProfileHandler, limiter ratelimit.Limiter) http.Handler {
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.MaxBodySize(64 * 1024)) // Max 64KB request body

	// CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Id", "X-User-Role", "X-User-Email"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Liveness / Readiness Probes
	r.Get("/health", profileHandler.HealthCheck)

	// API Routes Helper
	mountUserRoutes := func(sub chi.Router) {
		sub.Use(middleware.AuthIdentity)

		// Profile Read: e.g. 60 req/min
		sub.With(middleware.RateLimit(limiter, cfg.RateLimitMaxRequests, cfg.RateLimitWindow, "ratelimit:user:profile:read")).
			Get("/profile", profileHandler.GetProfile)

		// Profile Update: e.g. 15 req/min
		sub.With(middleware.RateLimit(limiter, 15, cfg.RateLimitWindow, "ratelimit:user:profile:update")).
			Put("/profile", profileHandler.UpdateProfile)

		// Account Deletion: e.g. 3 req/min
		sub.With(middleware.RateLimit(limiter, cfg.RateLimitDeleteMaxRequests, cfg.RateLimitDeleteWindow, "ratelimit:user:account:delete")).
			Delete("/account", profileHandler.DeleteAccount)
	}

	// Mount under standard /api/v1/users and alias /api/users
	r.Route("/api/v1/users", mountUserRoutes)
	r.Route("/api/users", mountUserRoutes)

	return r
}
