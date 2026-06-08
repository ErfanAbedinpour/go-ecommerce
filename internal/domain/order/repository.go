package order

import (
	"context"
	"time"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for order listing.
type ListFilter struct {
	Status        string
	PaymentStatus string
	Query         string
	From          *time.Time
	To            *time.Time
}

// Repository defines the port for order persistence.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]ListItem, int64, error)
	Create(ctx context.Context, order *Order) error
	Update(ctx context.Context, order *Order) error
	UpdateNotes(ctx context.Context, id uuid.UUID, notes string, updatedAt time.Time) error
	AddStatusHistory(ctx context.Context, entry *StatusHistory) error
	RestoreInventory(ctx context.Context, items []Item) error
	NextOrderNumber(ctx context.Context) (string, error)
	IncrementCouponUsage(ctx context.Context, couponID uuid.UUID) error
}
