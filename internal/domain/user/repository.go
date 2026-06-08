package user

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for admin user listing.
type ListFilter struct {
	Query string
	Role  *Role
}

// Repository defines the port for user persistence.
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]User, int64, error)
	Update(ctx context.Context, u *User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	CountByRole(ctx context.Context, role Role) (int64, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
}

// PasswordResetRepository defines the port for password reset token persistence.
type PasswordResetRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	FindByHash(ctx context.Context, hash string) (*PasswordResetToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	InvalidateByUser(ctx context.Context, userID uuid.UUID) error
}

// RefreshTokenRepository defines the port for refresh token persistence.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID) error
}
