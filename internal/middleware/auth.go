package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Sall-lah/store_user/internal/validator"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	userRoleKey  contextKey = "userRole"
	userEmailKey contextKey = "userEmail"
)

// AuthIdentity enforces and extracts verified gateway user headers into the request context.
// Why: Validates anti-spoofed claims injected by the perimeter gateway before request handlers execute.
func AuthIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.Header.Get("X-User-Id"))
		if userID == "" || !validator.ValidateUUID(userID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "unauthorized: missing or invalid X-User-Id header",
			})
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		if role := strings.TrimSpace(r.Header.Get("X-User-Role")); role != "" {
			ctx = context.WithValue(ctx, userRoleKey, role)
		}
		if email := strings.TrimSpace(r.Header.Get("X-User-Email")); email != "" {
			ctx = context.WithValue(ctx, userEmailKey, email)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID retrieves the authenticated user's ID from context.
// Why: Provides a safe accessor for downstream HTTP handlers.
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(userIDKey).(string); ok {
		return val
	}
	return ""
}

// GetUserRole retrieves the authenticated user's role from context.
// Why: Provides role-based access information to HTTP handlers.
func GetUserRole(ctx context.Context) string {
	if val, ok := ctx.Value(userRoleKey).(string); ok {
		return val
	}
	return ""
}

// RequireRole returns a middleware handler that restricts endpoint access to callers matching allowed role claims.
// Why: Enforces perimeter authorization checks so privileged routes cannot be invoked by lower-tiered users.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if userRole == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "forbidden: missing role claims",
				})
				return
			}

			for _, allowed := range allowedRoles {
				if strings.EqualFold(userRole, allowed) {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "forbidden: insufficient permissions",
			})
		})
	}
}

