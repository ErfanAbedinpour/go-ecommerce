package coupon

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for coupon listing.
type ListFilter struct {
	IsActive *bool
	Query    string
}

// Repository defines the port for coupon persistence.
type Repository interface {
	Create(ctx context.Context, coupon *Coupon) error
	Update(ctx context.Context, coupon *Coupon) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Coupon, error)
	FindByCode(ctx context.Context, code string) (*Coupon, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Coupon, int64, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
}
