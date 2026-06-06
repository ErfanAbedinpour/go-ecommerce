package product

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for product listing.
type ListFilter struct {
	Status     string
	CategoryID *uuid.UUID
	Brand      string
	IsFeatured *bool
}

// Repository defines the port for product persistence.
type Repository interface {
	Create(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Product, error)
	FindBySlug(ctx context.Context, slug string) (*Product, error)
	FindBySKU(ctx context.Context, sku string) (*Product, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Product, int64, error)
	Search(ctx context.Context, query string, page pagination.Params) ([]Product, int64, error)
	UpdateInventory(ctx context.Context, productID uuid.UUID, inventory Inventory) error
	ExistsInActiveOrders(ctx context.Context, productID uuid.UUID) (bool, error)
	CategoryExists(ctx context.Context, categoryID uuid.UUID) (bool, error)
}
