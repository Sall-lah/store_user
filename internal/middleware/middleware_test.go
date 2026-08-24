package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sall-lah/store_user/internal/ratelimit"
)

func TestAuthIdentityMiddleware(t *testing.T) {
	handler := AuthIdentity(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		role := GetUserRole(r.Context())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(userID + ":" + role))
	}))

	// 1. Missing header
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing header, got: %d", rec.Code)
	}

	// 2. Valid header
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
	req2 := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req2.Header.Set("X-User-Id", validUUID)
	req2.Header.Set("X-User-Role", "CUSTOMER")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", rec2.Code)
	}
	if rec2.Body.String() != validUUID+":CUSTOMER" {
		t.Errorf("expected body '%s:CUSTOMER', got: '%s'", validUUID, rec2.Body.String())
	}
}

func TestMaxBodySizeMiddleware(t *testing.T) {
	handler := MaxBodySize(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Payload > 10 bytes
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("this is a very long body exceeding 10 bytes"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got: %d", rec.Code)
	}
}

type mockLimiter struct {
	allowed bool
}

func (m *mockLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (*ratelimit.Result, error) {
	return &ratelimit.Result{
		Allowed:       m.allowed,
		Remaining:     0,
		Limit:         limit,
		RetryAfterSec: 10,
	}, nil
}

func (m *mockLimiter) Close() error {
	return nil
}

func TestRateLimitMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 1. Allowed
	limiterOk := &mockLimiter{allowed: true}
	mwOk := RateLimit(limiterOk, 10, time.Minute, "test")(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mwOk.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", rec.Code)
	}

	// 2. Blocked
	limiterBlock := &mockLimiter{allowed: false}
	mwBlock := RateLimit(limiterBlock, 10, time.Minute, "test")(nextHandler)
	rec2 := httptest.NewRecorder()
	mwBlock.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got: %d", rec2.Code)
	}
	if rec2.Header().Get("Retry-After") != "10" {
		t.Errorf("expected Retry-After=10, got: %s", rec2.Header().Get("Retry-After"))
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("admin access granted"))
	})

	mw := RequireRole("ADMIN", "SUPER_ADMIN")(nextHandler)
	fullPipeline := AuthIdentity(mw)

	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	// 1. Missing role claim in header -> 403 Forbidden
	req1 := httptest.NewRequest(http.MethodDelete, "/api/admin/users/123", nil)
	req1.Header.Set("X-User-Id", validUUID)
	rec1 := httptest.NewRecorder()
	fullPipeline.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for missing role, got: %d", rec1.Code)
	}

	// 2. Insufficient role (CUSTOMER) -> 403 Forbidden
	req2 := httptest.NewRequest(http.MethodDelete, "/api/admin/users/123", nil)
	req2.Header.Set("X-User-Id", validUUID)
	req2.Header.Set("X-User-Role", "CUSTOMER")
	rec2 := httptest.NewRecorder()
	fullPipeline.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for customer role, got: %d", rec2.Code)
	}

	// 3. Valid role (ADMIN) -> 200 OK
	req3 := httptest.NewRequest(http.MethodDelete, "/api/admin/users/123", nil)
	req3.Header.Set("X-User-Id", validUUID)
	req3.Header.Set("X-User-Role", "ADMIN")
	rec3 := httptest.NewRecorder()
	fullPipeline.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin role, got: %d", rec3.Code)
	}

	// 4. Case-insensitive valid role (admin) -> 200 OK
	req4 := httptest.NewRequest(http.MethodDelete, "/api/admin/users/123", nil)
	req4.Header.Set("X-User-Id", validUUID)
	req4.Header.Set("X-User-Role", "admin")
	rec4 := httptest.NewRecorder()
	fullPipeline.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Errorf("expected 200 OK for lowercase admin role, got: %d", rec4.Code)
	}
}

