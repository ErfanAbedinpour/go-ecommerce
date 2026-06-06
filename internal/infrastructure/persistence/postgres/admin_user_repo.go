package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/adminuser"
	"app/internal/infrastructure/persistence/models"
)

// AdminUserRepository implements adminuser.Repository using GORM.
type AdminUserRepository struct {
	db *gorm.DB
}

// NewAdminUserRepository creates a new AdminUserRepository.
func NewAdminUserRepository(db *gorm.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

func (r *AdminUserRepository) FindByEmail(ctx context.Context, email string) (*adminuser.AdminUser, error) {
	var m models.AdminUserModel
	err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Where("email = ?", email).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminuser.ErrNotFound
		}
		return nil, err
	}
	return toAdminUserDomain(&m), nil
}

func (r *AdminUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*adminuser.AdminUser, error) {
	var m models.AdminUserModel
	err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Where("id = ?", id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, adminuser.ErrNotFound
		}
		return nil, err
	}
	return toAdminUserDomain(&m), nil
}

func (r *AdminUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.AdminUserModel{}).
		Where("id = ?", id).
		Update("last_login_at", now).Error
}
