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
	Preferences   map[string]*NotificationPreferencesRecord

	// Injected errors for testing failure handling
	ErrList        error
	ErrGet         error
	ErrMarkRead    error
	ErrMarkAllRead error
	ErrDelete      error
	ErrGetPref     error
	ErrUpsertPref  error
	ErrCreate      error
}

// NewMockNotificationRepository initializes an empty in-memory repository.
// Why: Constructs isolated test state for unit testing scenarios.
func NewMockNotificationRepository() *MockNotificationRepository {
	return &MockNotificationRepository{
		Notifications: make(map[string]*NotificationRecord),
		Preferences:   make(map[string]*NotificationPreferencesRecord),
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

// GetPreferences retrieves or initializes mock channel preferences.
func (m *MockNotificationRepository) GetPreferences(ctx context.Context, userID string) (*NotificationPreferencesRecord, error) {
	if m.ErrGetPref != nil {
		return nil, m.ErrGetPref
	}
	p, exists := m.Preferences[userID]
	if !exists {
		now := time.Now().UTC()
		p = &NotificationPreferencesRecord{
			ID:           fmt.Sprintf("pref-%s", userID),
			UserID:       userID,
			EmailEnabled: true,
			PushEnabled:  true,
			SMSEnabled:   false,
			OrderUpdates: true,
			Promotions:   true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		m.Preferences[userID] = p
	}
	return p, nil
}

// UpsertPreferences creates or mutates mock channel preferences.
func (m *MockNotificationRepository) UpsertPreferences(ctx context.Context, userID string, params UpdateNotificationPreferencesParams) (*NotificationPreferencesRecord, error) {
	if m.ErrUpsertPref != nil {
		return nil, m.ErrUpsertPref
	}
	p, _ := m.GetPreferences(ctx, userID)
	now := time.Now().UTC()

	if params.EmailEnabled != nil {
		p.EmailEnabled = *params.EmailEnabled
	}
	if params.PushEnabled != nil {
		p.PushEnabled = *params.PushEnabled
	}
	if params.SMSEnabled != nil {
		p.SMSEnabled = *params.SMSEnabled
	}
	if params.OrderUpdates != nil {
		p.OrderUpdates = *params.OrderUpdates
	}
	if params.Promotions != nil {
		p.Promotions = *params.Promotions
	}
	p.UpdatedAt = now
	return p, nil
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
