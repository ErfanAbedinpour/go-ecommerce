package category

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/category"
	"app/pkg/pagination"
)

// Service handles category management use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new category Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput holds data for creating a category.
type CreateInput struct {
	Name        string
	Slug        string
	Description string
	ParentID    *uuid.UUID
	ImageURL    string
	SortOrder   int
	IsActive    bool
}

// UpdateInput holds partial update data for a category.
type UpdateInput struct {
	Name        *string
	Slug        *string
	Description *string
	ParentID    *uuid.UUID
	ImageURL    *string
	SortOrder   *int
	IsActive    *bool
}

// Create creates a new category.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Category, error) {
	if input.ParentID != nil {
		if _, err := s.repo.FindByID(ctx, *input.ParentID); err != nil {
			return nil, domain.ErrParentNotFound
		}
	}

	slug := input.Slug
	if slug == "" {
		slug = domain.GenerateSlug(input.Name)
	}
	if err := s.ensureUniqueSlug(ctx, slug, uuid.Nil); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	category := &domain.Category{
		ID:          uuid.New(),
		ParentID:    input.ParentID,
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		ImageURL:    input.ImageURL,
		SortOrder:   input.SortOrder,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

// Update updates an existing category.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Category, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Slug != nil {
		if err := s.ensureUniqueSlug(ctx, *input.Slug, id); err != nil {
			return nil, err
		}
		category.Slug = *input.Slug
	}
	if input.Description != nil {
		category.Description = *input.Description
	}
	if input.ParentID != nil {
		if *input.ParentID == id {
			return nil, domain.ErrInvalidParent
		}
		if _, err := s.repo.FindByID(ctx, *input.ParentID); err != nil {
			return nil, domain.ErrParentNotFound
		}
		isDesc, err := s.repo.IsDescendant(ctx, id, *input.ParentID)
		if err != nil {
			return nil, err
		}
		if isDesc {
			return nil, domain.ErrInvalidParent
		}
		category.ParentID = input.ParentID
	}
	if input.ImageURL != nil {
		category.ImageURL = *input.ImageURL
	}
	if input.SortOrder != nil {
		category.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		category.IsActive = *input.IsActive
	}

	category.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

// Delete soft-deletes a category.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return domain.ErrHasChildren
	}

	hasProducts, err := s.repo.HasProducts(ctx, id)
	if err != nil {
		return err
	}
	if hasProducts {
		return domain.ErrHasProducts
	}

	return s.repo.SoftDelete(ctx, id)
}

// GetByID returns a category by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns categories (flat paginated or nested tree).
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (any, error) {
	if filter.Tree {
		items, err := s.repo.ListAll(ctx, filter)
		if err != nil {
			return nil, err
		}
		return buildTree(items), nil
	}

	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return nil, err
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

func buildTree(items []domain.Category) []domain.Category {
	byID := make(map[uuid.UUID]*domain.Category, len(items))
	roots := make([]domain.Category, 0)

	for i := range items {
		item := items[i]
		item.Children = nil
		byID[item.ID] = &item
	}

	for _, item := range byID {
		if item.ParentID == nil {
			roots = append(roots, *item)
			continue
		}
		if parent, ok := byID[*item.ParentID]; ok {
			parent.Children = append(parent.Children, *item)
		} else {
			roots = append(roots, *item)
		}
	}

	// Rebuild roots from map to include nested children
	result := make([]domain.Category, 0, len(roots))
	for _, r := range roots {
		if built, ok := byID[r.ID]; ok {
			result = append(result, *built)
		}
	}
	return result
}
