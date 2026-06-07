package order

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for order listing.
type ListFilter struct {
	Status        string
	PaymentStatus string
	Query         string
}

// Repository defines the port for order persistence.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]ListItem, int64, error)
	Update(ctx context.Context, order *Order) error
	AddStatusHistory(ctx context.Context, entry *StatusHistory) error
	RestoreInventory(ctx context.Context, items []Item) error
}
