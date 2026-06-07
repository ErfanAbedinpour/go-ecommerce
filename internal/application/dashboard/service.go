package dashboard

import (
	"context"

	domain "app/internal/domain/dashboard"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

const (
	defaultRecentLimit = 10
	maxRecentLimit     = 50
)

// Service handles dashboard analytics use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new dashboard Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// GetStats returns aggregated dashboard KPIs.
func (s *Service) GetStats(ctx context.Context) (*domain.Stats, error) {
	return s.repo.GetStats(ctx)
}

// GetRevenueAnalytics returns time-series revenue data.
func (s *Service) GetRevenueAnalytics(ctx context.Context, filter domain.RevenueFilter) ([]domain.RevenueDataPoint, error) {
	dateRange, err := domain.ResolveDateRange(filter)
	if err != nil {
		return nil, err
	}
	return s.repo.GetRevenueAnalytics(ctx, dateRange)
}

// GetLowStockProducts returns paginated products at or below their low-stock threshold.
func (s *Service) GetLowStockProducts(ctx context.Context, page pagination.Params) (pagination.Paginated[domainproduct.Product], error) {
	items, total, err := s.repo.ListLowStockProducts(ctx, page)
	if err != nil {
		return pagination.Paginated[domainproduct.Product]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetRecentOrders returns the latest orders for the dashboard feed.
func (s *Service) GetRecentOrders(ctx context.Context, limit int) ([]domainorder.Summary, error) {
	if limit <= 0 {
		limit = defaultRecentLimit
	}
	if limit > maxRecentLimit {
		return nil, domain.ErrInvalidLimit
	}
	return s.repo.ListRecentOrders(ctx, limit)
}
