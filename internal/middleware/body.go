package middleware

import "net/http"

// MaxBodySize restricts the maximum allowed request body size to protect against memory exhaustion attacks.
// Why: Enforces perimeter payload limits before json deserializers allocate memory.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 && r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
