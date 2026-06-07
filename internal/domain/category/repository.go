package category

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for category listing.
type ListFilter struct {
	ParentID *uuid.UUID
	IsActive *bool
	Tree     bool
}

// Repository defines the port for category persistence.
type Repository interface {
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Category, int64, error)
	ListAll(ctx context.Context, filter ListFilter) ([]Category, error)
	HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
	HasProducts(ctx context.Context, id uuid.UUID) (bool, error)
	IsDescendant(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error)
}
