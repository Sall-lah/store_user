package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/Sall-lah/store_user/internal/middleware"
	"github.com/Sall-lah/store_user/internal/service"
)

// NotificationHandler handles incoming HTTP requests for user notifications and channel preferences.
// Why: Maps HTTP transport parameters, headers, and query strings to notification service operations.
type NotificationHandler struct {
	svc service.NotificationService
}

// NewNotificationHandler initializes a NotificationHandler with injected service.
// Why: Injects business notification services into the HTTP routing layer.
func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// ListNotifications handles GET /api/users/notifications.
// Why: Returns a paginated list of user notifications with unread badge count metadata.
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var isReadPtr *bool
	if isReadStr := r.URL.Query().Get("is_read"); isReadStr != "" {
		if val, err := strconv.ParseBool(isReadStr); err == nil {
			isReadPtr = &val
		}
	}

	resp, err := h.svc.ListNotifications(r.Context(), userID, page, limit, isReadPtr)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list notifications"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetNotification handles GET /api/users/notifications/{id}.
// Why: Returns details for an individual notification item.
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing notification id"})
		return
	}

	dto, err := h.svc.GetNotification(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "notification not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidNotificationID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid notification id"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// MarkAsRead handles PATCH /api/users/notifications/{id}/read.
// Why: Updates the read status of a single notification.
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing notification id"})
		return
	}

	dto, err := h.svc.MarkAsRead(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "notification not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidNotificationID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid notification id"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update notification read status"})
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// MarkAllAsRead handles POST /api/users/notifications/read-all.
// Why: Marks all unread notifications for the caller as read in a single request.
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := h.svc.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to mark all notifications as read"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteNotification handles DELETE /api/users/notifications/{id}.
// Why: Permanently deletes a single notification from the user's feed.
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing notification id"})
		return
	}

	err := h.svc.DeleteNotification(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "notification not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidNotificationID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid notification id"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete notification"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Notification successfully deleted",
	})
}

