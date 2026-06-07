package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/brand"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// BrandRepository implements brand.Repository using GORM.
type BrandRepository struct {
	db *gorm.DB
}

// NewBrandRepository creates a new BrandRepository.
func NewBrandRepository(db *gorm.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) Create(ctx context.Context, b *brand.Brand) error {
	m := toBrandModel(b)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return mapBrandDBError(err)
	}
	return nil
}

func (r *BrandRepository) Update(ctx context.Context, b *brand.Brand) error {
	result := r.db.WithContext(ctx).Model(&models.BrandModel{}).
		Where("id = ?", b.ID).
		Updates(map[string]any{
			"name":        b.Name,
			"slug":        b.Slug,
			"description": nullableString(b.Description),
			"is_active":   b.IsActive,
			"updated_at":  b.UpdatedAt,
		})
	if result.Error != nil {
		return mapBrandDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return brand.ErrNotFound
	}
	return nil
}

func (r *BrandRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.BrandModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return brand.ErrNotFound
	}
	return nil
}

func (r *BrandRepository) FindByID(ctx context.Context, id uuid.UUID) (*brand.Brand, error) {
	var m models.BrandModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, brand.ErrNotFound
		}
		return nil, err
	}
	return toBrandDomain(&m), nil
}

func (r *BrandRepository) FindBySlug(ctx context.Context, slug string) (*brand.Brand, error) {
	var m models.BrandModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, brand.ErrNotFound
		}
		return nil, err
	}
	return toBrandDomain(&m), nil
}

func (r *BrandRepository) List(ctx context.Context, filter brand.ListFilter, page pagination.Params) ([]brand.Brand, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.BrandModel{})
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.BrandModel
	err := query.Order(brandOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toBrandsDomain(items), total, nil
}

func (r *BrandRepository) HasProducts(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProductModel{}).
		Where("brand = ?", name).
		Count(&count).Error
	return count > 0, err
}

func brandOrderClause(page pagination.Params) string {
	allowed := map[string]string{
		"created_at": "created_at",
		"name":       "name",
		"updated_at": "updated_at",
	}
	column, ok := allowed[page.Sort]
	if !ok {
		column = "name"
	}
	order := "ASC"
	if strings.EqualFold(page.Order, "desc") {
		order = "DESC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

func toBrandModel(b *brand.Brand) *models.BrandModel {
	return &models.BrandModel{
		ID:          b.ID,
		Name:        b.Name,
		Slug:        b.Slug,
		Description: nullableString(b.Description),
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func toBrandDomain(m *models.BrandModel) *brand.Brand {
	b := &brand.Brand{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.Description != nil {
		b.Description = *m.Description
	}
	return b
}

func toBrandsDomain(items []models.BrandModel) []brand.Brand {
	result := make([]brand.Brand, len(items))
	for i, m := range items {
		result[i] = *toBrandDomain(&m)
	}
	return result
}

func mapBrandDBError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "slug") {
		return brand.ErrSlugConflict
	}
	if strings.Contains(msg, "name") {
		return brand.ErrNameConflict
	}
	return err
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
