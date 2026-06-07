package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/attributevalue"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// AttributeValueRepository implements attributevalue.Repository using GORM.
type AttributeValueRepository struct {
	db *gorm.DB
}

// NewAttributeValueRepository creates a new AttributeValueRepository.
func NewAttributeValueRepository(db *gorm.DB) *AttributeValueRepository {
	return &AttributeValueRepository{db: db}
}

func (r *AttributeValueRepository) Create(ctx context.Context, v *attributevalue.Value) error {
	m := toAttributeValueModel(v)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return mapAttributeValueDBError(err)
	}
	return nil
}

func (r *AttributeValueRepository) Update(ctx context.Context, v *attributevalue.Value) error {
	result := r.db.WithContext(ctx).Model(&models.AttributeValueModel{}).
		Where("id = ?", v.ID).
		Updates(map[string]any{
			"value":      v.Value,
			"sort_order": v.SortOrder,
			"is_active":  v.IsActive,
			"updated_at": v.UpdatedAt,
		})
	if result.Error != nil {
		return mapAttributeValueDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return attributevalue.ErrNotFound
	}
	return nil
}

func (r *AttributeValueRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.AttributeValueModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return attributevalue.ErrNotFound
	}
	return nil
}

func (r *AttributeValueRepository) FindByID(ctx context.Context, id uuid.UUID) (*attributevalue.Value, error) {
	var m models.AttributeValueModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, attributevalue.ErrNotFound
		}
		return nil, err
	}
	return toAttributeValueDomain(&m), nil
}

func (r *AttributeValueRepository) List(ctx context.Context, filter attributevalue.ListFilter, page pagination.Params) ([]attributevalue.Value, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.AttributeValueModel{})
	if filter.AttributeID != nil {
		query = query.Where("attribute_id = ?", *filter.AttributeID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.AttributeValueModel
	err := query.Order(attributeValueOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toAttributeValuesDomain(items), total, nil
}

func (r *AttributeValueRepository) AttributeExists(ctx context.Context, attributeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.AttributeDefinitionModel{}).
		Where("id = ?", attributeID).
		Count(&count).Error
	return count > 0, err
}

func attributeValueOrderClause(page pagination.Params) string {
	allowed := map[string]string{
		"created_at": "created_at",
		"value":      "value",
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

func toAttributeValueModel(v *attributevalue.Value) *models.AttributeValueModel {
	return &models.AttributeValueModel{
		ID:          v.ID,
		AttributeID: v.AttributeID,
		Value:       v.Value,
		SortOrder:   v.SortOrder,
		IsActive:    v.IsActive,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func toAttributeValueDomain(m *models.AttributeValueModel) *attributevalue.Value {
	return &attributevalue.Value{
		ID:          m.ID,
		AttributeID: m.AttributeID,
		Value:       m.Value,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toAttributeValuesDomain(items []models.AttributeValueModel) []attributevalue.Value {
	result := make([]attributevalue.Value, len(items))
	for i, m := range items {
		result[i] = *toAttributeValueDomain(&m)
	}
	return result
}

func mapAttributeValueDBError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "value") || strings.Contains(msg, "unique") {
		return attributevalue.ErrValueConflict
	}
	return err
}
