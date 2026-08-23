package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	orderv1 "github.com/Sall-lah/store_proto/gen/go/store/order/v1"
	"github.com/Sall-lah/store_user/internal/client/order"
	"github.com/Sall-lah/store_user/internal/kafka"
	"github.com/Sall-lah/store_user/internal/middleware"
	"github.com/Sall-lah/store_user/internal/repository"
	"github.com/Sall-lah/store_user/internal/service"
)

func setupTestHandler() (*ProfileHandler, *repository.MockUserProfileRepository, *order.MockClient, *kafka.MockProducer) {
	repo := repository.NewMockUserProfileRepository()
	orderMock := &order.MockClient{}
	kafkaMock := &kafka.MockProducer{}
	svc := service.NewUserService(repo, orderMock, kafkaMock, "user.events")
	return NewProfileHandler(svc), repo, orderMock, kafkaMock
}

func TestHandler_GetProfile(t *testing.T) {
	h, repo, _, _ := setupTestHandler()
	name := "Alice"
	_, _ = repo.Upsert(context.Background(), "usr-123", repository.UpdateProfileParams{FullName: &name})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/profile", nil)
	req.Header.Set("X-User-Id", "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d")

	rec := httptest.NewRecorder()
	handler := middleware.AuthIdentity(http.HandlerFunc(h.GetProfile))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	var resp repository.UserProfile
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestHandler_UpdateProfile(t *testing.T) {
	h, _, _, _ := setupTestHandler()

	body := `{"fullName":"Bob New","phoneNumber":"+628123456789","bio":"<b>Cool Bio</b>"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/profile", bytes.NewBufferString(body))
	req.Header.Set("X-User-Id", "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d")

	rec := httptest.NewRecorder()
	handler := middleware.AuthIdentity(http.HandlerFunc(h.UpdateProfile))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", rec.Code)
	}

	var resp repository.UserProfile
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.FullName != "Bob New" {
		t.Errorf("expected fullName Bob New, got: %s", resp.FullName)
	}
	if *resp.Bio != "Cool Bio" {
		t.Errorf("expected sanitized bio 'Cool Bio', got: %s", *resp.Bio)
	}
}

func TestHandler_UpdateProfile_InvalidPhone(t *testing.T) {
	h, _, _, _ := setupTestHandler()

	body := `{"phoneNumber":"invalid-phone"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/profile", bytes.NewBufferString(body))
	req.Header.Set("X-User-Id", "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d")

	rec := httptest.NewRecorder()
	handler := middleware.AuthIdentity(http.HandlerFunc(h.UpdateProfile))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got: %d", rec.Code)
	}
}

func TestHandler_DeleteAccount_Success(t *testing.T) {
	h, repo, _, kafkaMock := setupTestHandler()
	name := "Alice"
	_, _ = repo.Upsert(context.Background(), "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d", repository.UpdateProfileParams{FullName: &name})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/account", bytes.NewBufferString(`{"reason":"leaving"}`))
	req.Header.Set("X-User-Id", "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d")

	rec := httptest.NewRecorder()
	handler := middleware.AuthIdentity(http.HandlerFunc(h.DeleteAccount))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d (body: %s)", rec.Code, rec.Body.String())
	}

	if len(kafkaMock.PublishedMessages) != 1 {
		t.Errorf("expected 1 kafka message, got: %d", len(kafkaMock.PublishedMessages))
	}
}

func TestHandler_DeleteAccount_Conflict(t *testing.T) {
	h, repo, orderMock, _ := setupTestHandler()
	name := "Alice"
	_, _ = repo.Upsert(context.Background(), "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d", repository.UpdateProfileParams{FullName: &name})

	orderMock.CheckActiveOrdersFunc = func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
		return &orderv1.CheckActiveOrdersResponse{
			HasActiveOrders:  true,
			ActiveOrderCount: 1,
			ActiveOrders: []*orderv1.ActiveOrderSummary{
				{OrderId: "ord-1", Status: orderv1.OrderStatus_ORDER_STATUS_SHIPPED},
			},
			Message: "Active order blocking deletion",
		}, nil
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/account", nil)
	req.Header.Set("X-User-Id", "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d")

	rec := httptest.NewRecorder()
	handler := middleware.AuthIdentity(http.HandlerFunc(h.DeleteAccount))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got: %d", rec.Code)
	}
}

func TestHandler_HealthCheck(t *testing.T) {
	h, _, _, _ := setupTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.HealthCheck(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got: %d", rec.Code)
	}
}
