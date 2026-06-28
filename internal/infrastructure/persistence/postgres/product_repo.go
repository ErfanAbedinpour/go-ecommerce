package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/product"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

var allowedProductSorts = map[string]string{
	"created_at": "products.created_at",
	"name":       "products.name",
	"price":      "products.price",
	"updated_at": "products.updated_at",
}

// ProductRepository implements product.Repository using GORM.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := toProductModel(p)
		if err := tx.Omit("Images", "Attributes", "SKUs", "Inventory").Create(m).Error; err != nil {
			return mapProductDBError(err)
		}
		if err := r.saveChildren(tx, p); err != nil {
			return err
		}
		return nil
	})
}

func (r *ProductRepository) Update(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := toProductModel(p)
		if err := tx.Model(&models.ProductModel{}).
			Where("id = ?", p.ID).
			Updates(map[string]any{
				"category_id":       m.CategoryID,
				"name":              m.Name,
				"slug":              m.Slug,
				"description":       m.Description,
				"short_description": m.ShortDescription,
				"price":             m.Price,
				"sale_price":        m.SalePrice,
				"brand":             m.Brand,
				"is_featured":       m.IsFeatured,
				"status":            m.Status,
				"updated_at":        p.UpdatedAt,
			}).Error; err != nil {
			return mapProductDBError(err)
		}

		if err := tx.Where("product_id = ?", p.ID).Delete(&models.ProductImageModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("product_id = ?", p.ID).Delete(&models.ProductAttributeModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("product_id = ?", p.ID).Delete(&models.SkuModel{}).Error; err != nil {
			return err
		}
		return r.saveChildren(tx, p)
	})
}

func (r *ProductRepository) saveChildren(tx *gorm.DB, p *product.Product) error {
	if len(p.Images) > 0 {
		images := make([]models.ProductImageModel, len(p.Images))
		for i, img := range p.Images {
			images[i] = toImageModel(img)
		}
		if err := tx.Create(&images).Error; err != nil {
			return err
		}
	}

	if len(p.Attributes) > 0 {
		attrs := make([]models.ProductAttributeModel, len(p.Attributes))
		for i, attr := range p.Attributes {
			attrs[i] = models.ProductAttributeModel{
				ID:        attr.ID,
				ProductID: attr.ProductID,
				Name:      attr.Name,
			}
		}
		if err := tx.Select("ID", "ProductID", "Name").Create(&attrs).Error; err != nil {
			return err
		}

		var values []models.ProductAttributeValueModel
		for _, attr := range p.Attributes {
			for _, v := range attr.Values {
				values = append(values, models.ProductAttributeValueModel{
					ID:          v.ID,
					AttributeID: attr.ID,
					Value:       v.Value,
				})
			}
		}
		if len(values) > 0 {
			if err := tx.Select("ID", "AttributeID", "Value").Create(&values).Error; err != nil {
				return err
			}
		}
	}

	if len(p.SKUs) > 0 {
		skus := make([]models.SkuModel, len(p.SKUs))
		for i, sku := range p.SKUs {
			skus[i] = toSkuModel(sku)
		}
		if err := tx.Create(&skus).Error; err != nil {
			return err
		}
	}

	inv := models.InventoryModel{
		ID:                p.Inventory.ID,
		ProductID:         p.Inventory.ProductID,
		Quantity:          p.Inventory.Quantity,
		LowStockThreshold: p.Inventory.LowStockThreshold,
		UpdatedAt:         p.Inventory.UpdatedAt,
	}
	return tx.Save(&inv).Error
}

func (r *ProductRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.ProductModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return product.ErrNotFound
	}
	return nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	var m models.ProductModel
	err := r.db.WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Attributes.Values").
		Preload("SKUs").
		Preload("Inventory").
		Where("id = ?", id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	return toProductDomain(&m), nil
}

func (r *ProductRepository) FindBySlug(ctx context.Context, slug string) (*product.Product, error) {
	var m models.ProductModel
	err := r.db.WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Attributes.Values").
		Preload("SKUs").
		Preload("Inventory").
		Where("slug = ?", slug).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	return toProductDomain(&m), nil
}

func (r *ProductRepository) FindBySKU(ctx context.Context, skuCode string) (*product.Product, error) {
	var sku models.SkuModel
	err := r.db.WithContext(ctx).Where("code = ?", skuCode).First(&sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	return r.FindByID(ctx, sku.ProductID)
}

func (r *ProductRepository) List(ctx context.Context, filter product.ListFilter, page pagination.Params) ([]product.Product, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.ProductModel{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.Brand != "" {
		query = query.Where("brand = ?", filter.Brand)
	}
	if filter.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filter.IsFeatured)
	}
	query = r.applyStockLevelFilter(query, filter.StockLevel)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.ProductModel
	err := query.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Attributes.Values").
		Preload("SKUs").
		Preload("Inventory").
		Order(r.orderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return toProductsDomain(items), total, nil
}

func (r *ProductRepository) Search(ctx context.Context, query string, page pagination.Params) ([]product.Product, int64, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	db := r.db.WithContext(ctx).Model(&models.ProductModel{}).
		Where(
			"LOWER(products.name) LIKE ? OR LOWER(COALESCE(products.description, '')) LIKE ? OR EXISTS (SELECT 1 FROM skus WHERE skus.product_id = products.id AND LOWER(skus.code) LIKE ?)",
			pattern, pattern, pattern,
		)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.ProductModel
	err := db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Attributes.Values").
		Preload("SKUs").
		Preload("Inventory").
		Order("products.created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return toProductsDomain(items), total, nil
}

func (r *ProductRepository) ListStorefront(ctx context.Context, filter product.StoreListFilter, page pagination.Params) ([]product.Product, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.ProductModel{}).
		Where("status = ?", product.StatusActive.String())

	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"LOWER(products.name) LIKE ? OR LOWER(COALESCE(products.description, '')) LIKE ? OR LOWER(COALESCE(products.brand, '')) LIKE ? OR EXISTS (SELECT 1 FROM skus WHERE skus.product_id = products.id AND LOWER(skus.code) LIKE ?)",
			pattern, pattern, pattern, pattern,
		)
	}

	switch filter.Sort {
	case "discount":
		query = query.Where("sale_price IS NOT NULL AND sale_price < price").
			Order("(1 - sale_price / NULLIF(price, 0)) DESC")
	case "price":
		query = query.Order("COALESCE(sale_price, price) ASC")
	case "name":
		query = query.Order("products.name ASC")
	default:
		query = query.Order("products.created_at DESC")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.ProductModel
	err := query.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Attributes.Values").
		Preload("SKUs").
		Preload("Inventory").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return toProductsDomain(items), total, nil
}

func (r *ProductRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProductModel{}).
		Where("status = ?", product.StatusActive.String()).
		Count(&count).Error
	return count, err
}

func (r *ProductRepository) UpdateInventory(ctx context.Context, productID uuid.UUID, inventory product.Inventory) error {
	result := r.db.WithContext(ctx).
		Model(&models.InventoryModel{}).
		Where("product_id = ?", productID).
		Updates(map[string]any{
			"quantity":            inventory.Quantity,
			"low_stock_threshold": inventory.LowStockThreshold,
			"updated_at":          inventory.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return product.ErrNotFound
	}
	return nil
}

func (r *ProductRepository) ExistsInActiveOrders(ctx context.Context, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("order_items").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("order_items.product_id = ?", productID).
		Where("orders.status NOT IN ?", []string{"cancelled", "refunded", "delivered"}).
		Count(&count).Error
	return count > 0, err
}

func (r *ProductRepository) GetStats(ctx context.Context) (*product.Stats, error) {
	var stats product.Stats
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM products WHERE deleted_at IS NULL) AS total,
			(SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND status = 'active') AS active,
			(SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND status = 'draft') AS draft,
			COALESCE((
				SELECT COUNT(*)
				FROM inventories i
				INNER JOIN products p ON p.id = i.product_id
				WHERE p.deleted_at IS NULL AND i.quantity = 0
			), 0) AS out_of_stock
	`).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *ProductRepository) applyStockLevelFilter(query *gorm.DB, stockLevel string) *gorm.DB {
	switch stockLevel {
	case "low":
		return query.
			Joins("INNER JOIN inventories ON inventories.product_id = products.id").
			Where("inventories.quantity > 0").
			Where("inventories.quantity <= inventories.low_stock_threshold")
	case "out":
		return query.
			Joins("INNER JOIN inventories ON inventories.product_id = products.id").
			Where("inventories.quantity = 0")
	default:
		return query
	}
}

func (r *ProductRepository) CategoryExists(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CategoryModel{}).
		Where("id = ?", categoryID).
		Count(&count).Error
	return count > 0, err
}

func (r *ProductRepository) orderClause(page pagination.Params) string {
	column, ok := allowedProductSorts[page.Sort]
	if !ok {
		column = allowedProductSorts["created_at"]
	}
	order := "DESC"
	if strings.EqualFold(page.Order, "asc") {
		order = "ASC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

func mapProductDBError(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return product.ErrSlugConflict
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "products_slug") || strings.Contains(msg, "slug") {
		return product.ErrSlugConflict
	}
	if strings.Contains(msg, "skus_code") || strings.Contains(msg, "code") {
		return product.ErrSKUConflict
	}
	return err
}
