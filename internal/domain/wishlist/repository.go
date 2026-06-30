package wishlist

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// Repository defines the port for wishlist persistence.
type Repository interface {
	Add(ctx context.Context, item *Item) error
	Remove(ctx context.Context, customerID, productID uuid.UUID) error
	List(ctx context.Context, customerID uuid.UUID, page pagination.Params) ([]ListItem, int64, error)
	Exists(ctx context.Context, customerID, productID uuid.UUID) (bool, error)
	BatchCheck(ctx context.Context, customerID uuid.UUID, productIDs []uuid.UUID) ([]uuid.UUID, error)
	ListProductIDs(ctx context.Context, customerID uuid.UUID) ([]uuid.UUID, error)
	Count(ctx context.Context, customerID uuid.UUID) (int64, error)
}
