package wishlist

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaincustomer "app/internal/domain/customer"
	domainproduct "app/internal/domain/product"
	domain "app/internal/domain/wishlist"
	"app/pkg/pagination"
)

type Service struct {
	repo      domain.Repository
	products  domainproduct.Repository
	customers domaincustomer.Repository
}

func NewService(repo domain.Repository, products domainproduct.Repository, customers domaincustomer.Repository) *Service {
	return &Service{
		repo:      repo,
		products:  products,
		customers: customers,
	}
}

func (s *Service) Add(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*domain.Item, error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if _, err := s.products.FindByID(ctx, productID); err != nil {
		return nil, err
	}

	item := &domain.Item{
		ID:         uuid.New(),
		CustomerID: customer.ID,
		ProductID:  productID,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Add(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) Remove(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.Remove(ctx, customer.ID, productID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, page pagination.Params) (pagination.Paginated[domain.ListItem], error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return pagination.Paginated[domain.ListItem]{}, err
	}

	items, total, err := s.repo.List(ctx, customer.ID, page)
	if err != nil {
		return pagination.Paginated[domain.ListItem]{}, err
	}

	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) Check(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (bool, error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	return s.repo.Exists(ctx, customer.ID, productID)
}

func (s *Service) BatchCheck(ctx context.Context, userID uuid.UUID, productIDs []uuid.UUID) ([]uuid.UUID, error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.BatchCheck(ctx, customer.ID, productIDs)
}
