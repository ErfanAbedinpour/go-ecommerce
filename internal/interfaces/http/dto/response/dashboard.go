package response

import (
	domain "app/internal/domain/dashboard"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

// DashboardStatsResponse holds aggregated dashboard KPIs.
type DashboardStatsResponse struct {
	TotalRevenue   float64 `json:"total_revenue"`
	TotalOrders    int64   `json:"total_orders"`
	TotalCustomers int64   `json:"total_customers"`
	TotalProducts  int64   `json:"total_products"`
	PendingOrders  int64   `json:"pending_orders"`
	LowStockCount  int64   `json:"low_stock_count"`
}

// RevenueDataPointResponse is a single revenue analytics entry.
type RevenueDataPointResponse struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
}

// RevenueAnalyticsResponse is the revenue time-series response.
type RevenueAnalyticsResponse struct {
	Data []RevenueDataPointResponse `json:"data"`
}

// RecentOrdersResponse is the recent orders feed response.
type RecentOrdersResponse struct {
	Data []OrderSummaryResponse `json:"data"`
}

// ToDashboardStatsResponse maps domain stats to API response.
func ToDashboardStatsResponse(s *domain.Stats) DashboardStatsResponse {
	return DashboardStatsResponse{
		TotalRevenue:   s.TotalRevenue,
		TotalOrders:    s.TotalOrders,
		TotalCustomers: s.TotalCustomers,
		TotalProducts:  s.TotalProducts,
		PendingOrders:  s.PendingOrders,
		LowStockCount:  s.LowStockCount,
	}
}

// ToRevenueAnalyticsResponse maps revenue data points to API response.
func ToRevenueAnalyticsResponse(points []domain.RevenueDataPoint) RevenueAnalyticsResponse {
	items := make([]RevenueDataPointResponse, len(points))
	for i, p := range points {
		items[i] = RevenueDataPointResponse{
			Date:    p.Date,
			Revenue: p.Revenue,
			Orders:  p.Orders,
		}
	}
	return RevenueAnalyticsResponse{Data: items}
}

// ToLowStockProductListResponse maps low-stock products to a paginated API response.
func ToLowStockProductListResponse(result pagination.Paginated[domainproduct.Product]) ProductListResponse {
	return ToProductListResponse(result)
}

// ToRecentOrdersResponse maps recent order summaries to API response.
func ToRecentOrdersResponse(orders []domainorder.Summary) RecentOrdersResponse {
	items := make([]OrderSummaryResponse, len(orders))
	for i, o := range orders {
		items[i] = ToOrderSummaryResponse(o)
	}
	return RecentOrdersResponse{Data: items}
}
