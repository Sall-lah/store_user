package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Sall-lah/store_user/internal/client/order"
	"github.com/Sall-lah/store_user/internal/kafka"
	"github.com/Sall-lah/store_user/internal/repository"
	orderv1 "github.com/Sall-lah/store_proto/gen/go/store/order/v1"
)

func TestGetProfile_AutoCreateForNewUser(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	svc := NewUserService(repo, nil, nil, "user.events")
	ctx := context.Background()

	p, err := svc.GetProfile(ctx, "usr_new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.UserID != "usr_new" || p.FullName != "User" {
		t.Errorf("expected auto-initialized profile, got: %+v", p)
	}
}

func TestCreateUserProfile_SuccessAndIdempotent(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	svc := NewUserService(repo, nil, nil, "user.events")
	ctx := context.Background()

	phone := "+628123456789"
	req := CreateProfileRequest{
		UserID:      "usr_100",
		FullName:    "Charlie",
		Email:       "charlie@example.com",
		PhoneNumber: &phone,
	}

	// 1. Initial Creation (is_created = true)
	res1, err := svc.CreateUserProfile(ctx, req)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if !res1.IsCreated {
		t.Errorf("expected is_created to be true for new user, got false")
	}
	if res1.Profile == nil || res1.Profile.FullName != "Charlie" {
		t.Errorf("unexpected profile result: %+v", res1.Profile)
	}

	// 2. Idempotent Replay (is_created = false, returns existing record)
	res2, err := svc.CreateUserProfile(ctx, req)
	if err != nil {
		t.Fatalf("unexpected replay error: %v", err)
	}
	if res2.IsCreated {
		t.Errorf("expected is_created to be false on replay, got true")
	}
	if res2.Profile.ID != res1.Profile.ID {
		t.Errorf("expected same profile ID on replay, got %s vs %s", res2.Profile.ID, res1.Profile.ID)
	}
}

func TestCreateUserProfile_Validation(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	svc := NewUserService(repo, nil, nil, "user.events")
	ctx := context.Background()

	// Empty userID
	_, err := svc.CreateUserProfile(ctx, CreateProfileRequest{
		UserID:   "",
		FullName: "Charlie",
	})
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got: %v", err)
	}

	// Empty fullName
	_, err = svc.CreateUserProfile(ctx, CreateProfileRequest{
		UserID:   "usr_100",
		FullName: "   ",
	})
	if !errors.Is(err, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed, got: %v", err)
	}
}

func TestUpdateProfile_Validation(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	svc := NewUserService(repo, nil, nil, "user.events")
	ctx := context.Background()

	// 1. Empty full name
	emptyName := ""
	_, err := svc.UpdateProfile(ctx, "usr_1", UpdateProfileRequest{FullName: &emptyName})
	if !errors.Is(err, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed for empty full name, got: %v", err)
	}

	// 2. Valid update
	name := "Bob"
	addr := "456 Market St"
	p, err := svc.UpdateProfile(ctx, "usr_1", UpdateProfileRequest{
		FullName: &name,
		Address:  &addr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.FullName != "Bob" || *p.Address != "456 Market St" {
		t.Errorf("unexpected profile: %+v", p)
	}
}

func TestDeleteAccount_Success(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	name := "Alice"
	_, _ = repo.Upsert(context.Background(), "usr_1", repository.UpdateProfileParams{FullName: &name})

	orderClient := &order.MockClient{
		CheckActiveOrdersFunc: func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
			return &orderv1.CheckActiveOrdersResponse{
				HasActiveOrders:  false,
				ActiveOrderCount: 0,
			}, nil
		},
	}

	kafkaMock := &kafka.MockProducer{}
	svc := NewUserService(repo, orderClient, kafkaMock, "user.events")

	err := svc.DeleteAccount(context.Background(), "usr_1", DeleteAccountRequest{Reason: "user_leaving"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check DB hard deletion
	_, err = repo.GetByUserID(context.Background(), "usr_1")
	if err != repository.ErrNotFound {
		t.Errorf("expected profile to be deleted from DB, got: %v", err)
	}

	// Check Kafka event
	if len(kafkaMock.PublishedMessages) != 1 {
		t.Fatalf("expected 1 published kafka message, got: %d", len(kafkaMock.PublishedMessages))
	}
	if kafkaMock.PublishedMessages[0].Key != "usr_1" {
		t.Errorf("expected key usr_1, got: %s", kafkaMock.PublishedMessages[0].Key)
	}
}

func TestDeleteAccount_BlockedByActiveOrders(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	name := "Alice"
	_, _ = repo.Upsert(context.Background(), "usr_1", repository.UpdateProfileParams{FullName: &name})

	orderClient := &order.MockClient{
		CheckActiveOrdersFunc: func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
			return &orderv1.CheckActiveOrdersResponse{
				HasActiveOrders:  true,
				ActiveOrderCount: 2,
				ActiveOrders: []*orderv1.ActiveOrderSummary{
					{OrderId: "ord-1", Status: orderv1.OrderStatus_ORDER_STATUS_PROCESSING},
					{OrderId: "ord-2", Status: orderv1.OrderStatus_ORDER_STATUS_SHIPPED},
				},
				Message: "User has 2 active orders",
			}, nil
		},
	}

	kafkaMock := &kafka.MockProducer{}
	svc := NewUserService(repo, orderClient, kafkaMock, "user.events")

	err := svc.DeleteAccount(context.Background(), "usr_1", DeleteAccountRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var conflictErr *ActiveOrdersConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("expected ActiveOrdersConflictError, got: %v", err)
	}
	if conflictErr.Count != 2 {
		t.Errorf("expected 2 active orders, got: %d", conflictErr.Count)
	}

	// Profile must NOT be deleted
	p, err := repo.GetByUserID(context.Background(), "usr_1")
	if err != nil || p == nil {
		t.Errorf("expected profile to remain in DB, got: %v", err)
	}

	// Kafka event must NOT be emitted
	if len(kafkaMock.PublishedMessages) != 0 {
		t.Errorf("expected 0 published kafka messages, got: %d", len(kafkaMock.PublishedMessages))
	}
}

func TestDeleteAccount_OrderServiceUnavailable(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	name := "Alice"
	_, _ = repo.Upsert(context.Background(), "usr_1", repository.UpdateProfileParams{FullName: &name})

	orderClient := &order.MockClient{
		CheckActiveOrdersFunc: func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
			return nil, errors.New("connection refused")
		},
	}

	kafkaMock := &kafka.MockProducer{}
	svc := NewUserService(repo, orderClient, kafkaMock, "user.events")

	err := svc.DeleteAccount(context.Background(), "usr_1", DeleteAccountRequest{})
	if !errors.Is(err, ErrOrderServiceUnavailable) {
		t.Fatalf("expected ErrOrderServiceUnavailable, got: %v", err)
	}

	// Profile must NOT be deleted
	p, err := repo.GetByUserID(context.Background(), "usr_1")
	if err != nil || p == nil {
		t.Errorf("expected profile to remain in DB, got: %v", err)
	}
}
