package repository

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockNotificationRepository_Lifecycle(t *testing.T) {
	repo := NewMockNotificationRepository()
	ctx := context.Background()
	userID := "usr_100"
	otherUser := "usr_200"

	// 1. Create sample notifications
	n1, err := repo.Create(ctx, NotificationRecord{
		UserID:    userID,
		Title:     "Order Confirmed",
		Message:   "Your order #1001 has been placed.",
		Type:      "ORDER_UPDATE",
		CreatedAt: time.Now().Add(-10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("unexpected error creating notification: %v", err)
	}

	n2, err := repo.Create(ctx, NotificationRecord{
		UserID:    userID,
		Title:     "Special Discount",
		Message:   "Get 20% off your next purchase.",
		Type:      "PROMOTION",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error creating notification: %v", err)
	}

	// Create notification for another user
	_, err = repo.Create(ctx, NotificationRecord{
		UserID:  otherUser,
		Title:   "System Maintenance",
		Message: "Maintenance scheduled at midnight.",
		Type:    "SYSTEM_ALERT",
	})
	if err != nil {
		t.Fatalf("unexpected error creating other user notification: %v", err)
	}

	// 2. List notifications for user
	list, total, unread, err := repo.List(ctx, ListNotificationsParams{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error listing notifications: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if unread != 2 {
		t.Errorf("expected unread 2, got %d", unread)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 items, got %d", len(list))
	}
	// Verify descending order: n2 was created later than n1
	if list[0].ID != n2.ID {
		t.Errorf("expected newest notification first (%s), got %s", n2.ID, list[0].ID)
	}

	// 3. Get notification by ID
	fetched, err := repo.GetByID(ctx, userID, n1.ID)
	if err != nil {
		t.Fatalf("unexpected error getting notification: %v", err)
	}
	if fetched.Title != "Order Confirmed" {
		t.Errorf("expected title 'Order Confirmed', got %s", fetched.Title)
	}

	// 4. Get foreign notification returns not found
	_, err = repo.GetByID(ctx, userID, "non-existent")
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("expected ErrNotificationNotFound, got %v", err)
	}

	// 5. Mark single notification as read
	updatedN1, err := repo.MarkAsRead(ctx, userID, n1.ID)
	if err != nil {
		t.Fatalf("unexpected error marking read: %v", err)
	}
	if !updatedN1.IsRead || updatedN1.ReadAt == nil {
		t.Errorf("expected notification to be marked as read")
	}

	// Check unread count after marking 1 read
	_, _, unread, err = repo.List(ctx, ListNotificationsParams{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if unread != 1 {
		t.Errorf("expected unread 1, got %d", unread)
	}

	// 6. Mark all as read
	markedCount, err := repo.MarkAllAsRead(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error marking all read: %v", err)
	}
	if markedCount != 1 {
		t.Errorf("expected markedCount 1, got %d", markedCount)
	}

	_, _, unread, _ = repo.List(ctx, ListNotificationsParams{
		UserID: userID,
	})
	if unread != 0 {
		t.Errorf("expected unread 0, got %d", unread)
	}

	// 7. Delete notification
	if err := repo.Delete(ctx, userID, n1.ID); err != nil {
		t.Fatalf("unexpected error deleting notification: %v", err)
	}

	// Verify it cannot be retrieved
	_, err = repo.GetByID(ctx, userID, n1.ID)
	if !errors.Is(err, ErrNotificationNotFound) {
		t.Errorf("expected ErrNotificationNotFound after deletion, got %v", err)
	}

	// 8. Test Preferences
	pref, err := repo.GetPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error getting default preferences: %v", err)
	}
	if !pref.EmailEnabled || !pref.PushEnabled || pref.SMSEnabled {
		t.Errorf("unexpected default preferences: %+v", pref)
	}

	emailOptOut := false
	smsOptIn := true
	updatedPref, err := repo.UpsertPreferences(ctx, userID, UpdateNotificationPreferencesParams{
		EmailEnabled: &emailOptOut,
		SMSEnabled:   &smsOptIn,
	})
	if err != nil {
		t.Fatalf("unexpected error updating preferences: %v", err)
	}
	if updatedPref.EmailEnabled || !updatedPref.SMSEnabled {
		t.Errorf("preferences update not reflected: %+v", updatedPref)
	}
}
