package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/Sall-lah/store_user/internal/service"
	"github.com/Sall-lah/store_user/internal/validator"
)

// AdminHandler handles privileged administrative operations for user accounts and governance.
// Why: Isolates administrative transport layer mappings and security controls from customer-facing handlers.
type AdminHandler struct {
	svc service.UserService
}

// NewAdminHandler constructs a new AdminHandler with required business service dependencies.
// Why: Injects business logic contracts into the administrative HTTP handling layer.
func NewAdminHandler(svc service.UserService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// DeleteUser handles DELETE /api/admin/users/{id}.
// Why: Provides administrators with a forced account purge capability that bypasses customer-level active order checks.
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	targetUserID := strings.TrimSpace(chi.URLParam(r, "id"))
	if targetUserID == "" || !validator.ValidateUUID(targetUserID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid target user UUID format",
		})
		return
	}

	err := h.svc.AdminDeleteUser(r.Context(), targetUserID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid target user UUID format",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete user account",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "User account successfully deleted by admin",
	})
}
