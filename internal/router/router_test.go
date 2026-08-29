package router

import (
	"bytes"
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
	notifRepo := repository.NewMockNotificationRepository()
	svc := service.NewUserService(repo, notifRepo, nil, nil, "user.events")
	profileH := handler.NewProfileHandler(svc)
	notifSvc := service.NewNotificationService(notifRepo)
	notifH := handler.NewNotificationHandler(notifSvc)

	docH := handler.NewDocHandler("docs/openapi.yaml", "docs/openapi.json")
	adminH := handler.NewAdminHandler(svc)

	r := NewRouter(cfg, profileH, adminH, notifH, docH, nil)

	// 1. Health check
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /health, got: %d", rec.Code)
	}

	// 2. Documentation routes
	docRoutes := []string{
		"/docs",
		"/docs/",
		"/swagger",
		"/swagger/",
		"/docs/openapi.yaml",
		"/docs/openapi.json",
		"/swagger/openapi.yaml",
		"/swagger/openapi.json",
	}
	for _, path := range docRoutes {
		reqDoc := httptest.NewRequest(http.MethodGet, path, nil)
		recDoc := httptest.NewRecorder()
		r.ServeHTTP(recDoc, reqDoc)
		if recDoc.Code != http.StatusOK {
			t.Errorf("expected 200 for doc route %s, got: %d", path, recDoc.Code)
		}
	}

	// 3. User creation route with auth header
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(`{"fullName":"Bob"}`))
	reqCreate.Header.Set("X-User-Id", validUUID)
	recCreate := httptest.NewRecorder()
	r.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Errorf("expected 201 for POST /api/users, got: %d", recCreate.Code)
	}

	// 4. Profile with auth header
	req2 := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
	req2.Header.Set("X-User-Id", validUUID)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 for /api/users/profile, got: %d", rec2.Code)
	}

	// 5. Notifications route
	reqNotif := httptest.NewRequest(http.MethodGet, "/api/users/notifications", nil)
	reqNotif.Header.Set("X-User-Id", validUUID)
	recNotif := httptest.NewRecorder()
	r.ServeHTTP(recNotif, reqNotif)
	if recNotif.Code != http.StatusOK {
		t.Errorf("expected 200 for /api/users/notifications, got: %d", recNotif.Code)
	}

	// 6. Admin user deletion route - 403 for missing/customer role
	targetUUID := "11111111-2222-3333-4444-555555555555"
	reqAdminCust := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+targetUUID, nil)
	reqAdminCust.Header.Set("X-User-Id", validUUID)
	reqAdminCust.Header.Set("X-User-Role", "CUSTOMER")
	recAdminCust := httptest.NewRecorder()
	r.ServeHTTP(recAdminCust, reqAdminCust)
	if recAdminCust.Code != http.StatusForbidden {
		t.Errorf("expected 403 for customer accessing /api/admin/users, got: %d", recAdminCust.Code)
	}

	// 7. Admin user deletion route - 200 for ADMIN role
	reqAdminOK := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+targetUUID, nil)
	reqAdminOK.Header.Set("X-User-Id", validUUID)
	reqAdminOK.Header.Set("X-User-Role", "ADMIN")
	recAdminOK := httptest.NewRecorder()
	r.ServeHTTP(recAdminOK, reqAdminOK)
	if recAdminOK.Code != http.StatusOK {
		t.Errorf("expected 200 for admin accessing /api/admin/users, got: %d", recAdminOK.Code)
	}

	// 8. Old v1 route should return 404 Not Found
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	req3.Header.Set("X-User-Id", validUUID)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusNotFound {
		t.Errorf("expected 404 for removed /api/v1/users/profile, got: %d", rec3.Code)
	}
}
