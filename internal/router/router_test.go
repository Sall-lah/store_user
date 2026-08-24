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
	profileH := handler.NewProfileHandler(svc)

	notifRepo := repository.NewMockNotificationRepository()
	notifSvc := service.NewNotificationService(notifRepo)
	notifH := handler.NewNotificationHandler(notifSvc)

	docH := handler.NewDocHandler("docs/openapi.yaml", "docs/openapi.json")

	r := NewRouter(cfg, profileH, notifH, docH, nil)

	// 1. Health check
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /health, got: %d", rec.Code)
	}

	// 2. Documentation routes
	docRoutes := []string{"/docs", "/swagger", "/docs/openapi.yaml", "/docs/openapi.json"}
	for _, path := range docRoutes {
		reqDoc := httptest.NewRequest(http.MethodGet, path, nil)
		recDoc := httptest.NewRecorder()
		r.ServeHTTP(recDoc, reqDoc)
		if recDoc.Code != http.StatusOK {
			t.Errorf("expected 200 for doc route %s, got: %d", path, recDoc.Code)
		}
	}

	// 3. Profile with auth header
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
	req2 := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
	req2.Header.Set("X-User-Id", validUUID)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 for /api/users/profile, got: %d", rec2.Code)
	}

	// 4. Notifications route
	reqNotif := httptest.NewRequest(http.MethodGet, "/api/users/notifications", nil)
	reqNotif.Header.Set("X-User-Id", validUUID)
	recNotif := httptest.NewRecorder()
	r.ServeHTTP(recNotif, reqNotif)
	if recNotif.Code != http.StatusOK {
		t.Errorf("expected 200 for /api/users/notifications, got: %d", recNotif.Code)
	}

	// 5. Notification preferences route
	reqPref := httptest.NewRequest(http.MethodGet, "/api/users/notifications/preferences", nil)
	reqPref.Header.Set("X-User-Id", validUUID)
	recPref := httptest.NewRecorder()
	r.ServeHTTP(recPref, reqPref)
	if recPref.Code != http.StatusOK {
		t.Errorf("expected 200 for /api/users/notifications/preferences, got: %d", recPref.Code)
	}

	// 6. Old v1 route should return 404 Not Found
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	req3.Header.Set("X-User-Id", validUUID)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusNotFound {
		t.Errorf("expected 404 for removed /api/v1/users/profile, got: %d", rec3.Code)
	}
}
