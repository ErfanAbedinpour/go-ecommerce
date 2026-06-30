package cart

import "context"

// Repository persists carts for guest tokens and authenticated users.
type Repository interface {
	Get(ctx context.Context, owner Owner) (*Cart, error)
	Save(ctx context.Context, owner Owner, cart *Cart) error
	Delete(ctx context.Context, owner Owner) error
}
