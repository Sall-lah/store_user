package repository

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// MockNotificationRepository provides an in-memory test double for notification storage.
// Why: Enables fast, deterministic unit testing of notification services without live database engines.
type MockNotificationRepository struct {
	Notifications map[string]*NotificationRecord

	// Injected errors for testing failure handling
	ErrList        error
	ErrGet         error
	ErrMarkRead    error
	ErrMarkAllRead error
	ErrDelete      error
	ErrCreate      error
}

// NewMockNotificationRepository initializes an empty in-memory repository.
// Why: Constructs isolated test state for unit testing scenarios.
func NewMockNotificationRepository() *MockNotificationRepository {
	return &MockNotificationRepository{
		Notifications: make(map[string]*NotificationRecord),
	}
}

// List returns filtered and paginated mock notifications.
func (m *MockNotificationRepository) List(ctx context.Context, params ListNotificationsParams) ([]NotificationRecord, int, int, error) {
	if m.ErrList != nil {
		return nil, 0, 0, m.ErrList
	}

	var allUserNotifications []NotificationRecord
	unreadCount := 0

	for _, n := range m.Notifications {
		if n.UserID == params.UserID {
			if !n.IsRead {
				unreadCount++
			}
			if params.IsRead == nil || n.IsRead == *params.IsRead {
				allUserNotifications = append(allUserNotifications, *n)
			}
		}
	}

	// Sort descending by CreatedAt
	sort.Slice(allUserNotifications, func(i, j int) bool {
		return allUserNotifications[i].CreatedAt.After(allUserNotifications[j].CreatedAt)
	})

	totalCount := len(allUserNotifications)
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	if offset >= len(allUserNotifications) {
		return []NotificationRecord{}, totalCount, unreadCount, nil
	}

	end := offset + limit
	if end > len(allUserNotifications) {
		end = len(allUserNotifications)
	}

	return allUserNotifications[offset:end], totalCount, unreadCount, nil
}

// GetByID retrieves a mock notification by ID and validates ownership.
func (m *MockNotificationRepository) GetByID(ctx context.Context, userID string, id string) (*NotificationRecord, error) {
	if m.ErrGet != nil {
		return nil, m.ErrGet
	}
	n, exists := m.Notifications[id]
	if !exists || n.UserID != userID {
		return nil, ErrNotificationNotFound
	}
	return n, nil
}

// MarkAsRead marks a mock notification as read.
func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, userID string, id string) (*NotificationRecord, error) {
	if m.ErrMarkRead != nil {
		return nil, m.ErrMarkRead
	}
	n, err := m.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	n.IsRead = true
	n.ReadAt = &now
	return n, nil
}

// MarkAllAsRead marks all unread mock notifications for a user as read.
func (m *MockNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) (int, error) {
	if m.ErrMarkAllRead != nil {
		return 0, m.ErrMarkAllRead
	}
	now := time.Now().UTC()
	count := 0
	for _, n := range m.Notifications {
		if n.UserID == userID && !n.IsRead {
			n.IsRead = true
			n.ReadAt = &now
			count++
		}
	}
	return count, nil
}

// Delete removes a mock notification from memory.
func (m *MockNotificationRepository) Delete(ctx context.Context, userID string, id string) error {
	if m.ErrDelete != nil {
		return m.ErrDelete
	}
	_, err := m.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	delete(m.Notifications, id)
	return nil
}


// Create inserts a mock notification into memory.
func (m *MockNotificationRepository) Create(ctx context.Context, item NotificationRecord) (*NotificationRecord, error) {
	if m.ErrCreate != nil {
		return nil, m.ErrCreate
	}
	cp := item
	if cp.ID == "" {
		cp.ID = fmt.Sprintf("notif-%d-%d", time.Now().UnixNano(), len(m.Notifications)+1)
	}
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = time.Now().UTC()
	}
	m.Notifications[cp.ID] = &cp
	return &cp, nil
}
