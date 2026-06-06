package middleware

import (
	"net/http"
	"slices"

	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
)

// RequirePermission checks that the authenticated user has the required permission.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			perms := GetUserPermissions(r.Context())
			if !slices.Contains(perms, permission) {
				response.Error(w, nil, apperror.Forbidden("insufficient permissions"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission checks that the user has at least one of the given permissions.
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userPerms := GetUserPermissions(r.Context())
			for _, required := range permissions {
				if slices.Contains(userPerms, required) {
					next.ServeHTTP(w, r)
					return
				}
			}
			response.Error(w, nil, apperror.Forbidden("insufficient permissions"))
		})
	}
}
