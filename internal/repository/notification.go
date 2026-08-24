package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Sall-lah/store_user/internal/db"
	"github.com/steebchen/prisma-client-go/runtime/types"
)

// ErrNotificationNotFound indicates the target notification record does not exist.
var ErrNotificationNotFound = errors.New("notification not found")

// NotificationRepository defines database abstraction for user notifications and channel preferences.
// Why: Decouples service business logic from Prisma ORM for simplified unit testing and database independence.
type NotificationRepository interface {
	List(ctx context.Context, params ListNotificationsParams) ([]NotificationRecord, int, int, error)
	GetByID(ctx context.Context, userID string, id string) (*NotificationRecord, error)
	MarkAsRead(ctx context.Context, userID string, id string) (*NotificationRecord, error)
	MarkAllAsRead(ctx context.Context, userID string) (int, error)
	Delete(ctx context.Context, userID string, id string) error
	Create(ctx context.Context, item NotificationRecord) (*NotificationRecord, error)
}

// PrismaNotificationRepository implements NotificationRepository using Prisma Client Go.
// Why: Provides persistent database queries targeting PostgreSQL table schemas.
type PrismaNotificationRepository struct {
	client *db.PrismaClient
}

// NewPrismaNotificationRepository initializes repository instance with active Prisma client.
// Why: Injects database connection pool into repository query methods.
func NewPrismaNotificationRepository(client *db.PrismaClient) *PrismaNotificationRepository {
	return &PrismaNotificationRepository{client: client}
}

// List queries paginated notifications for a given user with optional read status filter.
// Why: Supports in-app notification center queries with count aggregations for badge counters.
func (r *PrismaNotificationRepository) List(ctx context.Context, params ListNotificationsParams) ([]NotificationRecord, int, int, error) {
	var filters []db.UserNotificationsWhereParam
	filters = append(filters, db.UserNotifications.UserID.Equals(params.UserID))

	if params.IsRead != nil {
		filters = append(filters, db.UserNotifications.IsRead.Equals(*params.IsRead))
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// 1. Fetch paginated records
	models, err := r.client.UserNotifications.FindMany(
		filters...,
	).OrderBy(
		db.UserNotifications.CreatedAt.Order(db.SortOrderDesc),
	).Take(limit).Skip(offset).Exec(ctx)

	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to query notifications: %w", err)
	}

	// 2. Count total notifications for user (all or filtered)
	totalModels, err := r.client.UserNotifications.FindMany(
		filters...,
	).Exec(ctx)
	totalCount := len(totalModels)
	if err != nil {
		totalCount = len(models)
	}

	// 3. Count unread notifications
	unreadModels, err := r.client.UserNotifications.FindMany(
		db.UserNotifications.UserID.Equals(params.UserID),
		db.UserNotifications.IsRead.Equals(false),
	).Exec(ctx)
	unreadCount := len(unreadModels)
	if err != nil {
		unreadCount = 0
	}

	records := make([]NotificationRecord, len(models))
	for i, m := range models {
		records[i] = *mapPrismaNotificationModel(&m)
	}

	return records, totalCount, unreadCount, nil
}

// GetByID fetches a specific notification ensuring it belongs to the authenticated caller.
// Why: Enforces user multi-tenancy isolation and prevents cross-user information leakage.
func (r *PrismaNotificationRepository) GetByID(ctx context.Context, userID string, id string) (*NotificationRecord, error) {
	m, err := r.client.UserNotifications.FindUnique(
		db.UserNotifications.ID.Equals(id),
	).Exec(ctx)

	if err != nil {
		if db.IsErrNotFound(err) || errors.Is(err, db.ErrNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to find notification by id: %w", err)
	}

	if m.UserID != userID {
		return nil, ErrNotificationNotFound
	}

	return mapPrismaNotificationModel(m), nil
}

// MarkAsRead flags a notification as read and records current timestamp.
// Why: Provides atomic read-state transition for individual notification items.
func (r *PrismaNotificationRepository) MarkAsRead(ctx context.Context, userID string, id string) (*NotificationRecord, error) {
	// Verify ownership first
	existing, err := r.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	m, err := r.client.UserNotifications.FindUnique(
		db.UserNotifications.ID.Equals(existing.ID),
	).Update(
		db.UserNotifications.IsRead.Set(true),
		db.UserNotifications.ReadAt.Set(types.DateTime(now)),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update notification read status: %w", err)
	}

	return mapPrismaNotificationModel(m), nil
}

// MarkAllAsRead marks all unread notifications for a user as read.
// Why: Enables 'Mark all as read' UX action in a single batch operation.
func (r *PrismaNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) (int, error) {
	now := time.Now().UTC()
	res, err := r.client.UserNotifications.FindMany(
		db.UserNotifications.UserID.Equals(userID),
		db.UserNotifications.IsRead.Equals(false),
	).Update(
		db.UserNotifications.IsRead.Set(true),
		db.UserNotifications.ReadAt.Set(types.DateTime(now)),
	).Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return res.Count, nil
}

// Delete permanently removes a notification by ID if owned by the caller.
// Why: Allows users to dismiss and purge individual notification entries.
func (r *PrismaNotificationRepository) Delete(ctx context.Context, userID string, id string) error {
	existing, err := r.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	_, err = r.client.UserNotifications.FindUnique(
		db.UserNotifications.ID.Equals(existing.ID),
	).Delete().Exec(ctx)

	if err != nil {
		if db.IsErrNotFound(err) || errors.Is(err, db.ErrNotFound) || strings.Contains(err.Error(), "Record to delete does not exist") {
			return nil
		}
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

// Create inserts a new notification record.
// Why: Facilitates integration testing and internal event ingestion.
func (r *PrismaNotificationRepository) Create(ctx context.Context, item NotificationRecord) (*NotificationRecord, error) {
	var optionalParams []db.UserNotificationsSetParam
	if item.Data != nil {
		optionalParams = append(optionalParams, db.UserNotifications.Data.Set(*item.Data))
	}
	if item.Type != "" {
		optionalParams = append(optionalParams, db.UserNotifications.Type.Set(item.Type))
	}

	m, err := r.client.UserNotifications.CreateOne(
		db.UserNotifications.UserID.Set(item.UserID),
		db.UserNotifications.Title.Set(item.Title),
		db.UserNotifications.Message.Set(item.Message),
		optionalParams...,
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create notification record: %w", err)
	}

	return mapPrismaNotificationModel(m), nil
}

func mapPrismaNotificationModel(m *db.UserNotificationsModel) *NotificationRecord {
	if m == nil {
		return nil
	}
	var readAt *time.Time
	if m.InnerUserNotifications.ReadAt != nil {
		t := time.Time(*m.InnerUserNotifications.ReadAt)
		readAt = &t
	}
	return &NotificationRecord{
		ID:        m.ID,
		UserID:    m.UserID,
		Title:     m.Title,
		Message:   m.Message,
		Type:      m.Type,
		Data:      m.InnerUserNotifications.Data,
		IsRead:    m.IsRead,
		ReadAt:    readAt,
		CreatedAt: time.Time(m.CreatedAt),
	}
}

