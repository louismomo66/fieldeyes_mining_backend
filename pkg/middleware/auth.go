package middleware

import (
	"mineral/data"
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
		r.Header.Set("X-Chain-Role", claims.ChainRole)

		next.ServeHTTP(w, r)
	})
}

// GetChainRoleFromRequest returns the supply-chain role from the validated token
// (empty for legacy users, who are treated as full-access).
func GetChainRoleFromRequest(r *http.Request) string {
	return r.Header.Get("X-Chain-Role")
}

// RequireChainRole restricts a route to the given supply-chain roles. It is
// deliberately permissive so nothing breaks on the deployed system:
//   - platform admins always pass,
//   - legacy users (empty chain role) always pass,
//   - otherwise the user's chain role must be in the allowed set.
func RequireChainRole(roles ...data.ChainRole) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[string(role)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-User-Role") == string(data.RoleAdmin) {
				next.ServeHTTP(w, r)
				return
			}
			chainRole := r.Header.Get("X-Chain-Role")
			if chainRole == "" || allowed[chainRole] {
				next.ServeHTTP(w, r)
				return
			}
			utils.WriteErrorResponse(w, "Your role does not permit this action", http.StatusForbidden)
		})
	}
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
