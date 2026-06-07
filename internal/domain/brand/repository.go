package brand

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for brand listing.
type ListFilter struct {
	IsActive *bool
}

// Repository defines the port for brand persistence.
type Repository interface {
	Create(ctx context.Context, b *Brand) error
	Update(ctx context.Context, b *Brand) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Brand, error)
	FindBySlug(ctx context.Context, slug string) (*Brand, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Brand, int64, error)
	HasProducts(ctx context.Context, name string) (bool, error)
}
