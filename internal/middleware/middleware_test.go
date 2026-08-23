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
