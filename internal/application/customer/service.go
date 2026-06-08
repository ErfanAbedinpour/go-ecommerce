package customer

import (
	"context"
	"strings"
	"time"

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

// GetByID returns a customer with addresses and last order date.
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

	lastOrderAt, err := s.repo.GetLastOrderAt(ctx, id)
	if err != nil {
		return nil, err
	}
	customer.LastOrderAt = lastOrderAt

	return customer, nil
}

// UpdateInput holds partial update data for a customer.
type UpdateInput struct {
	Email     *string
	FirstName *string
	LastName  *string
	Phone     *string
	Type      *string
}

// Update updates an existing customer profile.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Customer, error) {
	customer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		existing, err := s.repo.FindByEmail(ctx, email)
		if err == nil && existing.ID != id {
			return nil, domain.ErrEmailTaken
		}
		if err != nil && err != domain.ErrNotFound {
			return nil, err
		}
		customer.Email = email
	}
	if input.FirstName != nil {
		customer.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		customer.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.Phone != nil {
		customer.Phone = strings.TrimSpace(*input.Phone)
	}
	if input.Type != nil {
		customerType, err := domain.ParseCustomerType(*input.Type)
		if err != nil {
			return nil, err
		}
		customer.Type = customerType
	}

	customer.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

// Delete removes a customer without order history.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	hasOrders, err := s.repo.HasOrders(ctx, id)
	if err != nil {
		return err
	}
	if hasOrders {
		return domain.ErrHasOrders
	}
	return s.repo.Delete(ctx, id)
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
