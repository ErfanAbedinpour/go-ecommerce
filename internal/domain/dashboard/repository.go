package dashboard

import (
	"context"

	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

// Repository defines read-only dashboard analytics queries.
type Repository interface {
	GetStats(ctx context.Context) (*Stats, error)
	GetRevenueAnalytics(ctx context.Context, range_ DateRange) ([]RevenueDataPoint, error)
	ListLowStockProducts(ctx context.Context, page pagination.Params) ([]domainproduct.Product, int64, error)
	ListRecentOrders(ctx context.Context, limit int) ([]domainorder.DashboardSummary, error)
	ListFeaturedProducts(ctx context.Context, limit int) ([]domainproduct.Product, error)
}
