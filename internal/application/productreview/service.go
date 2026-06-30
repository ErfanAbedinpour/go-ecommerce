package productreview

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaincustomer "app/internal/domain/customer"
	domainproduct "app/internal/domain/product"
	domain "app/internal/domain/productreview"
	"app/internal/pkg/productref"
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

type SubmitInput struct {
	ProductID  uuid.UUID
	UserID     *uuid.UUID
	AuthorName string
	Rating     int
	Title      string
	Content    string
}

func (s *Service) Submit(ctx context.Context, input SubmitInput) (*domain.Review, error) {
	if _, err := s.products.FindByID(ctx, input.ProductID); err != nil {
		return nil, err
	}

	var customerID *uuid.UUID
	if input.UserID != nil {
		customer, err := s.customers.FindByUserID(ctx, *input.UserID)
		if err != nil {
			return nil, err
		}
		customerID = &customer.ID

		if exists, err := s.repo.ExistsByCustomer(ctx, input.ProductID, *customerID); err != nil {
			return nil, err
		} else if exists {
			return nil, domain.ErrAlreadyReviewed
		}
	}

	review := &domain.Review{
		ID:         uuid.New(),
		ProductID:  input.ProductID,
		CustomerID: customerID,
		AuthorName: input.AuthorName,
		Rating:     input.Rating,
		Title:      input.Title,
		Content:    input.Content,
		Status:     domain.StatusPending,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *Service) FindByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByProduct(ctx context.Context, productID uuid.UUID, sort string, page pagination.Params) (pagination.Paginated[domain.Review], error) {
	items, total, err := s.repo.ListByProduct(ctx, productID, sort, page)
	if err != nil {
		return pagination.Paginated[domain.Review]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) ListAdmin(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.AdminListItem], error) {
	items, total, err := s.repo.ListAdmin(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.AdminListItem]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetSummary(ctx context.Context, productID uuid.UUID) (*domain.Summary, error) {
	return s.repo.GetSummary(ctx, productID)
}

// ResolveProductID resolves a product slug or UUID to a product ID.
func (s *Service) ResolveProductID(ctx context.Context, ref string) (uuid.UUID, error) {
	return productref.ResolveID(ctx, s.products, ref)
}
