package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/attributedef"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// AttributeDefinitionRepository implements attributedef.Repository using GORM.
type AttributeDefinitionRepository struct {
	db *gorm.DB
}

// NewAttributeDefinitionRepository creates a new AttributeDefinitionRepository.
func NewAttributeDefinitionRepository(db *gorm.DB) *AttributeDefinitionRepository {
	return &AttributeDefinitionRepository{db: db}
}

func (r *AttributeDefinitionRepository) Create(ctx context.Context, d *attributedef.Definition) error {
	m := toAttributeDefModel(d)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return mapAttributeDefDBError(err)
	}
	return nil
}

func (r *AttributeDefinitionRepository) Update(ctx context.Context, d *attributedef.Definition) error {
	result := r.db.WithContext(ctx).Model(&models.AttributeDefinitionModel{}).
		Where("id = ?", d.ID).
		Updates(map[string]any{
			"name":       d.Name,
			"slug":       d.Slug,
			"sort_order": d.SortOrder,
			"is_active":  d.IsActive,
			"updated_at": d.UpdatedAt,
		})
	if result.Error != nil {
		return mapAttributeDefDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return attributedef.ErrNotFound
	}
	return nil
}

func (r *AttributeDefinitionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.AttributeDefinitionModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return attributedef.ErrNotFound
	}
	return nil
}

func (r *AttributeDefinitionRepository) FindByID(ctx context.Context, id uuid.UUID) (*attributedef.Definition, error) {
	var m models.AttributeDefinitionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, attributedef.ErrNotFound
		}
		return nil, err
	}
	return toAttributeDefDomain(&m), nil
}

func (r *AttributeDefinitionRepository) FindBySlug(ctx context.Context, slug string) (*attributedef.Definition, error) {
	var m models.AttributeDefinitionModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, attributedef.ErrNotFound
		}
		return nil, err
	}
	return toAttributeDefDomain(&m), nil
}

func (r *AttributeDefinitionRepository) List(ctx context.Context, filter attributedef.ListFilter, page pagination.Params) ([]attributedef.Definition, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.AttributeDefinitionModel{})
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.AttributeDefinitionModel
	err := query.Order(attributeDefOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toAttributeDefsDomain(items), total, nil
}

func (r *AttributeDefinitionRepository) HasValues(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.AttributeValueModel{}).
		Where("attribute_id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func attributeDefOrderClause(page pagination.Params) string {
	allowed := map[string]string{
		"created_at": "created_at",
		"name":       "name",
		"sort_order": "sort_order",
		"updated_at": "updated_at",
	}
	column, ok := allowed[page.Sort]
	if !ok {
		column = "sort_order"
	}
	order := "ASC"
	if strings.EqualFold(page.Order, "desc") {
		order = "DESC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

func toAttributeDefModel(d *attributedef.Definition) *models.AttributeDefinitionModel {
	return &models.AttributeDefinitionModel{
		ID:        d.ID,
		Name:      d.Name,
		Slug:      d.Slug,
		SortOrder: d.SortOrder,
		IsActive:  d.IsActive,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func toAttributeDefDomain(m *models.AttributeDefinitionModel) *attributedef.Definition {
	return &attributedef.Definition{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toAttributeDefsDomain(items []models.AttributeDefinitionModel) []attributedef.Definition {
	result := make([]attributedef.Definition, len(items))
	for i, m := range items {
		result[i] = *toAttributeDefDomain(&m)
	}
	return result
}

func mapAttributeDefDBError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "slug") {
		return attributedef.ErrSlugConflict
	}
	if strings.Contains(msg, "name") {
		return attributedef.ErrNameConflict
	}
	return err
}
