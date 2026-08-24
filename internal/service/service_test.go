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
	svc := NewUserService(repo, nil, nil, nil, "user.events")
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
	svc := NewUserService(repo, nil, nil, nil, "user.events")
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
	svc := NewUserService(repo, nil, nil, nil, "user.events")
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
	svc := NewUserService(repo, nil, nil, nil, "user.events")
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
	svc := NewUserService(repo, nil, orderClient, kafkaMock, "user.events")

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
	svc := NewUserService(repo, nil, orderClient, kafkaMock, "user.events")

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
	svc := NewUserService(repo, nil, orderClient, kafkaMock, "user.events")

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

func TestAdminDeleteUser_Success_BypassesActiveOrders(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	notifRepo := repository.NewMockNotificationRepository()
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	name := "Target User"
	_, _ = repo.Upsert(context.Background(), validUUID, repository.UpdateProfileParams{FullName: &name})
	_, _ = notifRepo.Create(context.Background(), repository.NotificationRecord{
		UserID:  validUUID,
		Title:   "Alert",
		Message: "Important message",
	})

	// Order service client that would block regular customer deletion
	orderClient := &order.MockClient{
		CheckActiveOrdersFunc: func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
			t.Fatal("CheckActiveOrders MUST NOT be called during admin user deletion")
			return nil, nil
		},
	}

	kafkaMock := &kafka.MockProducer{}
	svc := NewUserService(repo, notifRepo, orderClient, kafkaMock, "user.events")

	err := svc.AdminDeleteUser(context.Background(), validUUID)
	if err != nil {
		t.Fatalf("unexpected error during admin user delete: %v", err)
	}

	// 1. Profile must be deleted from DB
	_, err = repo.GetByUserID(context.Background(), validUUID)
	if err != repository.ErrNotFound {
		t.Errorf("expected profile to be deleted, got err: %v", err)
	}

	// 2. Notifications must be deleted
	notifs, count, _, _ := notifRepo.List(context.Background(), repository.ListNotificationsParams{UserID: validUUID})
	if count != 0 || len(notifs) != 0 {
		t.Errorf("expected 0 notifications remaining, got: %d", count)
	}

	// 3. Kafka event must be emitted with reason "admin_deletion"
	if len(kafkaMock.PublishedMessages) != 1 {
		t.Fatalf("expected 1 published kafka message, got: %d", len(kafkaMock.PublishedMessages))
	}
	if kafkaMock.PublishedMessages[0].Key != validUUID {
		t.Errorf("expected kafka key %s, got %s", validUUID, kafkaMock.PublishedMessages[0].Key)
	}
}

func TestAdminDeleteUser_Idempotent_NonExistentProfile(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	validUUID := "a1111111-2222-3333-4444-555555555555"

	kafkaMock := &kafka.MockProducer{}
	svc := NewUserService(repo, nil, nil, kafkaMock, "user.events")

	// Target profile does not exist in store_user DB
	err := svc.AdminDeleteUser(context.Background(), validUUID)
	if err != nil {
		t.Fatalf("expected nil for idempotent non-existent profile deletion, got: %v", err)
	}

	// Kafka event must still be dispatched so store_auth cleans up credentials
	if len(kafkaMock.PublishedMessages) != 1 {
		t.Fatalf("expected 1 published kafka message, got: %d", len(kafkaMock.PublishedMessages))
	}
}

func TestAdminDeleteUser_InvalidUUID(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	svc := NewUserService(repo, nil, nil, nil, "user.events")

	err := svc.AdminDeleteUser(context.Background(), "invalid-non-uuid-string")
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got: %v", err)
	}
}

func TestAdminDeleteUser_KafkaFailure(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	kafkaMock := &kafka.MockProducer{
		PublishError: errors.New("kafka broker down"),
	}
	svc := NewUserService(repo, nil, nil, kafkaMock, "user.events")

	err := svc.AdminDeleteUser(context.Background(), validUUID)
	if !errors.Is(err, ErrKafkaPublishFailed) {
		t.Fatalf("expected ErrKafkaPublishFailed, got: %v", err)
	}
}

func TestAdminBanUser_Success_PreservesProfile(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	// Seed profile to ensure it is preserved
	_, _ = repo.Create(context.Background(), validUUID, "Alice Fraud", nil, nil)

	kafkaMock := &kafka.MockProducer{}
	svc := NewUserService(repo, nil, nil, kafkaMock, "user.events")

	err := svc.AdminBanUser(context.Background(), validUUID, BanUserRequest{Reason: "payment_chargeback"})
	if err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}

	// Profile MUST still exist in repository (not deleted)
	profile, err := repo.GetByUserID(context.Background(), validUUID)
	if err != nil || profile == nil {
		t.Errorf("expected profile to be preserved in repo after ban, got err: %v", err)
	}

	// Kafka event must be dispatched with user.banned
	if len(kafkaMock.PublishedMessages) != 1 {
		t.Fatalf("expected 1 published kafka message, got: %d", len(kafkaMock.PublishedMessages))
	}
	if kafkaMock.PublishedMessages[0].Key != validUUID {
		t.Errorf("expected key '%s', got '%s'", validUUID, kafkaMock.PublishedMessages[0].Key)
	}
}

func TestAdminBanUser_DefaultReason(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	kafkaMock := &kafka.MockProducer{}
	svc := NewUserService(repo, nil, nil, kafkaMock, "user.events")

	err := svc.AdminBanUser(context.Background(), validUUID, BanUserRequest{Reason: ""})
	if err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}

	if len(kafkaMock.PublishedMessages) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(kafkaMock.PublishedMessages))
	}
}

func TestAdminBanUser_InvalidUUID(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	svc := NewUserService(repo, nil, nil, nil, "user.events")

	err := svc.AdminBanUser(context.Background(), "not-a-valid-uuid", BanUserRequest{Reason: "test"})
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got: %v", err)
	}
}

func TestAdminBanUser_KafkaFailure(t *testing.T) {
	repo := repository.NewMockUserProfileRepository()
	validUUID := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"

	kafkaMock := &kafka.MockProducer{
		PublishError: errors.New("kafka broker unreachable"),
	}
	svc := NewUserService(repo, nil, nil, kafkaMock, "user.events")

	err := svc.AdminBanUser(context.Background(), validUUID, BanUserRequest{Reason: "fraud"})
	if !errors.Is(err, ErrKafkaPublishFailed) {
		t.Fatalf("expected ErrKafkaPublishFailed, got: %v", err)
	}
}

