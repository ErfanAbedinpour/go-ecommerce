package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/productquestion"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// ProductQuestionRepository implements productquestion.Repository using GORM.
type ProductQuestionRepository struct {
	db *gorm.DB
}

// NewProductQuestionRepository creates a new ProductQuestionRepository.
func NewProductQuestionRepository(db *gorm.DB) *ProductQuestionRepository {
	return &ProductQuestionRepository{db: db}
}

func (r *ProductQuestionRepository) Create(ctx context.Context, q *productquestion.Question) error {
	m := toProductQuestionModel(q)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ProductQuestionRepository) FindByID(ctx context.Context, id uuid.UUID) (*productquestion.Question, error) {
	var m models.ProductQuestionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, productquestion.ErrNotFound
		}
		return nil, err
	}
	return toProductQuestionDomain(&m), nil
}

func (r *ProductQuestionRepository) ListByProduct(ctx context.Context, productID uuid.UUID, page pagination.Params) ([]productquestion.Question, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.ProductQuestionModel{}).
		Where("product_id = ? AND status = ?", productID, string(productquestion.StatusAnswered))

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.ProductQuestionModel
	err := query.Order("created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toProductQuestionsDomain(items), total, nil
}

func (r *ProductQuestionRepository) ListAdmin(ctx context.Context, filter productquestion.ListFilter, page pagination.Params) ([]productquestion.AdminListItem, int64, error) {
	base := r.db.WithContext(ctx).
		Table("product_questions").
		Joins("LEFT JOIN products ON products.id = product_questions.product_id")
	base = r.applyAdminFilters(base, filter)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		models.ProductQuestionModel
		ProductName string `gorm:"column:product_name"`
	}

	var rows []row
	err := base.
		Select("product_questions.*, products.name AS product_name").
		Order("product_questions.created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]productquestion.AdminListItem, len(rows))
	for i, rw := range rows {
		result[i] = productquestion.AdminListItem{
			Question:    *toProductQuestionDomain(&rw.ProductQuestionModel),
			ProductName: rw.ProductName,
		}
	}
	return result, total, nil
}

func (r *ProductQuestionRepository) Answer(ctx context.Context, id uuid.UUID, answer string, answeredBy uuid.UUID) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&models.ProductQuestionModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"answer":      answer,
			"answered_at": now,
			"answered_by": answeredBy,
			"status":      string(productquestion.StatusAnswered),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return productquestion.ErrNotFound
	}
	return nil
}

func (r *ProductQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ProductQuestionModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return productquestion.ErrNotFound
	}
	return nil
}

func (r *ProductQuestionRepository) applyAdminFilters(query *gorm.DB, filter productquestion.ListFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("product_questions.status = ?", filter.Status)
	}
	if filter.ProductID != nil {
		query = query.Where("product_questions.product_id = ?", *filter.ProductID)
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"LOWER(product_questions.question) LIKE ? OR LOWER(product_questions.asker_name) LIKE ?",
			pattern, pattern,
		)
	}
	return query
}

func toProductQuestionModel(q *productquestion.Question) *models.ProductQuestionModel {
	m := &models.ProductQuestionModel{
		ID:         q.ID,
		ProductID:  q.ProductID,
		AskerName:  q.AskerName,
		Question:   q.Question,
		Status:     string(q.Status),
		AnsweredAt: q.AnsweredAt,
		AnsweredBy: q.AnsweredBy,
		CreatedAt:  q.CreatedAt,
	}
	if q.AskerEmail != "" {
		m.AskerEmail = &q.AskerEmail
	}
	if q.Answer != "" {
		m.Answer = &q.Answer
	}
	return m
}

func toProductQuestionDomain(m *models.ProductQuestionModel) *productquestion.Question {
	q := &productquestion.Question{
		ID:         m.ID,
		ProductID:  m.ProductID,
		AskerName:  m.AskerName,
		Question:   m.Question,
		Status:     productquestion.Status(m.Status),
		AnsweredAt: m.AnsweredAt,
		AnsweredBy: m.AnsweredBy,
		CreatedAt:  m.CreatedAt,
	}
	if m.AskerEmail != nil {
		q.AskerEmail = *m.AskerEmail
	}
	if m.Answer != nil {
		q.Answer = *m.Answer
	}
	return q
}

func toProductQuestionsDomain(items []models.ProductQuestionModel) []productquestion.Question {
	result := make([]productquestion.Question, len(items))
	for i, m := range items {
		result[i] = *toProductQuestionDomain(&m)
	}
	return result
}

var _ productquestion.Repository = (*ProductQuestionRepository)(nil)
