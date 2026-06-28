package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/productreview"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// ProductReviewRepository implements productreview.Repository using GORM.
type ProductReviewRepository struct {
	db *gorm.DB
}

// NewProductReviewRepository creates a new ProductReviewRepository.
func NewProductReviewRepository(db *gorm.DB) *ProductReviewRepository {
	return &ProductReviewRepository{db: db}
}

func (r *ProductReviewRepository) Create(ctx context.Context, rev *productreview.Review) error {
	m := toProductReviewModel(rev)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") ||
			strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return productreview.ErrAlreadyReviewed
		}
		return err
	}
	return nil
}

func (r *ProductReviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*productreview.Review, error) {
	var m models.ProductReviewModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, productreview.ErrNotFound
		}
		return nil, err
	}
	return toProductReviewDomain(&m), nil
}

func (r *ProductReviewRepository) ListByProduct(ctx context.Context, productID uuid.UUID, sort string, page pagination.Params) ([]productreview.Review, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.ProductReviewModel{}).
		Where("product_id = ? AND status = ?", productID, string(productreview.StatusApproved))

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC"
	switch sort {
	case "highest":
		orderClause = "rating DESC, created_at DESC"
	case "lowest":
		orderClause = "rating ASC, created_at DESC"
	}

	var items []models.ProductReviewModel
	err := query.Order(orderClause).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toProductReviewsDomain(items), total, nil
}

func (r *ProductReviewRepository) ListAdmin(ctx context.Context, filter productreview.ListFilter, page pagination.Params) ([]productreview.AdminListItem, int64, error) {
	base := r.db.WithContext(ctx).
		Table("product_reviews").
		Joins("LEFT JOIN products ON products.id = product_reviews.product_id")
	base = r.applyAdminFilters(base, filter)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		models.ProductReviewModel
		ProductName string `gorm:"column:product_name"`
	}

	var rows []row
	err := base.
		Select("product_reviews.*, products.name AS product_name").
		Order("product_reviews.created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]productreview.AdminListItem, len(rows))
	for i, rw := range rows {
		result[i] = productreview.AdminListItem{
			Review:      *toProductReviewDomain(&rw.ProductReviewModel),
			ProductName: rw.ProductName,
		}
	}
	return result, total, nil
}

func (r *ProductReviewRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status productreview.Status) error {
	result := r.db.WithContext(ctx).
		Model(&models.ProductReviewModel{}).
		Where("id = ?", id).
		Update("status", string(status))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return productreview.ErrNotFound
	}
	return nil
}

func (r *ProductReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ProductReviewModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return productreview.ErrNotFound
	}
	return nil
}

func (r *ProductReviewRepository) GetSummary(ctx context.Context, productID uuid.UUID) (*productreview.Summary, error) {
	type summaryRow struct {
		AvgRating float64 `gorm:"column:avg_rating"`
		Total     int64   `gorm:"column:total"`
	}

	var sr summaryRow
	err := r.db.WithContext(ctx).
		Model(&models.ProductReviewModel{}).
		Select("COALESCE(AVG(rating), 0) AS avg_rating, COUNT(*) AS total").
		Where("product_id = ? AND status = ?", productID, string(productreview.StatusApproved)).
		Scan(&sr).Error
	if err != nil {
		return nil, err
	}

	type distRow struct {
		Rating int   `gorm:"column:rating"`
		Count  int64 `gorm:"column:count"`
	}

	var distribution []distRow
	err = r.db.WithContext(ctx).
		Model(&models.ProductReviewModel{}).
		Select("rating, COUNT(*) AS count").
		Where("product_id = ? AND status = ?", productID, string(productreview.StatusApproved)).
		Group("rating").
		Scan(&distribution).Error
	if err != nil {
		return nil, err
	}

	dist := map[int]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	for _, d := range distribution {
		dist[d.Rating] = d.Count
	}

	return &productreview.Summary{
		AverageRating: sr.AvgRating,
		TotalCount:    sr.Total,
		Distribution:  dist,
	}, nil
}

func (r *ProductReviewRepository) ExistsByCustomer(ctx context.Context, productID, customerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProductReviewModel{}).
		Where("product_id = ? AND customer_id = ?", productID, customerID).
		Count(&count).Error
	return count > 0, err
}

func (r *ProductReviewRepository) applyAdminFilters(query *gorm.DB, filter productreview.ListFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("product_reviews.status = ?", filter.Status)
	}
	if filter.ProductID != nil {
		query = query.Where("product_reviews.product_id = ?", *filter.ProductID)
	}
	if filter.Rating != nil {
		query = query.Where("product_reviews.rating = ?", *filter.Rating)
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"LOWER(product_reviews.author_name) LIKE ? OR LOWER(product_reviews.content) LIKE ?",
			pattern, pattern,
		)
	}
	return query
}

func toProductReviewModel(rev *productreview.Review) *models.ProductReviewModel {
	m := &models.ProductReviewModel{
		ID:         rev.ID,
		ProductID:  rev.ProductID,
		CustomerID: rev.CustomerID,
		AuthorName: rev.AuthorName,
		Rating:     rev.Rating,
		Content:    rev.Content,
		Status:     string(rev.Status),
		CreatedAt:  rev.CreatedAt,
	}
	if rev.Title != "" {
		m.Title = &rev.Title
	}
	return m
}

func toProductReviewDomain(m *models.ProductReviewModel) *productreview.Review {
	rev := &productreview.Review{
		ID:         m.ID,
		ProductID:  m.ProductID,
		CustomerID: m.CustomerID,
		AuthorName: m.AuthorName,
		Rating:     m.Rating,
		Content:    m.Content,
		Status:     productreview.Status(m.Status),
		CreatedAt:  m.CreatedAt,
	}
	if m.Title != nil {
		rev.Title = *m.Title
	}
	return rev
}

func toProductReviewsDomain(items []models.ProductReviewModel) []productreview.Review {
	result := make([]productreview.Review, len(items))
	for i, m := range items {
		result[i] = *toProductReviewDomain(&m)
	}
	return result
}

// Compile-time check.
var _ productreview.Repository = (*ProductReviewRepository)(nil)

// Unexported helper used in orderClause patterns (not needed here but kept for consistency).
var _ = fmt.Sprintf
