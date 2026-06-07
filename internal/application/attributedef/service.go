package attributedef

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/attributedef"
	"app/pkg/pagination"
)

// Service handles global product attribute definition use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new attribute definition Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput holds data for creating an attribute definition.
type CreateInput struct {
	Name      string
	Slug      string
	SortOrder int
	IsActive  bool
}

// UpdateInput holds partial update data.
type UpdateInput struct {
	Name      *string
	Slug      *string
	SortOrder *int
	IsActive  *bool
}

// Create creates a new attribute definition.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Definition, error) {
	slug := input.Slug
	if slug == "" {
		slug = domain.GenerateSlug(input.Name)
	}
	if err := s.ensureUniqueSlug(ctx, slug, uuid.Nil); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	d := &domain.Definition{
		ID:        uuid.New(),
		Name:      input.Name,
		Slug:      slug,
		SortOrder: input.SortOrder,
		IsActive:  input.IsActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// Update updates an existing attribute definition.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Definition, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		d.Name = *input.Name
	}
	if input.Slug != nil {
		if err := s.ensureUniqueSlug(ctx, *input.Slug, id); err != nil {
			return nil, err
		}
		d.Slug = *input.Slug
	}
	if input.SortOrder != nil {
		d.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		d.IsActive = *input.IsActive
	}

	d.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// Delete soft-deletes an attribute definition.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	hasValues, err := s.repo.HasValues(ctx, id)
	if err != nil {
		return err
	}
	if hasValues {
		return domain.ErrHasValues
	}

	return s.repo.SoftDelete(ctx, id)
}

// GetByID returns an attribute definition by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Definition, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns a paginated attribute definition list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.Definition], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Definition]{}, err
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
