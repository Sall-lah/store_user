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

// ListNotificationsParams contains filtering and pagination query boundaries for repository lookups.
// Why: Encapsulates optional filtering flags and offset math for database query drivers.
type ListNotificationsParams struct {
	UserID string
	IsRead *bool
	Limit  int
	Offset int
}

