package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/user"
	"app/internal/infrastructure/persistence/models"
)

// PasswordResetRepository implements user.PasswordResetRepository using GORM.
type PasswordResetRepository struct {
	db *gorm.DB
}

// NewPasswordResetRepository creates a new PasswordResetRepository.
func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token *user.PasswordResetToken) error {
	m := &models.PasswordResetTokenModel{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash:   token.TokenHash,
		ExpiresAt:   token.ExpiresAt,
		CreatedAt:   token.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *PasswordResetRepository) FindByHash(ctx context.Context, hash string) (*user.PasswordResetToken, error) {
	var m models.PasswordResetTokenModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrInvalidResetToken
		}
		return nil, err
	}
	return toPasswordResetDomain(&m), nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetTokenModel{}).
		Where("id = ?", id).
		Update("used_at", now).Error
}

func (r *PasswordResetRepository) InvalidateByUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetTokenModel{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}

func toPasswordResetDomain(m *models.PasswordResetTokenModel) *user.PasswordResetToken {
	return &user.PasswordResetToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		CreatedAt: m.CreatedAt,
	}
}
