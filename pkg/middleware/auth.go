package middleware

import (
	"mineral/pkg/utils"
	"net/http"
	"strconv"
	"strings"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.WriteErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			utils.WriteErrorResponse(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		r.Header.Set("X-User-ID", claims.UserID)
		r.Header.Set("X-User-Email", claims.Email)
		r.Header.Set("X-User-Role", claims.Role)

		next.ServeHTTP(w, r)
	})
}

// AdminMiddleware checks if user has admin role
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Header.Get("X-User-Role")
		if userRole != "admin" {
			utils.WriteErrorResponse(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromRequest extracts user ID from request headers
func GetUserIDFromRequest(r *http.Request) uint {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return 0
	}
	return uint(userID)
}
