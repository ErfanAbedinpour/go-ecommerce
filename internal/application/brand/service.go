package brand

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/brand"
	"app/pkg/pagination"
)

// Service handles brand catalog use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new brand Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput holds data for creating a brand.
type CreateInput struct {
	Name        string
	Slug        string
	Description string
	IsActive    bool
}

// UpdateInput holds partial update data for a brand.
type UpdateInput struct {
	Name        *string
	Slug        *string
	Description *string
	IsActive    *bool
}

// Create creates a new brand.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Brand, error) {
	slug := input.Slug
	if slug == "" {
		slug = domain.GenerateSlug(input.Name)
	}
	if err := s.ensureUniqueSlug(ctx, slug, uuid.Nil); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	b := &domain.Brand{
		ID:          uuid.New(),
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// Update updates an existing brand.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Brand, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		b.Name = *input.Name
	}
	if input.Slug != nil {
		if err := s.ensureUniqueSlug(ctx, *input.Slug, id); err != nil {
			return nil, err
		}
		b.Slug = *input.Slug
	}
	if input.Description != nil {
		b.Description = *input.Description
	}
	if input.IsActive != nil {
		b.IsActive = *input.IsActive
	}

	b.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// Delete soft-deletes a brand.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	hasProducts, err := s.repo.HasProducts(ctx, b.Name)
	if err != nil {
		return err
	}
	if hasProducts {
		return domain.ErrHasProducts
	}

	return s.repo.SoftDelete(ctx, id)
}

// GetByID returns a brand by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Brand, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns a paginated brand list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.Brand], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Brand]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) ensureUniqueSlug(ctx context.Context, slug string, excludeID uuid.UUID) error {
	existing, err := s.repo.FindBySlug(ctx, slug)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return domain.ErrSlugConflict
	}
	return nil
}
