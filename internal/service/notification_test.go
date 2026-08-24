package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sall-lah/store_user/internal/repository"
)

func TestListNotifications_ValidationAndPagination(t *testing.T) {
	repo := repository.NewMockNotificationRepository()
	svc := NewNotificationService(repo)
	ctx := context.Background()
	userID := "usr_test_1"

	// 1. Validation - empty userID
	_, err := svc.ListNotifications(ctx, "", 1, 10, nil)
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got: %v", err)
	}

	// 2. Populate 5 notifications
	for i := 1; i <= 5; i++ {
		_, _ = repo.Create(ctx, repository.NotificationRecord{
			UserID:    userID,
			Title:     "Title",
			Message:   "Message",
			Type:      "ORDER_UPDATE",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		})
	}

	// 3. Paginated query (page 1, limit 2)
	res, err := svc.ListNotifications(ctx, userID, 1, 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalCount != 5 {
		t.Errorf("expected TotalCount 5, got %d", res.TotalCount)
	}
	if res.TotalPages != 3 {
		t.Errorf("expected TotalPages 3, got %d", res.TotalPages)
	}
	if len(res.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(res.Items))
	}
	if res.Page != 1 || res.Limit != 2 {
		t.Errorf("unexpected page metadata: %+v", res)
	}

	// 4. Default bounds (page <= 0, limit <= 0)
	resDefault, err := svc.ListNotifications(ctx, userID, 0, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resDefault.Page != 1 || resDefault.Limit != 20 {
		t.Errorf("expected default page 1 limit 20, got page %d limit %d", resDefault.Page, resDefault.Limit)
	}
}

func TestGetNotification_Scenarios(t *testing.T) {
	repo := repository.NewMockNotificationRepository()
	svc := NewNotificationService(repo)
	ctx := context.Background()
	userID := "usr_test_2"

	n, _ := repo.Create(ctx, repository.NotificationRecord{
		UserID:  userID,
		Title:   "Security Alert",
		Message: "New login detected.",
		Type:    "SECURITY",
	})

	// 1. Valid retrieval
	dto, err := svc.GetNotification(ctx, userID, n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.Title != "Security Alert" {
		t.Errorf("expected 'Security Alert', got %s", dto.Title)
	}

	// 2. Empty ID
	_, err = svc.GetNotification(ctx, userID, " ")
	if !errors.Is(err, ErrInvalidNotificationID) {
		t.Errorf("expected ErrInvalidNotificationID, got: %v", err)
	}

	// 3. Foreign user lookup
	_, err = svc.GetNotification(ctx, "usr_other", n.ID)
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("expected ErrNotificationNotFound, got: %v", err)
	}
}

func TestMarkAsRead_And_MarkAllAsRead(t *testing.T) {
	repo := repository.NewMockNotificationRepository()
	svc := NewNotificationService(repo)
	ctx := context.Background()
	userID := "usr_test_3"

	n1, _ := repo.Create(ctx, repository.NotificationRecord{UserID: userID, Title: "N1", Message: "M1"})
	_, _ = repo.Create(ctx, repository.NotificationRecord{UserID: userID, Title: "N2", Message: "M2"})

	// 1. Mark single as read
	dto, err := svc.MarkAsRead(ctx, userID, n1.ID)
	if err != nil {
		t.Fatalf("unexpected mark read error: %v", err)
	}
	if !dto.IsRead || dto.ReadAt == nil {
		t.Errorf("expected notification to be marked as read")
	}

	// 2. Mark non-existent
	_, err = svc.MarkAsRead(ctx, userID, "not-found-id")
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("expected ErrNotificationNotFound, got: %v", err)
	}

	// 3. Mark all as read
	batchRes, err := svc.MarkAllAsRead(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected mark all read error: %v", err)
	}
	if batchRes.UpdatedCount != 1 {
		t.Errorf("expected 1 remaining updated record, got %d", batchRes.UpdatedCount)
	}
}

func TestDeleteNotification(t *testing.T) {
	repo := repository.NewMockNotificationRepository()
	svc := NewNotificationService(repo)
	ctx := context.Background()
	userID := "usr_test_4"

	n, _ := repo.Create(ctx, repository.NotificationRecord{UserID: userID, Title: "ToDelete", Message: "ToDelete"})

	// 1. Invalid input
	if err := svc.DeleteNotification(ctx, "", n.ID); !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got: %v", err)
	}
	if err := svc.DeleteNotification(ctx, userID, ""); !errors.Is(err, ErrInvalidNotificationID) {
		t.Errorf("expected ErrInvalidNotificationID, got: %v", err)
	}

	// 2. Delete valid
	if err := svc.DeleteNotification(ctx, userID, n.ID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	// 3. Delete again returns not found
	if err := svc.DeleteNotification(ctx, userID, n.ID); !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("expected ErrNotificationNotFound after deletion, got %v", err)
	}
}

