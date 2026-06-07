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

type statsRow struct {
	dashboard.Stats
	RevenueCurrent30d    float64 `gorm:"column:revenue_current_30d"`
	RevenuePrevious30d   float64 `gorm:"column:revenue_previous_30d"`
	OrdersCurrent30d     int64   `gorm:"column:orders_current_30d"`
	OrdersPrevious30d    int64   `gorm:"column:orders_previous_30d"`
	CustomersCurrent30d  int64   `gorm:"column:customers_current_30d"`
	CustomersPrevious30d int64   `gorm:"column:customers_previous_30d"`
	ProductsCurrent30d   int64   `gorm:"column:products_current_30d"`
	ProductsPrevious30d  int64   `gorm:"column:products_previous_30d"`
	PendingCurrent30d    int64   `gorm:"column:pending_current_30d"`
	PendingPrevious30d   int64   `gorm:"column:pending_previous_30d"`
	LowStockCurrent30d   int64   `gorm:"column:low_stock_current_30d"`
	LowStockPrevious30d  int64   `gorm:"column:low_stock_previous_30d"`
}

func (r *DashboardRepository) GetStats(ctx context.Context) (*dashboard.Stats, error) {
	var row statsRow

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
			), 0) AS low_stock_count,
			COALESCE((
				SELECT SUM(total)
				FROM orders
				WHERE payment_status = 'paid'
				  AND created_at >= NOW() - INTERVAL '30 days'
			), 0) AS revenue_current_30d,
			COALESCE((
				SELECT SUM(total)
				FROM orders
				WHERE payment_status = 'paid'
				  AND created_at >= NOW() - INTERVAL '60 days'
				  AND created_at < NOW() - INTERVAL '30 days'
			), 0) AS revenue_previous_30d,
			(SELECT COUNT(*) FROM orders
			 WHERE created_at >= NOW() - INTERVAL '30 days') AS orders_current_30d,
			(SELECT COUNT(*) FROM orders
			 WHERE created_at >= NOW() - INTERVAL '60 days'
			   AND created_at < NOW() - INTERVAL '30 days') AS orders_previous_30d,
			(SELECT COUNT(*) FROM customers
			 WHERE created_at >= NOW() - INTERVAL '30 days') AS customers_current_30d,
			(SELECT COUNT(*) FROM customers
			 WHERE created_at >= NOW() - INTERVAL '60 days'
			   AND created_at < NOW() - INTERVAL '30 days') AS customers_previous_30d,
			(SELECT COUNT(*) FROM products
			 WHERE deleted_at IS NULL
			   AND created_at >= NOW() - INTERVAL '30 days') AS products_current_30d,
			(SELECT COUNT(*) FROM products
			 WHERE deleted_at IS NULL
			   AND created_at >= NOW() - INTERVAL '60 days'
			   AND created_at < NOW() - INTERVAL '30 days') AS products_previous_30d,
			(SELECT COUNT(*) FROM orders
			 WHERE status = 'pending'
			   AND created_at >= NOW() - INTERVAL '30 days') AS pending_current_30d,
			(SELECT COUNT(*) FROM orders
			 WHERE status = 'pending'
			   AND created_at >= NOW() - INTERVAL '60 days'
			   AND created_at < NOW() - INTERVAL '30 days') AS pending_previous_30d,
			COALESCE((
				SELECT COUNT(*)
				FROM inventories i
				INNER JOIN products p ON p.id = i.product_id
				WHERE p.deleted_at IS NULL
				  AND i.quantity <= i.low_stock_threshold
				  AND p.updated_at >= NOW() - INTERVAL '30 days'
			), 0) AS low_stock_current_30d,
			COALESCE((
				SELECT COUNT(*)
				FROM inventories i
				INNER JOIN products p ON p.id = i.product_id
				WHERE p.deleted_at IS NULL
				  AND i.quantity <= i.low_stock_threshold
				  AND p.updated_at >= NOW() - INTERVAL '60 days'
				  AND p.updated_at < NOW() - INTERVAL '30 days'
			), 0) AS low_stock_previous_30d
	`).Scan(&row).Error
	if err != nil {
		return nil, err
	}

	stats := row.Stats
	stats.Growth = dashboard.StatsGrowth{
		TotalRevenue:   dashboard.CalcGrowthPercent(row.RevenueCurrent30d, row.RevenuePrevious30d),
		TotalOrders:    dashboard.CalcGrowthPercent(float64(row.OrdersCurrent30d), float64(row.OrdersPrevious30d)),
		TotalCustomers: dashboard.CalcGrowthPercent(float64(row.CustomersCurrent30d), float64(row.CustomersPrevious30d)),
		TotalProducts:  dashboard.CalcGrowthPercent(float64(row.ProductsCurrent30d), float64(row.ProductsPrevious30d)),
		PendingOrders:  dashboard.CalcGrowthPercent(float64(row.PendingCurrent30d), float64(row.PendingPrevious30d)),
		LowStockCount:  dashboard.CalcGrowthPercent(float64(row.LowStockCurrent30d), float64(row.LowStockPrevious30d)),
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

func (r *DashboardRepository) ListRecentOrders(ctx context.Context, limit int) ([]domainorder.DashboardSummary, error) {
	type orderRow struct {
		models.OrderModel
		ItemCount    int64  `gorm:"column:item_count"`
		CustomerName string `gorm:"column:customer_name"`
		ProductName  string `gorm:"column:product_name"`
	}

	var rows []orderRow
	err := r.db.WithContext(ctx).
		Table("orders").
		Select(`
			orders.*,
			COUNT(order_items.id) AS item_count,
			TRIM(CONCAT(customers.first_name, ' ', customers.last_name)) AS customer_name,
			(
				SELECT oi.product_name
				FROM order_items oi
				WHERE oi.order_id = orders.id
				ORDER BY oi.id ASC
				LIMIT 1
			) AS product_name
		`).
		Joins("LEFT JOIN customers ON customers.id = orders.customer_id").
		Joins("LEFT JOIN order_items ON order_items.order_id = orders.id").
		Group("orders.id, customers.first_name, customers.last_name").
		Order("orders.created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]domainorder.DashboardSummary, len(rows))
	for i, row := range rows {
		result[i] = domainorder.DashboardSummary{
			Summary: domainorder.Summary{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				Status:        row.Status,
				PaymentStatus: row.PaymentStatus,
				Total:         row.Total,
				ItemCount:     int(row.ItemCount),
				CreatedAt:     row.CreatedAt,
			},
			CustomerName: row.CustomerName,
			ProductName:  row.ProductName,
		}
	}
	return result, nil
}

func (r *DashboardRepository) ListFeaturedProducts(ctx context.Context, limit int) ([]domainproduct.Product, error) {
	var items []models.ProductModel
	err := r.db.WithContext(ctx).
		Where("is_featured = ? AND status = ? AND deleted_at IS NULL", true, "active").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Attributes").
		Preload("Inventory").
		Order("updated_at DESC").
		Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toProductsDomain(items), nil
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
