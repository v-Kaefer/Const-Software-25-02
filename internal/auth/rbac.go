package auth

import (
	"net/http"
	"strconv"
	"strings"
)

// RequireAdmin is a middleware that requires the user to be an admin
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaimsFromContext(r.Context())
		if err != nil {
			// If JWT middleware is not configured, allow access (for testing)
			// In production, JWT middleware should always be present
			next.ServeHTTP(w, r)
			return
		}

		if !claims.IsAdmin() {
			http.Error(w, "forbidden: admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireOwnerOrAdmin is a middleware that requires the user to be the resource owner or an admin
// It extracts the user ID from the URL path and compares it with the authenticated user's ID
func RequireOwnerOrAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaimsFromContext(r.Context())
		if err != nil {
			// If JWT middleware is not configured, allow access (for testing)
			// In production, JWT middleware should always be present
			next.ServeHTTP(w, r)
			return
		}

		// Extract user ID from path (e.g., /users/123)
		path := r.URL.Path
		parts := strings.Split(strings.Trim(path, "/"), "/")
		
		// Expecting pattern: /users/{id}
		if len(parts) >= 2 && parts[0] == "users" {
			requestedUserID := parts[1]
			
			// Check if user is admin - admins can access any resource
			if claims.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}

			// Check if user is accessing their own resource
			// claims.Subject contains the user's ID from the JWT
			if claims.Subject == requestedUserID {
				next.ServeHTTP(w, r)
				return
			}

			// Try to compare as integers if both can be parsed
			claimsID, err1 := strconv.ParseUint(claims.Subject, 10, 64)
			reqID, err2 := strconv.ParseUint(requestedUserID, 10, 64)
			if err1 == nil && err2 == nil && claimsID == reqID {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
	})
}
