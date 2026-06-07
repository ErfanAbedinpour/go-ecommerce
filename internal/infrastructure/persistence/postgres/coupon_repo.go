package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/coupon"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

var allowedCouponSorts = map[string]string{
	"created_at": "coupons.created_at",
	"code":       "coupons.code",
	"updated_at": "coupons.updated_at",
}

// CouponRepository implements coupon.Repository using GORM.
type CouponRepository struct {
	db *gorm.DB
}

// NewCouponRepository creates a new CouponRepository.
func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) Create(ctx context.Context, c *coupon.Coupon) error {
	m := toCouponModel(c)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return mapCouponDBError(err)
	}
	return nil
}

func (r *CouponRepository) Update(ctx context.Context, c *coupon.Coupon) error {
	m := toCouponModel(c)
	result := r.db.WithContext(ctx).Model(&models.CouponModel{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"code":             m.Code,
			"discount_type":    m.DiscountType,
			"discount_value":   m.DiscountValue,
			"min_order_amount": m.MinOrderAmount,
			"max_usage":        m.MaxUsage,
			"expires_at":       m.ExpiresAt,
			"is_active":        m.IsActive,
			"note":             m.Note,
			"updated_at":       c.UpdatedAt,
		})
	if result.Error != nil {
		return mapCouponDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return coupon.ErrNotFound
	}
	return nil
}

func (r *CouponRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.CouponModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return coupon.ErrNotFound
	}
	return nil
}

func (r *CouponRepository) FindByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error) {
	var m models.CouponModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, coupon.ErrNotFound
		}
		return nil, err
	}
	return toCouponDomain(&m), nil
}

func (r *CouponRepository) FindByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	var m models.CouponModel
	err := r.db.WithContext(ctx).Where("code = ?", coupon.NormalizeCode(code)).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, coupon.ErrNotFound
		}
		return nil, err
	}
	return toCouponDomain(&m), nil
}

func (r *CouponRepository) List(ctx context.Context, filter coupon.ListFilter, page pagination.Params) ([]coupon.Coupon, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.CouponModel{})

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where("LOWER(code) LIKE ? OR LOWER(COALESCE(note, '')) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.CouponModel
	err := query.
		Order(r.orderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toCouponsDomain(items), total, nil
}

func (r *CouponRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	result := r.db.WithContext(ctx).
		Model(&models.CouponModel{}).
		Where("id = ?", id).
		Update("is_active", active)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return coupon.ErrNotFound
	}
	return nil
}

func (r *CouponRepository) orderClause(page pagination.Params) string {
	column, ok := allowedCouponSorts[page.Sort]
	if !ok {
		column = allowedCouponSorts["created_at"]
	}
	order := "DESC"
	if strings.EqualFold(page.Order, "asc") {
		order = "ASC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

func mapCouponDBError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "code") {
		return coupon.ErrCodeConflict
	}
	return err
}
