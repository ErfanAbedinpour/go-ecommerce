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

// RefreshTokenRepository implements user.RefreshTokenRepository using GORM.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *user.RefreshToken) error {
	m := &models.RefreshTokenModel{
		ID:          token.ID,
		AdminUserID: token.UserID,
		TokenHash:   token.TokenHash,
		FamilyID:    token.FamilyID,
		ExpiresAt:   token.ExpiresAt,
		CreatedAt:   token.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*user.RefreshToken, error) {
	var m models.RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrInvalidToken
		}
		return nil, err
	}
	return toRefreshTokenDomain(&m), nil
}

func (r *RefreshTokenRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Update("revoked_at", now).Error
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("id = ?", id).
		Update("revoked_at", now).Error
}
