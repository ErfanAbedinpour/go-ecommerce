package user

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a one-time password reset token.
type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// IsUsed returns true if the token has already been consumed.
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsExpired returns true if the token has expired.
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// IsValid reports whether the token can still be used.
func (t *PasswordResetToken) IsValid() bool {
	return !t.IsUsed() && !t.IsExpired()
}
