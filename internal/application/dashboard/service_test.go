package dashboard

import (
	"context"
	"testing"
	"time"

	domain "app/internal/domain/dashboard"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

type mockRepo struct {
	stats    *domain.Stats
	revenue  []domain.RevenueDataPoint
	lowStock []domainproduct.Product
	orders   []domainorder.DashboardSummary
	featured []domainproduct.Product
}

func (m *mockRepo) GetStats(_ context.Context) (*domain.Stats, error) {
	return m.stats, nil
}

func (m *mockRepo) GetRevenueAnalytics(_ context.Context, _ domain.DateRange) ([]domain.RevenueDataPoint, error) {
	return m.revenue, nil
}

func (m *mockRepo) ListLowStockProducts(_ context.Context, _ pagination.Params) ([]domainproduct.Product, int64, error) {
	return m.lowStock, int64(len(m.lowStock)), nil
}

func (m *mockRepo) ListRecentOrders(_ context.Context, limit int) ([]domainorder.DashboardSummary, error) {
	if limit > len(m.orders) {
		limit = len(m.orders)
	}
	return m.orders[:limit], nil
}

func (m *mockRepo) ListFeaturedProducts(_ context.Context, limit int) ([]domainproduct.Product, error) {
	if limit > len(m.featured) {
		limit = len(m.featured)
	}
	return m.featured[:limit], nil
}

func TestService_GetStats(t *testing.T) {
	svc := NewService(&mockRepo{
		stats: &domain.Stats{
			TotalRevenue: 1000, TotalOrders: 10, TotalCustomers: 5,
			TotalProducts: 20, PendingOrders: 2, LowStockCount: 3,
		},
	})

	stats, err := svc.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if stats.TotalRevenue != 1000 {
		t.Errorf("total revenue = %v, want 1000", stats.TotalRevenue)
	}
}

func TestService_GetRevenueAnalytics(t *testing.T) {
	svc := NewService(&mockRepo{
		revenue: []domain.RevenueDataPoint{
			{Date: "2026-06-01", Revenue: 150, Orders: 2},
		},
	})

	points, err := svc.GetRevenueAnalytics(context.Background(), domain.RevenueFilter{Period: domain.PeriodMonth})
	if err != nil {
		t.Fatalf("GetRevenueAnalytics() error = %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("points = %d, want 1", len(points))
	}
}

func TestService_GetRecentOrders_DefaultLimit(t *testing.T) {
	svc := NewService(&mockRepo{
		orders: []domainorder.DashboardSummary{
			{Summary: domainorder.Summary{OrderNumber: "ORD-1", CreatedAt: time.Now()}},
			{Summary: domainorder.Summary{OrderNumber: "ORD-2", CreatedAt: time.Now()}},
		},
	})

	orders, err := svc.GetRecentOrders(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetRecentOrders() error = %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("orders = %d, want 2", len(orders))
	}
}

func TestService_GetRecentOrders_InvalidLimit(t *testing.T) {
	svc := NewService(&mockRepo{})
	_, err := svc.GetRecentOrders(context.Background(), 100)
	if err != domain.ErrInvalidLimit {
		t.Errorf("expected invalid limit, got %v", err)
	}
}

func TestService_GetLowStockProducts(t *testing.T) {
	svc := NewService(&mockRepo{
		lowStock: []domainproduct.Product{{Name: "Low Item"}},
	})

	result, err := svc.GetLowStockProducts(context.Background(), pagination.Params{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("GetLowStockProducts() error = %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("products = %d, want 1", len(result.Data))
	}
}
