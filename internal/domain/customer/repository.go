package customer

import (
	"context"

	"github.com/google/uuid"

	domainorder "app/internal/domain/order"
	"app/pkg/pagination"
)

// ListFilter holds optional filters for customer listing.
type ListFilter struct {
	Query string
	Type  *CustomerType
}

// Repository defines the port for customer persistence.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Customer, int64, error)
	ListAddresses(ctx context.Context, customerID uuid.UUID) ([]Address, error)
	ListOrders(ctx context.Context, customerID uuid.UUID, page pagination.Params) ([]domainorder.Summary, int64, error)
}
