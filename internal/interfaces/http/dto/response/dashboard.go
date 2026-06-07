package response

import (
	"time"

	domain "app/internal/domain/dashboard"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

// DashboardStatsGrowthResponse holds period-over-period growth percentages.
type DashboardStatsGrowthResponse struct {
	TotalRevenue   float64 `json:"total_revenue"`
	TotalOrders    float64 `json:"total_orders"`
	TotalCustomers float64 `json:"total_customers"`
	TotalProducts  float64 `json:"total_products"`
	PendingOrders  float64 `json:"pending_orders"`
	LowStockCount  float64 `json:"low_stock_count"`
}

// DashboardStatsResponse holds aggregated dashboard KPIs.
type DashboardStatsResponse struct {
	TotalRevenue   float64                      `json:"total_revenue"`
	TotalOrders    int64                        `json:"total_orders"`
	TotalCustomers int64                        `json:"total_customers"`
	TotalProducts  int64                        `json:"total_products"`
	PendingOrders  int64                        `json:"pending_orders"`
	LowStockCount  int64                        `json:"low_stock_count"`
	Growth         DashboardStatsGrowthResponse `json:"growth"`
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

// RecentOrderResponse is a dashboard recent-order row with customer and product context.
type RecentOrderResponse struct {
	ID            string    `json:"id"`
	OrderNumber   string    `json:"order_number"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	Total         float64   `json:"total"`
	ItemCount     int       `json:"item_count"`
	CustomerName  string    `json:"customer_name"`
	ProductName   string    `json:"product_name"`
	CreatedAt     time.Time `json:"created_at"`
}

// RecentOrdersResponse is the recent orders feed response.
type RecentOrdersResponse struct {
	Data []RecentOrderResponse `json:"data"`
}

// FeaturedProductsResponse is the featured products widget response.
type FeaturedProductsResponse struct {
	Data []ProductResponse `json:"data"`
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
		Growth: DashboardStatsGrowthResponse{
			TotalRevenue:   s.Growth.TotalRevenue,
			TotalOrders:    s.Growth.TotalOrders,
			TotalCustomers: s.Growth.TotalCustomers,
			TotalProducts:  s.Growth.TotalProducts,
			PendingOrders:  s.Growth.PendingOrders,
			LowStockCount:  s.Growth.LowStockCount,
		},
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
func ToRecentOrdersResponse(orders []domainorder.DashboardSummary) RecentOrdersResponse {
	items := make([]RecentOrderResponse, len(orders))
	for i, o := range orders {
		items[i] = RecentOrderResponse{
			ID:            o.ID.String(),
			OrderNumber:   o.OrderNumber,
			Status:        o.Status,
			PaymentStatus: o.PaymentStatus,
			Total:         o.Total,
			ItemCount:     o.ItemCount,
			CustomerName:  o.CustomerName,
			ProductName:   o.ProductName,
			CreatedAt:     o.CreatedAt,
		}
	}
	return RecentOrdersResponse{Data: items}
}

// ToFeaturedProductsResponse maps featured products to API response.
func ToFeaturedProductsResponse(products []domainproduct.Product) FeaturedProductsResponse {
	items := make([]ProductResponse, len(products))
	for i, p := range products {
		items[i] = ToProductResponse(&p)
	}
	return FeaturedProductsResponse{Data: items}
}
