package customer

import (
	"context"
	"time"

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
	Create(ctx context.Context, customer *Customer) error
	FindByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	FindByEmail(ctx context.Context, email string) (*Customer, error)
	FindGuestByEmail(ctx context.Context, email string) (*Customer, error)
	FindGuestByPhone(ctx context.Context, phone string) (*Customer, error)
	FindRegisteredByPhone(ctx context.Context, phone string) (*Customer, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*Customer, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Customer, int64, error)
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
	HasOrders(ctx context.Context, customerID uuid.UUID) (bool, error)
	GetLastOrderAt(ctx context.Context, customerID uuid.UUID) (*time.Time, error)
	ListAddresses(ctx context.Context, customerID uuid.UUID) ([]Address, error)
	ReplaceAddresses(ctx context.Context, customerID uuid.UUID, addresses []Address) error
	ListOrders(ctx context.Context, customerID uuid.UUID, page pagination.Params) ([]domainorder.Summary, int64, error)
	Count(ctx context.Context) (int64, error)
	RecordOrderPlaced(ctx context.Context, customerID uuid.UUID, orderTotal float64, orderedAt time.Time) error
	RecordOrderCancelled(ctx context.Context, customerID uuid.UUID, orderTotal float64) error
}
