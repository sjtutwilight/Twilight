package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CorsMiddleware handles CORS (Cross-Origin Resource Sharing)
func CorsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get allowed origins from environment
			allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
			if allowedOrigins == "" {
				allowedOrigins = "*" // Default to all origins in development
			}

			// Get origin from request
			origin := r.Header.Get("Origin")

			// If specific origins are set, check if the request origin is allowed
			if allowedOrigins != "*" {
				allowed := false
				for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
					if origin == strings.TrimSpace(allowedOrigin) {
						allowed = true
						break
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// Get allowed methods from environment
			allowedMethods := os.Getenv("CORS_ALLOWED_METHODS")
			if allowedMethods == "" {
				allowedMethods = "GET,POST,OPTIONS"
			}
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)

			// Get allowed headers from environment
			allowedHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
			if allowedHeaders == "" {
				allowedHeaders = "Accept,Content-Type,Content-Length,Accept-Encoding,Authorization,X-CSRF-Token"
			}
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			// Get allow credentials from environment
			allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS")
			if allowCredentials == "" {
				allowCredentials = "true"
			}
			w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Pass to next handler
			next.ServeHTTP(w, r)
		})
	}
}
