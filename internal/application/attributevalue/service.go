package attributevalue

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/attributevalue"
	"app/pkg/pagination"
)

// Service handles global product attribute value use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new attribute value Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput holds data for creating an attribute value.
type CreateInput struct {
	AttributeID uuid.UUID
	Value       string
	SortOrder   int
	IsActive    bool
}

// UpdateInput holds partial update data.
type UpdateInput struct {
	Value     *string
	SortOrder *int
	IsActive  *bool
}

// Create creates a new attribute value.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Value, error) {
	exists, err := s.repo.AttributeExists(ctx, input.AttributeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrAttributeMissing
	}

	now := time.Now().UTC()
	v := &domain.Value{
		ID:          uuid.New(),
		AttributeID: input.AttributeID,
		Value:       strings.TrimSpace(input.Value),
		SortOrder:   input.SortOrder,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

// Update updates an existing attribute value.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Value, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Value != nil {
		v.Value = strings.TrimSpace(*input.Value)
	}
	if input.SortOrder != nil {
		v.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		v.IsActive = *input.IsActive
	}

	v.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

// Delete soft-deletes an attribute value.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

// GetByID returns an attribute value by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Value, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns a paginated attribute value list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.Value], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Value]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}
