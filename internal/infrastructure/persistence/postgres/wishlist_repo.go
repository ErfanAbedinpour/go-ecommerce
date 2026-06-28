package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/wishlist"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// WishlistRepository implements wishlist.Repository using GORM.
type WishlistRepository struct {
	db *gorm.DB
}

// NewWishlistRepository creates a new WishlistRepository.
func NewWishlistRepository(db *gorm.DB) *WishlistRepository {
	return &WishlistRepository{db: db}
}

func (r *WishlistRepository) Add(ctx context.Context, item *wishlist.Item) error {
	m := &models.WishlistItemModel{
		ID:         item.ID,
		CustomerID: item.CustomerID,
		ProductID:  item.ProductID,
		CreatedAt:  item.CreatedAt,
	}
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			return wishlist.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *WishlistRepository) Remove(ctx context.Context, customerID, productID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("customer_id = ? AND product_id = ?", customerID, productID).
		Delete(&models.WishlistItemModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return wishlist.ErrNotFound
	}
	return nil
}

func (r *WishlistRepository) List(ctx context.Context, customerID uuid.UUID, page pagination.Params) ([]wishlist.ListItem, int64, error) {
	base := r.db.WithContext(ctx).
		Table("wishlist_items").
		Where("wishlist_items.customer_id = ?", customerID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		ID         uuid.UUID `gorm:"column:id"`
		CustomerID uuid.UUID `gorm:"column:customer_id"`
		ProductID  uuid.UUID `gorm:"column:product_id"`
		CreatedAt  string    `gorm:"column:created_at"`
		// Product fields
		ProductName  string   `gorm:"column:product_name"`
		ProductSlug  string   `gorm:"column:product_slug"`
		ProductPrice float64  `gorm:"column:product_price"`
		SalePrice    *float64 `gorm:"column:sale_price"`
		ImageURL     *string  `gorm:"column:image_url"`
		StockQty     int      `gorm:"column:stock_qty"`
	}

	var rows []row
	err := base.
		Select(`
			wishlist_items.id,
			wishlist_items.customer_id,
			wishlist_items.product_id,
			wishlist_items.created_at,
			products.name AS product_name,
			products.slug AS product_slug,
			products.price AS product_price,
			products.sale_price AS sale_price,
			(SELECT url FROM product_images WHERE product_images.product_id = products.id ORDER BY sort_order LIMIT 1) AS image_url,
			COALESCE(inventories.quantity, 0) AS stock_qty
		`).
		Joins("INNER JOIN products ON products.id = wishlist_items.product_id AND products.deleted_at IS NULL").
		Joins("LEFT JOIN inventories ON inventories.product_id = products.id").
		Order("wishlist_items.created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]wishlist.ListItem, len(rows))
	for i, rw := range rows {
		result[i] = wishlist.ListItem{
			Item: wishlist.Item{
				ID:         rw.ID,
				CustomerID: rw.CustomerID,
				ProductID:  rw.ProductID,
			},
			Product: wishlist.ProductSummary{
				Name:      rw.ProductName,
				Slug:      rw.ProductSlug,
				Price:     rw.ProductPrice,
				SalePrice: rw.SalePrice,
				IsInStock: rw.StockQty > 0,
			},
		}
		if rw.ImageURL != nil {
			result[i].Product.ImageURL = *rw.ImageURL
		}
	}
	return result, total, nil
}

func (r *WishlistRepository) Exists(ctx context.Context, customerID, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.WishlistItemModel{}).
		Where("customer_id = ? AND product_id = ?", customerID, productID).
		Count(&count).Error
	return count > 0, err
}

func (r *WishlistRepository) BatchCheck(ctx context.Context, customerID uuid.UUID, productIDs []uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&models.WishlistItemModel{}).
		Where("customer_id = ? AND product_id IN ?", customerID, productIDs).
		Pluck("product_id", &ids).Error
	return ids, err
}

var _ wishlist.Repository = (*WishlistRepository)(nil)
