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
			if origin == "" {
				// Not a CORS request
				next.ServeHTTP(w, r)
				return
			}

			// Get allow credentials from environment
			allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS")
			if allowCredentials == "" {
				allowCredentials = "true"
			}

			// Get allowed methods from environment
			allowedMethods := os.Getenv("CORS_ALLOWED_METHODS")
			if allowedMethods == "" {
				allowedMethods = "GET,POST,OPTIONS"
			}

			// Get allowed headers from environment
			allowedHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
			if allowedHeaders == "" {
				allowedHeaders = "Accept,Content-Type,Content-Length,Accept-Encoding,Authorization,X-CSRF-Token"
			}

			// Set basic CORS headers
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			// Set max age for preflight cache
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Allow credentials if enabled
			if allowCredentials == "true" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				// When credentials are enabled, must use specific origin
				allowed := false
				for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
					if origin == strings.TrimSpace(allowedOrigin) {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						// Also set Vary header when using specific origin
						w.Header().Set("Vary", "Origin")
						break
					}
				}
				if !allowed {
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}
			} else if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				// Check if origin is in allowed list
				allowed := false
				for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
					if origin == strings.TrimSpace(allowedOrigin) {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Vary", "Origin")
						break
					}
				}
				if !allowed {
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				// Add extra headers for preflight
				w.Header().Set("Access-Control-Allow-Private-Network", "true")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Pass to next handler
			next.ServeHTTP(w, r)
		})
	}
}
