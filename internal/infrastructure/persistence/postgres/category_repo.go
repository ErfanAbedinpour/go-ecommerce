package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/category"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

var allowedCategorySorts = map[string]string{
	"created_at": "categories.created_at",
	"name":       "categories.name",
	"sort_order": "categories.sort_order",
	"updated_at": "categories.updated_at",
}

// CategoryRepository implements category.Repository using GORM.
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, c *category.Category) error {
	m := toCategoryModel(c)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return mapCategoryDBError(err)
	}
	return nil
}

func (r *CategoryRepository) Update(ctx context.Context, c *category.Category) error {
	m := toCategoryModel(c)
	result := r.db.WithContext(ctx).Model(&models.CategoryModel{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"parent_id":   m.ParentID,
			"name":        m.Name,
			"slug":        m.Slug,
			"description": m.Description,
			"image_url":   m.ImageURL,
			"sort_order":  m.SortOrder,
			"is_active":   m.IsActive,
			"updated_at":  c.UpdatedAt,
		})
	if result.Error != nil {
		return mapCategoryDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return category.ErrNotFound
	}
	return nil
}

func (r *CategoryRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.CategoryModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return category.ErrNotFound
	}
	return nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*category.Category, error) {
	var m models.CategoryModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, category.ErrNotFound
		}
		return nil, err
	}
	return toCategoryDomain(&m), nil
}

func (r *CategoryRepository) FindBySlug(ctx context.Context, slug string) (*category.Category, error) {
	var m models.CategoryModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, category.ErrNotFound
		}
		return nil, err
	}
	return toCategoryDomain(&m), nil
}

func (r *CategoryRepository) List(ctx context.Context, filter category.ListFilter, page pagination.Params) ([]category.Category, int64, error) {
	query := r.applyFilters(r.db.WithContext(ctx).Model(&models.CategoryModel{}), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.CategoryModel
	err := query.
		Order(r.orderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toCategoriesDomain(items), total, nil
}

func (r *CategoryRepository) ListAll(ctx context.Context, filter category.ListFilter) ([]category.Category, error) {
	query := r.applyFilters(r.db.WithContext(ctx).Model(&models.CategoryModel{}), filter)

	var items []models.CategoryModel
	err := query.Order("sort_order ASC, name ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toCategoriesDomain(items), nil
}

func (r *CategoryRepository) HasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CategoryModel{}).
		Where("parent_id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func (r *CategoryRepository) HasProducts(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProductModel{}).
		Where("category_id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func (r *CategoryRepository) IsDescendant(ctx context.Context, ancestorID, nodeID uuid.UUID) (bool, error) {
	if ancestorID == nodeID {
		return true, nil
	}

	current := nodeID
	visited := make(map[uuid.UUID]struct{})

	for {
		if _, ok := visited[current]; ok {
			return false, nil
		}
		visited[current] = struct{}{}

		var m models.CategoryModel
		err := r.db.WithContext(ctx).Select("id", "parent_id").Where("id = ?", current).First(&m).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}
			return false, err
		}
		if m.ParentID == nil {
			return false, nil
		}
		if *m.ParentID == ancestorID {
			return true, nil
		}
		current = *m.ParentID
	}
}

func (r *CategoryRepository) applyFilters(query *gorm.DB, filter category.ListFilter) *gorm.DB {
	if filter.ParentID != nil {
		query = query.Where("parent_id = ?", *filter.ParentID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	return query
}

func (r *CategoryRepository) orderClause(page pagination.Params) string {
	column, ok := allowedCategorySorts[page.Sort]
	if !ok {
		column = allowedCategorySorts["sort_order"]
	}
	order := "ASC"
	if strings.EqualFold(page.Order, "desc") {
		order = "DESC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

func mapCategoryDBError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "slug") {
		return category.ErrSlugConflict
	}
	return err
}
