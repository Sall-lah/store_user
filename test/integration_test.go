package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	orderv1 "github.com/Sall-lah/store_proto/gen/go/store/order/v1"
	"github.com/Sall-lah/store_user/internal/client/order"
	"github.com/Sall-lah/store_user/internal/config"
	"github.com/Sall-lah/store_user/internal/handler"
	"github.com/Sall-lah/store_user/internal/kafka"
	"github.com/Sall-lah/store_user/internal/repository"
	"github.com/Sall-lah/store_user/internal/router"
	"github.com/Sall-lah/store_user/internal/service"
)

type testEnvironment struct {
	server      http.Handler
	repo        *repository.MockUserProfileRepository
	orderClient *order.MockClient
	kafkaMock   *kafka.MockProducer
}

func setupTestEnvironment() *testEnvironment {
	cfg := &config.Config{
		ServerPort:                 "8082",
		Env:                        "test",
		RateLimitMaxRequests:       60,
		RateLimitWindow:            time.Minute,
		RateLimitDeleteMaxRequests: 3,
		RateLimitDeleteWindow:      time.Minute,
		KafkaTopicUserEvents:       "user.events",
	}

	repo := repository.NewMockUserProfileRepository()
	orderClient := &order.MockClient{}
	kafkaMock := &kafka.MockProducer{}

	svc := service.NewUserService(repo, nil, orderClient, kafkaMock, cfg.KafkaTopicUserEvents)
	h := handler.NewProfileHandler(svc)
	adminH := handler.NewAdminHandler(svc)
	r := router.NewRouter(cfg, h, adminH, nil, nil, nil)

	return &testEnvironment{
		server:      r,
		repo:        repo,
		orderClient: orderClient,
		kafkaMock:   kafkaMock,
	}
}

func TestIntegration_UserProfileLifecycleAndAccountDeletion(t *testing.T) {
	env := setupTestEnvironment()
	userID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	// 1. Initial User Creation via REST POST /api/users -> Expect 201 Created
	createPayload := `{"fullName":"Budi Initial","phoneNumber":"+628123456789","address":"Jl. Sudirman"}`
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(createPayload))
	reqCreate.Header.Set("X-User-Id", userID)
	recCreate := httptest.NewRecorder()
	env.server.ServeHTTP(recCreate, reqCreate)

	if recCreate.Code != http.StatusCreated {
		t.Fatalf("Step 1 failed: expected 201 Created, got: %d (body: %s)", recCreate.Code, recCreate.Body.String())
	}

	var pCreated repository.UserProfile
	if err := json.NewDecoder(recCreate.Body).Decode(&pCreated); err != nil {
		t.Fatalf("Step 1 failed: decode error %v", err)
	}
	if pCreated.UserID != userID || pCreated.FullName != "Budi Initial" {
		t.Errorf("Step 1 failed: unexpected created profile: %+v", pCreated)
	}

	// 2. Idempotent Replay via POST /api/users -> Expect 200 OK
	reqReplay := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(createPayload))
	reqReplay.Header.Set("X-User-Id", userID)
	recReplay := httptest.NewRecorder()
	env.server.ServeHTTP(recReplay, reqReplay)

	if recReplay.Code != http.StatusOK {
		t.Fatalf("Step 2 failed: expected 200 OK for idempotent creation, got: %d", recReplay.Code)
	}

	// 3. Profile Read GET /api/users/profile
	reqRead := httptest.NewRequest(http.MethodGet, "/api/users/profile", nil)
	reqRead.Header.Set("X-User-Id", userID)
	recRead := httptest.NewRecorder()
	env.server.ServeHTTP(recRead, reqRead)

	if recRead.Code != http.StatusOK {
		t.Fatalf("Step 3 failed: expected 200 OK, got: %d", recRead.Code)
	}

	// 4. Profile Update PUT /api/users/profile
	updatePayload := `{"fullName":"Budi Pratama","phoneNumber":"+628123456789","address":"Jakarta, Indonesia"}`
	req2 := httptest.NewRequest(http.MethodPut, "/api/users/profile", bytes.NewBufferString(updatePayload))
	req2.Header.Set("X-User-Id", userID)
	rec2 := httptest.NewRecorder()
	env.server.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Step 4 failed: expected 200 OK, got: %d", rec2.Code)
	}

	var p2 repository.UserProfile
	if err := json.NewDecoder(rec2.Body).Decode(&p2); err != nil {
		t.Fatalf("Step 4 failed: decode error %v", err)
	}
	if p2.FullName != "Budi Pratama" || *p2.PhoneNumber != "+628123456789" || *p2.Address != "Jakarta, Indonesia" {
		t.Errorf("Step 4 failed: unexpected updated data: %+v", p2)
	}

	// 5. Attempt Account Deletion when User has Active Orders in store_order -> Expect 409 Conflict
	env.orderClient.CheckActiveOrdersFunc = func(ctx context.Context, uid string) (*orderv1.CheckActiveOrdersResponse, error) {
		return &orderv1.CheckActiveOrdersResponse{
			HasActiveOrders:  true,
			ActiveOrderCount: 1,
			ActiveOrders: []*orderv1.ActiveOrderSummary{
				{
					OrderId:     "ord-active-99",
					OrderNumber: "ORD-2026-99",
					Status:      orderv1.OrderStatus_ORDER_STATUS_PROCESSING,
					TotalAmount: 250000,
				},
			},
			Message: "Active order ord-active-99 in PROCESSING state",
		}, nil
	}

	req3 := httptest.NewRequest(http.MethodDelete, "/api/users/account", bytes.NewBufferString(`{"reason":"deleting"}`))
	req3.Header.Set("X-User-Id", userID)
	rec3 := httptest.NewRecorder()
	env.server.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusConflict {
		t.Fatalf("Step 5 failed: expected 409 Conflict, got: %d (body: %s)", rec3.Code, rec3.Body.String())
	}

	// Profile record must still exist in DB
	pCheck, err := env.repo.GetByUserID(context.Background(), userID)
	if err != nil || pCheck == nil {
		t.Fatalf("Step 5 failed: profile was improperly deleted on conflict!")
	}
	if len(env.kafkaMock.PublishedMessages) != 0 {
		t.Fatalf("Step 5 failed: kafka event was improperly published on conflict!")
	}

	// 6. Attempt Account Deletion when Order Service gRPC is down -> Expect 503 Service Unavailable
	env.orderClient.CheckActiveOrdersFunc = func(ctx context.Context, uid string) (*orderv1.CheckActiveOrdersResponse, error) {
		return nil, errors.New("rpc error: code = Unavailable desc = connection refused")
	}

	req4 := httptest.NewRequest(http.MethodDelete, "/api/users/account", nil)
	req4.Header.Set("X-User-Id", userID)
	rec4 := httptest.NewRecorder()
	env.server.ServeHTTP(rec4, req4)

	if rec4.Code != http.StatusServiceUnavailable {
		t.Fatalf("Step 6 failed: expected 503 Service Unavailable, got: %d", rec4.Code)
	}

	// 7. Successful Account Deletion (Active orders cleared) -> Expect 200 OK, DB deleted, Kafka emitted
	env.orderClient.CheckActiveOrdersFunc = func(ctx context.Context, uid string) (*orderv1.CheckActiveOrdersResponse, error) {
		return &orderv1.CheckActiveOrdersResponse{
			HasActiveOrders:  false,
			ActiveOrderCount: 0,
			ActiveOrders:     []*orderv1.ActiveOrderSummary{},
		}, nil
	}

	req5 := httptest.NewRequest(http.MethodDelete, "/api/users/account", bytes.NewBufferString(`{"reason":"moving to another platform"}`))
	req5.Header.Set("X-User-Id", userID)
	rec5 := httptest.NewRecorder()
	env.server.ServeHTTP(rec5, req5)

	if rec5.Code != http.StatusOK {
		t.Fatalf("Step 7 failed: expected 200 OK, got: %d (body: %s)", rec5.Code, rec5.Body.String())
	}

	// Verify hard-deletion from DB
	_, err = env.repo.GetByUserID(context.Background(), userID)
	if err != repository.ErrNotFound {
		t.Errorf("Step 7 failed: profile record still exists in DB: %v", err)
	}

	// Verify domain event emitted to Kafka topic 'user.events'
	if len(env.kafkaMock.PublishedMessages) != 1 {
		t.Fatalf("Step 7 failed: expected 1 published Kafka message, got: %d", len(env.kafkaMock.PublishedMessages))
	}
	msg := env.kafkaMock.PublishedMessages[0]
	if msg.Topic != "user.events" {
		t.Errorf("Step 7 failed: expected topic user.events, got: %s", msg.Topic)
	}
	if msg.Key != userID {
		t.Errorf("Step 7 failed: expected message key %s, got: %s", userID, msg.Key)
	}

	var event kafka.LifecycleEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		t.Fatalf("Step 7 failed: invalid kafka event JSON: %v", err)
	}
	if event.Event != "user.deleted" || event.UserID != userID || event.Reason != "moving to another platform" {
		t.Errorf("Step 7 failed: unexpected event payload: %+v", event)
	}
}

func TestIntegration_AdminForcedAccountDeletion(t *testing.T) {
	env := setupTestEnvironment()
	adminID := "00000000-1111-2222-3333-444444444444"
	targetUserID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	// 1. Create target profile in DB
	name := "Alice Target"
	_, err := env.repo.Upsert(context.Background(), targetUserID, repository.UpdateProfileParams{FullName: &name})
	if err != nil {
		t.Fatalf("Failed to seed profile: %v", err)
	}

	// 2. Mock orderClient to report active orders (which would block customer deletion)
	env.orderClient.CheckActiveOrdersFunc = func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
		t.Fatal("Order pre-flight check should NOT be called for admin deletion")
		return &orderv1.CheckActiveOrdersResponse{
			HasActiveOrders:  true,
			ActiveOrderCount: 5,
		}, nil
	}

	// 3. Attempt deletion as customer -> 403 Forbidden
	reqCust := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+targetUserID, nil)
	reqCust.Header.Set("X-User-Id", targetUserID)
	reqCust.Header.Set("X-User-Role", "CUSTOMER")
	recCust := httptest.NewRecorder()
	env.server.ServeHTTP(recCust, reqCust)

	if recCust.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for customer, got: %d", recCust.Code)
	}

	// 4. Perform deletion as ADMIN -> 200 OK
	reqAdmin := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+targetUserID, nil)
	reqAdmin.Header.Set("X-User-Id", adminID)
	reqAdmin.Header.Set("X-User-Role", "ADMIN")
	recAdmin := httptest.NewRecorder()
	env.server.ServeHTTP(recAdmin, reqAdmin)

	if recAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for admin delete, got: %d (body: %s)", recAdmin.Code, recAdmin.Body.String())
	}

	// 5. Verify profile deleted from DB
	_, err = env.repo.GetByUserID(context.Background(), targetUserID)
	if err != repository.ErrNotFound {
		t.Errorf("expected profile to be deleted from DB, got: %v", err)
	}

	// 6. Verify Kafka event dispatched with reason admin_deletion
	if len(env.kafkaMock.PublishedMessages) != 1 {
		t.Fatalf("expected 1 published kafka message, got: %d", len(env.kafkaMock.PublishedMessages))
	}
	msg := env.kafkaMock.PublishedMessages[0]
	if msg.Key != targetUserID {
		t.Errorf("expected message key %s, got: %s", targetUserID, msg.Key)
	}
	var event kafka.LifecycleEvent
	_ = json.Unmarshal(msg.Payload, &event)
	if event.Reason != "admin_deletion" {
		t.Errorf("expected reason admin_deletion, got: %s", event.Reason)
	}
}

