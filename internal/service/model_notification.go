package service

import (
	"errors"
	"time"
)

// ErrNotificationNotFound indicates the requested notification ID does not exist or does not belong to the user.
var ErrNotificationNotFound = errors.New("notification not found")

// ErrInvalidNotificationID indicates that the provided notification identifier is malformed or invalid.
var ErrInvalidNotificationID = errors.New("invalid notification id")

// NotificationDTO represents an in-app notification returned to API consumers.
// Why: Exposes domain-safe notification data over HTTP without leaking internal database schemas.
type NotificationDTO struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Type      string     `json:"type"`
	Data      *string    `json:"data,omitempty"`
	IsRead    bool       `json:"isRead"`
	ReadAt    *time.Time `json:"readAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// NotificationListResponse represents the paginated envelope for notification queries.
// Why: Provides metadata required for client-side pagination controls and unread badge counters.
type NotificationListResponse struct {
	Items       []NotificationDTO `json:"items"`
	TotalCount  int               `json:"totalCount"`
	UnreadCount int               `json:"unreadCount"`
	Page        int               `json:"page"`
	Limit       int               `json:"limit"`
	TotalPages  int               `json:"totalPages"`
}

// NotificationPreferencesDTO represents channel preferences returned to API callers.
// Why: Standardizes user notification settings across UI preference toggles.
type NotificationPreferencesDTO struct {
	UserID       string    `json:"userId"`
	EmailEnabled bool      `json:"emailEnabled"`
	PushEnabled  bool      `json:"pushEnabled"`
	SMSEnabled   bool      `json:"smsEnabled"`
	OrderUpdates bool      `json:"orderUpdates"`
	Promotions   bool      `json:"promotions"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// UpdateNotificationPreferencesRequest represents the JSON request payload for altering channel preferences.
// Why: Allows partial updates to user notification flags without overwriting unmodified channels.
type UpdateNotificationPreferencesRequest struct {
	EmailEnabled *bool `json:"emailEnabled,omitempty"`
	PushEnabled  *bool `json:"pushEnabled,omitempty"`
	SMSEnabled   *bool `json:"smsEnabled,omitempty"`
	OrderUpdates *bool `json:"orderUpdates,omitempty"`
	Promotions   *bool `json:"promotions,omitempty"`
}

// MarkAllReadResponse conveys the outcome of batch-marking notifications as read.
// Why: Returns the number of affected notification records for immediate client badge synchronization.
type MarkAllReadResponse struct {
	UpdatedCount int    `json:"updatedCount"`
	Message      string `json:"message"`
}
