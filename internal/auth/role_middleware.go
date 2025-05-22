package auth

import (
	"net/http"
)

// RoleMiddleware returns a middleware function that checks if the user's role
// is one of the allowed roles.
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ContextKeyClaims).(*Claims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Check if the role in claims is in the allowed list
			allowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "Forbidden: insufficient privileges", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
