package customer

import (
	"context"

	"github.com/google/uuid"

	domain "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	"app/pkg/pagination"
)

// Service handles customer management use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new customer Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// List returns a paginated customer list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.Customer], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Customer]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetByID returns a customer with addresses.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	customer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	addresses, err := s.repo.ListAddresses(ctx, id)
	if err != nil {
		return nil, err
	}
	customer.Addresses = addresses
	return customer, nil
}

// GetPurchaseHistory returns paginated orders for a customer.
func (s *Service) GetPurchaseHistory(ctx context.Context, customerID uuid.UUID, page pagination.Params) (pagination.Paginated[domainorder.Summary], error) {
	if _, err := s.repo.FindByID(ctx, customerID); err != nil {
		return pagination.Paginated[domainorder.Summary]{}, err
	}

	items, total, err := s.repo.ListOrders(ctx, customerID, page)
	if err != nil {
		return pagination.Paginated[domainorder.Summary]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}
