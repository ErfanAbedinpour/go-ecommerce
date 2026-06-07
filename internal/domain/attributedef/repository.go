package attributedef

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for attribute definition listing.
type ListFilter struct {
	IsActive *bool
}

// Repository defines the port for attribute definition persistence.
type Repository interface {
	Create(ctx context.Context, d *Definition) error
	Update(ctx context.Context, d *Definition) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Definition, error)
	FindBySlug(ctx context.Context, slug string) (*Definition, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Definition, int64, error)
	HasValues(ctx context.Context, id uuid.UUID) (bool, error)
}
