package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Sall-lah/store_user/internal/middleware"
	"github.com/Sall-lah/store_user/internal/service"
	"github.com/Sall-lah/store_user/internal/validator"
)

// ProfileHandler handles incoming HTTP requests for user profile management and account lifecycle.
// Why: Bridges HTTP transport parsing, status mapping, and JSON response encoding with domain services.
type ProfileHandler struct {
	svc service.UserService
}

// NewProfileHandler initializes a ProfileHandler with the provided service dependency.
// Why: Injects business service logic into the HTTP handler layer.
func NewProfileHandler(svc service.UserService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

// GetProfile handles GET /api/v1/users/profile.
// Why: Returns the authenticated user's profile details.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "profile not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile handles PUT /api/v1/users/profile.
// Why: Sanitizes input fields and mutates user profile attributes.
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req service.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "malformed JSON payload"})
		return
	}

	// Sanitize text inputs
	if req.FullName != nil {
		cleanName := validator.SanitizeText(*req.FullName)
		req.FullName = &cleanName
	}
	if req.Bio != nil {
		cleanBio := validator.SanitizeText(*req.Bio)
		req.Bio = &cleanBio
	}
	if req.Address != nil {
		cleanAddr := validator.SanitizeText(*req.Address)
		req.Address = &cleanAddr
	}
	if req.PhoneNumber != nil && *req.PhoneNumber != "" {
		if !validator.ValidatePhoneNumber(*req.PhoneNumber) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid phone number format"})
			return
		}
	}
	if req.Gender != nil && *req.Gender != "" {
		if !validator.ValidateGender(*req.Gender) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid gender value; must be MALE, FEMALE, OTHER, or PREFER_NOT_TO_SAY"})
			return
		}
	}

	profile, err := h.svc.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrValidationFailed) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// DeleteAccount handles DELETE /api/v1/users/account.
// Why: Validates active order status over gRPC and permanently deletes account credentials.
func (h *ProfileHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req service.DeleteAccountRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	err := h.svc.DeleteAccount(r.Context(), userID, req)
	if err != nil {
		var conflictErr *service.ActiveOrdersConflictError
		if errors.As(err, &conflictErr) {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"error":            conflictErr.Error(),
				"activeOrderCount": conflictErr.Count,
				"activeOrders":     conflictErr.ActiveOrders,
			})
			return
		}

		if errors.Is(err, service.ErrOrderServiceUnavailable) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "order verification service is temporarily unavailable, please retry later",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete user account",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Account successfully deleted",
	})
}

// HealthCheck handles GET /health.
// Why: Provides container health and readiness probe endpoint.
func (h *ProfileHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "UP",
		"service": "store_user",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
