package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Sall-lah/store_user/internal/repository"
)

// NotificationService defines domain business operations for notifications and preferences.
// Why: Encapsulates business validation, pagination boundaries, and mapping between repository entities and API DTOs.
type NotificationService interface {
	ListNotifications(ctx context.Context, userID string, page int, limit int, isRead *bool) (*NotificationListResponse, error)
	GetNotification(ctx context.Context, userID string, id string) (*NotificationDTO, error)
	MarkAsRead(ctx context.Context, userID string, id string) (*NotificationDTO, error)
	MarkAllAsRead(ctx context.Context, userID string) (*MarkAllReadResponse, error)
	DeleteNotification(ctx context.Context, userID string, id string) error
}

// NotificationServiceImpl implements NotificationService with repository dependency.
type NotificationServiceImpl struct {
	repo repository.NotificationRepository
}

// NewNotificationService constructs a new NotificationServiceImpl with injected repository.
// Why: Enables decoupling and dependency injection for unit testing.
func NewNotificationService(repo repository.NotificationRepository) *NotificationServiceImpl {
	return &NotificationServiceImpl{repo: repo}
}

// ListNotifications retrieves paginated in-app notifications for the caller.
// Why: Sanitizes pagination bounds, calculates offset/total pages, and returns structured envelope with unread counters.
func (s *NotificationServiceImpl) ListNotifications(ctx context.Context, userID string, page int, limit int, isRead *bool) (*NotificationListResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	records, totalCount, unreadCount, err := s.repo.List(ctx, repository.ListNotificationsParams{
		UserID: userID,
		IsRead: isRead,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}

	items := make([]NotificationDTO, len(records))
	for i, r := range records {
		items[i] = mapRepoNotificationToDTO(r)
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	return &NotificationListResponse{
		Items:       items,
		TotalCount:  totalCount,
		UnreadCount: unreadCount,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
	}, nil
}

// GetNotification retrieves a specific notification by ID with user access verification.
// Why: Protects against unauthorized cross-user notification data access.
func (s *NotificationServiceImpl) GetNotification(ctx context.Context, userID string, id string) (*NotificationDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidNotificationID
	}

	rec, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	dto := mapRepoNotificationToDTO(*rec)
	return &dto, nil
}

// MarkAsRead flags a specific notification as read.
// Why: Updates read state and returns the updated DTO to synchronize client state.
func (s *NotificationServiceImpl) MarkAsRead(ctx context.Context, userID string, id string) (*NotificationDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidNotificationID
	}

	rec, err := s.repo.MarkAsRead(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to mark notification as read: %w", err)
	}

	dto := mapRepoNotificationToDTO(*rec)
	return &dto, nil
}

// MarkAllAsRead marks all unread notifications for a user as read.
// Why: Provides bulk state transition for notifications in a single transaction.
func (s *NotificationServiceImpl) MarkAllAsRead(ctx context.Context, userID string) (*MarkAllReadResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	count, err := s.repo.MarkAllAsRead(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return &MarkAllReadResponse{
		UpdatedCount: count,
		Message:      fmt.Sprintf("Successfully marked %d notifications as read", count),
	}, nil
}

// DeleteNotification removes a notification record owned by the user.
// Why: Empowers users to clean up individual notification feed items.
func (s *NotificationServiceImpl) DeleteNotification(ctx context.Context, userID string, id string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(id) == "" {
		return ErrInvalidNotificationID
	}

	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

func mapRepoNotificationToDTO(r repository.NotificationRecord) NotificationDTO {
	return NotificationDTO{
		ID:        r.ID,
		UserID:    r.UserID,
		Title:     r.Title,
		Message:   r.Message,
		Type:      r.Type,
		Data:      r.Data,
		IsRead:    r.IsRead,
		ReadAt:    r.ReadAt,
		CreatedAt: r.CreatedAt,
	}
}

