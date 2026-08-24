package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/Sall-lah/store_user/internal/middleware"
	"github.com/Sall-lah/store_user/internal/repository"
	"github.com/Sall-lah/store_user/internal/service"
)

func setupTestNotificationHandler() (*NotificationHandler, *repository.MockNotificationRepository) {
	repo := repository.NewMockNotificationRepository()
	svc := service.NewNotificationService(repo)
	return NewNotificationHandler(svc), repo
}

func TestNotificationHandler_ListNotifications(t *testing.T) {
	h, repo := setupTestNotificationHandler()
	userID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	_, _ = repo.Create(context.Background(), repository.NotificationRecord{
		UserID:  userID,
		Title:   "Order Shipped",
		Message: "Your order is on the way.",
		Type:    "ORDER_UPDATE",
	})

	// 1. Success with auth header
	req := httptest.NewRequest(http.MethodGet, "/api/users/notifications?page=1&limit=10", nil)
	req.Header.Set("X-User-Id", userID)

	rec := httptest.NewRecorder()
	handler := middleware.AuthIdentity(http.HandlerFunc(h.ListNotifications))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	var resp service.NotificationListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.TotalCount != 1 || len(resp.Items) != 1 {
		t.Errorf("expected 1 item in list response, got %+v", resp)
	}

	// 2. Unauthorized without header
	unauthReq := httptest.NewRequest(http.MethodGet, "/api/users/notifications", nil)
	unauthRec := httptest.NewRecorder()
	handler.ServeHTTP(unauthRec, unauthReq)

	if unauthRec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got: %d", unauthRec.Code)
	}
}

func TestNotificationHandler_GetNotification(t *testing.T) {
	h, repo := setupTestNotificationHandler()
	userID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	n, _ := repo.Create(context.Background(), repository.NotificationRecord{
		UserID:  userID,
		Title:   "Flash Sale",
		Message: "Sale ends tonight.",
		Type:    "PROMOTION",
	})

	// 1. Success lookup via Chi router context
	r := chi.NewRouter()
	r.Use(middleware.AuthIdentity)
	r.Get("/api/users/notifications/{id}", h.GetNotification)

	req := httptest.NewRequest(http.MethodGet, "/api/users/notifications/"+n.ID, nil)
	req.Header.Set("X-User-Id", userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	// 2. Not found lookup
	reqNF := httptest.NewRequest(http.MethodGet, "/api/users/notifications/non-existent-id", nil)
	reqNF.Header.Set("X-User-Id", userID)
	recNF := httptest.NewRecorder()
	r.ServeHTTP(recNF, reqNF)

	if recNF.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got: %d", recNF.Code)
	}
}

func TestNotificationHandler_MarkAsRead_And_MarkAllAsRead(t *testing.T) {
	h, repo := setupTestNotificationHandler()
	userID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	n1, _ := repo.Create(context.Background(), repository.NotificationRecord{UserID: userID, Title: "N1", Message: "M1"})
	_, _ = repo.Create(context.Background(), repository.NotificationRecord{UserID: userID, Title: "N2", Message: "M2"})

	r := chi.NewRouter()
	r.Use(middleware.AuthIdentity)
	r.Patch("/api/users/notifications/{id}/read", h.MarkAsRead)
	r.Post("/api/users/notifications/read-all", h.MarkAllAsRead)

	// 1. Mark single read
	req1 := httptest.NewRequest(http.MethodPatch, "/api/users/notifications/"+n1.ID+"/read", nil)
	req1.Header.Set("X-User-Id", userID)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec1.Code)
	}

	var dto service.NotificationDTO
	_ = json.NewDecoder(rec1.Body).Decode(&dto)
	if !dto.IsRead {
		t.Errorf("expected notification to be marked read")
	}

	// 2. Mark all read
	reqAll := httptest.NewRequest(http.MethodPost, "/api/users/notifications/read-all", nil)
	reqAll.Header.Set("X-User-Id", userID)
	recAll := httptest.NewRecorder()
	r.ServeHTTP(recAll, reqAll)

	if recAll.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", recAll.Code)
	}

	var batchResp service.MarkAllReadResponse
	_ = json.NewDecoder(recAll.Body).Decode(&batchResp)
	if batchResp.UpdatedCount != 1 {
		t.Errorf("expected 1 remaining notification updated, got %d", batchResp.UpdatedCount)
	}
}

func TestNotificationHandler_DeleteNotification(t *testing.T) {
	h, repo := setupTestNotificationHandler()
	userID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	n, _ := repo.Create(context.Background(), repository.NotificationRecord{UserID: userID, Title: "Del", Message: "Del"})

	r := chi.NewRouter()
	r.Use(middleware.AuthIdentity)
	r.Delete("/api/users/notifications/{id}", h.DeleteNotification)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/notifications/"+n.ID, nil)
	req.Header.Set("X-User-Id", userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	// Deleting again returns 404
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got: %d", rec2.Code)
	}
}

func TestNotificationHandler_Preferences(t *testing.T) {
	h, _ := setupTestNotificationHandler()
	userID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	r := chi.NewRouter()
	r.Use(middleware.AuthIdentity)
	r.Get("/api/users/notifications/preferences", h.GetPreferences)
	r.Put("/api/users/notifications/preferences", h.UpdatePreferences)

	// 1. Get preferences
	reqGet := httptest.NewRequest(http.MethodGet, "/api/users/notifications/preferences", nil)
	reqGet.Header.Set("X-User-Id", userID)
	recGet := httptest.NewRecorder()
	r.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", recGet.Code)
	}

	// 2. Update preferences
	body := `{"emailEnabled":false,"smsEnabled":true}`
	reqPut := httptest.NewRequest(http.MethodPut, "/api/users/notifications/preferences", bytes.NewBufferString(body))
	reqPut.Header.Set("X-User-Id", userID)
	recPut := httptest.NewRecorder()
	r.ServeHTTP(recPut, reqPut)

	if recPut.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", recPut.Code)
	}

	var pref service.NotificationPreferencesDTO
	_ = json.NewDecoder(recPut.Body).Decode(&pref)
	if pref.EmailEnabled || !pref.SMSEnabled {
		t.Errorf("unexpected updated preferences: %+v", pref)
	}

	// 3. Bad JSON
	reqBad := httptest.NewRequest(http.MethodPut, "/api/users/notifications/preferences", bytes.NewBufferString("invalid json"))
	reqBad.Header.Set("X-User-Id", userID)
	recBad := httptest.NewRecorder()
	r.ServeHTTP(recBad, reqBad)

	if recBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for malformed JSON, got: %d", recBad.Code)
	}
}
