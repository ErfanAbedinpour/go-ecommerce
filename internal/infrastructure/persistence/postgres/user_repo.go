package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/user"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

var allowedUserSorts = map[string]string{
	"created_at": "admin_users.created_at",
	"email":      "admin_users.email",
	"first_name": "admin_users.first_name",
	"last_name":  "admin_users.last_name",
	"updated_at": "admin_users.updated_at",
}

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

func (r *UserRepository) List(ctx context.Context, filter user.ListFilter, page pagination.Params) ([]user.User, int64, error) {
	query := r.applyListFilters(r.db.WithContext(ctx).Model(&models.UserModel{}), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.UserModel
	err := query.
		Order(r.userOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toUsersDomain(items), total, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	m := toUserModel(u)
	result := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ?", u.ID).
		Updates(map[string]any{
			"email":         m.Email,
			"password_hash": m.PasswordHash,
			"first_name":    m.FirstName,
			"last_name":     m.LastName,
			"phone":         m.Phone,
			"role":          m.Role,
			"is_active":     m.IsActive,
			"updated_at":    m.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.UserModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("role = ? AND is_active = ?", role.String(), true).
		Count(&count).Error
	return count, err
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

func (r *UserRepository) applyListFilters(query *gorm.DB, filter user.ListFilter) *gorm.DB {
	if filter.Role != nil {
		query = query.Where("role = ?", filter.Role.String())
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(CONCAT(first_name, ' ', last_name)) LIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}
	return query
}

func (r *UserRepository) userOrderClause(page pagination.Params) string {
	column, ok := allowedUserSorts[page.Sort]
	if !ok {
		column = allowedUserSorts["created_at"]
	}
	order := "DESC"
	if strings.EqualFold(page.Order, "asc") {
		order = "ASC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

var _ user.Repository = (*UserRepository)(nil)
