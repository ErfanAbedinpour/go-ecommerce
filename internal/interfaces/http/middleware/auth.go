package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"app/internal/domain/user"
	"app/internal/infrastructure/auth"
	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
)

type contextKeyAuth string

const (
	UserIDKey    contextKeyAuth = "user_id"
	UserEmailKey contextKeyAuth = "user_email"
	UserRoleKey  contextKeyAuth = "user_role"
)

// TokenValidator validates JWT access tokens.
type TokenValidator interface {
	ValidateAccessToken(token string) (*auth.Claims, error)
}

// Authenticate validates the Bearer token and injects user claims into context.
// Must be applied before any role-based route guard.
func Authenticate(jwtService TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)
			if err != nil {
				response.Error(w, r, nil, err)
				return
			}

			claims, err := jwtService.ValidateAccessToken(token)
			if err != nil {
				response.Error(w, r, nil, apperror.Unauthorized("invalid or expired access token"))
				return
			}

			role, err := user.ParseRole(claims.Role)
			if err != nil {
				response.Error(w, r, nil, apperror.Unauthorized("invalid role in token"))
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the authenticated user ID from context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(UserIDKey).(string)
	if !ok || id == "" {
		return uuid.Nil, apperror.Unauthorized("authentication required")
	}
	return uuid.Parse(id)
}

// GetUserRole extracts the authenticated user's role from context.
func GetUserRole(ctx context.Context) (user.Role, error) {
	role, ok := ctx.Value(UserRoleKey).(user.Role)
	if !ok || !role.IsValid() {
		return "", apperror.Unauthorized("authentication required")
	}
	return role, nil
}

func extractBearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", apperror.Unauthorized("authorization header required")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", apperror.Unauthorized("invalid authorization header format")
	}

	if parts[1] == "" {
		return "", apperror.Unauthorized("token is required")
	}

	return parts[1], nil
}
