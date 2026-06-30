package productquestion

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainproduct "app/internal/domain/product"
	domain "app/internal/domain/productquestion"
	"app/internal/pkg/productref"
	"app/pkg/pagination"
)

type Service struct {
	repo     domain.Repository
	products domainproduct.Repository
}

func NewService(repo domain.Repository, products domainproduct.Repository) *Service {
	return &Service{
		repo:     repo,
		products: products,
	}
}

type AskInput struct {
	ProductID  uuid.UUID
	AskerName  string
	AskerEmail string
	Question   string
}

func (s *Service) Ask(ctx context.Context, input AskInput) (*domain.Question, error) {
	if _, err := s.products.FindByID(ctx, input.ProductID); err != nil {
		return nil, err
	}

	question := &domain.Question{
		ID:         uuid.New(),
		ProductID:  input.ProductID,
		AskerName:  input.AskerName,
		AskerEmail: input.AskerEmail,
		Question:   input.Question,
		Status:     domain.StatusOpen,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, question); err != nil {
		return nil, err
	}

	return question, nil
}

func (s *Service) FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByProduct(ctx context.Context, productID uuid.UUID, page pagination.Params) (pagination.Paginated[domain.Question], error) {
	items, total, err := s.repo.ListByProduct(ctx, productID, page)
	if err != nil {
		return pagination.Paginated[domain.Question]{}, err
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

func (s *Service) Answer(ctx context.Context, id uuid.UUID, answer string, answeredBy uuid.UUID) error {
	return s.repo.Answer(ctx, id, answer, answeredBy)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ResolveProductID resolves a product slug or UUID to a product ID.
func (s *Service) ResolveProductID(ctx context.Context, ref string) (uuid.UUID, error) {
	return productref.ResolveID(ctx, s.products, ref)
}
