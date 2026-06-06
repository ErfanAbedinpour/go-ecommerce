package middleware

import (
	"net/http"

	"app/internal/domain/user"
	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
)

// RequireRole enforces that the authenticated user has one of the allowed roles.
// Authorization is enforced at the router/application layer using JWT role claims.
// Authenticate middleware must run before this guard.
func RequireRole(roles ...user.Role) func(http.Handler) http.Handler {
	allowed := make(map[user.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := GetUserRole(r.Context())
			if err != nil {
				response.Error(w, r, nil, err)
				return
			}

			if _, ok := allowed[role]; !ok {
				response.Error(w, r, nil, apperror.Forbidden("access denied for role: "+role.String()))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a convenience guard for admin-only routes.
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRole(user.RoleAdmin)
}

// RequireCustomer is a convenience guard for customer-only routes.
func RequireCustomer() func(http.Handler) http.Handler {
	return RequireRole(user.RoleCustomer)
}
