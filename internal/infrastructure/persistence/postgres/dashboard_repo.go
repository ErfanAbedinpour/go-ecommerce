package postgres

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"app/internal/domain/dashboard"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// DashboardRepository implements dashboard.Repository using GORM and raw SQL.
type DashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a new DashboardRepository.
func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetStats(ctx context.Context) (*dashboard.Stats, error) {
	var stats dashboard.Stats

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE((
				SELECT SUM(total)
				FROM orders
				WHERE payment_status = 'paid'
			), 0) AS total_revenue,
			(SELECT COUNT(*) FROM orders) AS total_orders,
			(SELECT COUNT(*) FROM customers) AS total_customers,
			(SELECT COUNT(*) FROM products WHERE deleted_at IS NULL) AS total_products,
			(SELECT COUNT(*) FROM orders WHERE status = 'pending') AS pending_orders,
			COALESCE((
				SELECT COUNT(*)
				FROM inventories i
				INNER JOIN products p ON p.id = i.product_id
				WHERE p.deleted_at IS NULL
				  AND i.quantity <= i.low_stock_threshold
			), 0) AS low_stock_count
	`).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *DashboardRepository) GetRevenueAnalytics(ctx context.Context, range_ dashboard.DateRange) ([]dashboard.RevenueDataPoint, error) {
	truncExpr, dateFormat := revenueTruncExpr(range_.Granularity)

	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(%s, '%s') AS date,
			COALESCE(SUM(total) FILTER (WHERE payment_status = 'paid'), 0) AS revenue,
			COUNT(*) AS orders
		FROM orders
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY %s
		ORDER BY %s ASC
	`, truncExpr, dateFormat, truncExpr, truncExpr)

	var points []dashboard.RevenueDataPoint
	err := r.db.WithContext(ctx).Raw(query, range_.From, range_.To).Scan(&points).Error
	if err != nil {
		return nil, err
	}
	if points == nil {
		points = []dashboard.RevenueDataPoint{}
	}
	return points, nil
}

func revenueTruncExpr(g dashboard.Granularity) (truncExpr, dateFormat string) {
	switch g {
	case dashboard.GranularityHour:
		return "date_trunc('hour', created_at)", "YYYY-MM-DD HH24:00"
	case dashboard.GranularityMonth:
		return "date_trunc('month', created_at)", "YYYY-MM"
	default:
		return "date_trunc('day', created_at)", "YYYY-MM-DD"
	}
}

func (r *DashboardRepository) ListLowStockProducts(ctx context.Context, page pagination.Params) ([]domainproduct.Product, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.ProductModel{}).
		Joins("INNER JOIN inventories ON inventories.product_id = products.id").
		Where("inventories.quantity <= inventories.low_stock_threshold")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.ProductModel
	err := query.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Attributes").
		Preload("Inventory").
		Order(r.lowStockOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toProductsDomain(items), total, nil
}

func (r *DashboardRepository) ListRecentOrders(ctx context.Context, limit int) ([]domainorder.Summary, error) {
	type orderRow struct {
		models.OrderModel
		ItemCount int64 `gorm:"column:item_count"`
	}

	var rows []orderRow
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("orders.*, COUNT(order_items.id) AS item_count").
		Joins("LEFT JOIN order_items ON order_items.order_id = orders.id").
		Group("orders.id").
		Order("orders.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]domainorder.Summary, len(rows))
	for i, row := range rows {
		result[i] = domainorder.Summary{
			ID:            row.ID,
			OrderNumber:   row.OrderNumber,
			Status:        row.Status,
			PaymentStatus: row.PaymentStatus,
			Total:         row.Total,
			ItemCount:     int(row.ItemCount),
			CreatedAt:     row.CreatedAt,
		}
	}
	return result, nil
}

func (r *DashboardRepository) lowStockOrderClause(page pagination.Params) string {
	allowed := map[string]string{
		"created_at": "products.created_at",
		"name":       "products.name",
		"quantity":   "inventories.quantity",
	}
	column, ok := allowed[page.Sort]
	if !ok {
		column = "inventories.quantity"
	}
	order := "ASC"
	if strings.EqualFold(page.Order, "desc") {
		order = "DESC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

var _ dashboard.Repository = (*DashboardRepository)(nil)
