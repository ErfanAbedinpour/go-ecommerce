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
