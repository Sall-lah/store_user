package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sall-lah/store_user/internal/config"
	"github.com/Sall-lah/store_user/internal/handler"
	"github.com/Sall-lah/store_user/internal/repository"
	"github.com/Sall-lah/store_user/internal/service"
)

func TestRouterRoutes(t *testing.T) {
	cfg := &config.Config{
		RateLimitMaxRequests:       60,
		RateLimitWindow:            time.Minute,
		RateLimitDeleteMaxRequests: 3,
		RateLimitDeleteWindow:      time.Minute,
	}

	repo := repository.NewMockUserProfileRepository()
	svc := service.NewUserService(repo, nil, nil, "user.events")
	h := handler.NewProfileHandler(svc)

	r := NewRouter(cfg, h, nil)

	// 1. Health check
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /health, got: %d", rec.Code)
	}

	// 2. Profile with auth header
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	req2.Header.Set("X-User-Id", validUUID)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 for /api/v1/users/profile, got: %d", rec2.Code)
	}
}
