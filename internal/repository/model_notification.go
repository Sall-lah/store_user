package repository

import "time"

// NotificationRecord represents the storage entity for in-app user notifications.
// Why: Provides a persistence-layer abstraction decoupled from Prisma-generated code.
type NotificationRecord struct {
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

// NotificationPreferencesRecord represents persistent channel communication preferences for a user.
// Why: Stores user-specific opt-ins across email, push, SMS, and promotional topics.
type NotificationPreferencesRecord struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	EmailEnabled bool      `json:"emailEnabled"`
	PushEnabled  bool      `json:"pushEnabled"`
	SMSEnabled   bool      `json:"smsEnabled"`
	OrderUpdates bool      `json:"orderUpdates"`
	Promotions   bool      `json:"promotions"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ListNotificationsParams contains filtering and pagination query boundaries for repository lookups.
// Why: Encapsulates optional filtering flags and offset math for database query drivers.
type ListNotificationsParams struct {
	UserID string
	IsRead *bool
	Limit  int
	Offset int
}

// UpdateNotificationPreferencesParams encapsulates mutable flags when updating user notification settings.
// Why: Distinguishes between explicitly modified preference flags and omitted fields.
type UpdateNotificationPreferencesParams struct {
	EmailEnabled *bool
	PushEnabled  *bool
	SMSEnabled   *bool
	OrderUpdates *bool
	Promotions   *bool
}
