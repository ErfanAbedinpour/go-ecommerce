package adminuser

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the port for admin user persistence.
type Repository interface {
	FindByEmail(ctx context.Context, email string) (*AdminUser, error)
	FindByID(ctx context.Context, id uuid.UUID) (*AdminUser, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

// RefreshTokenRepository defines the port for refresh token persistence.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
}
