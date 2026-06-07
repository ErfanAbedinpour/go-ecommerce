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

// UserRepository implements user.Repository using GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	m := &models.UserModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Role:         u.Role.String(),
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
	if u.Phone != "" {
		m.Phone = &u.Phone
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var m models.UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return toUserDomain(&m), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var m models.UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return toUserDomain(&m), nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ?", id).
		Update("last_login_at", now).Error
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password_hash": passwordHash,
			"updated_at":    now,
		}).Error
}
