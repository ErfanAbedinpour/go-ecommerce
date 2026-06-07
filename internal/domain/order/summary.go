package order

import (
	"time"

	"github.com/google/uuid"
)

// Summary is a read-model projection of an order for list views.
type Summary struct {
	ID            uuid.UUID
	OrderNumber   string
	Status        string
	PaymentStatus string
	Total         float64
	ItemCount     int
	CreatedAt     time.Time
}

// ListItem extends Summary with customer info for admin order lists.
type ListItem struct {
	Summary
	CustomerID    uuid.UUID
	CustomerName  string
	CustomerEmail string
}
