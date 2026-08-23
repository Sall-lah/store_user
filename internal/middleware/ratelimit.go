package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sall-lah/store_user/internal/ratelimit"
)

// RateLimit creates an HTTP middleware that throttles requests based on user identity or remote IP address.
// Why: Shields backend microservices from burst traffic, brute-force attacks, and denial-of-service abuse.
func RateLimit(limiter ratelimit.Limiter, limit int, window time.Duration, keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Key based on authenticated userID or fallback IP
			key := GetUserID(r.Context())
			if key == "" {
				key = extractIP(r)
			}

			rateKey := fmt.Sprintf("%s:%s", keyPrefix, key)
			res, err := limiter.Allow(r.Context(), rateKey, limit, window)
			if err != nil || res == nil {
				// Fail-open
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(res.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))

			if !res.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(res.RetryAfterSec))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error":       "too many requests, please slow down",
					"retryAfter": res.RetryAfterSec,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
