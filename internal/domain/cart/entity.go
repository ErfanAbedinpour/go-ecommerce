package cart

import (
	"time"

	"github.com/google/uuid"
)

// Owner identifies a guest or authenticated cart.
type Owner struct {
	GuestToken string
	UserID     *uuid.UUID
}

// Item is a single cart line stored server-side.
type Item struct {
	ProductID uuid.UUID
	SkuID     *uuid.UUID
	Quantity  int
	AddedAt   time.Time
}

// Cart is the persisted cart state.
type Cart struct {
	Items     []Item
	UpdatedAt time.Time
}
