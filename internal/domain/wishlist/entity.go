package wishlist

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a product saved to a customer's wishlist.
type Item struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	ProductID  uuid.UUID
	CreatedAt  time.Time
}

// ProductSummary is a lightweight product view for wishlist display.
type ProductSummary struct {
	Name      string
	Slug      string
	Price     float64
	SalePrice *float64
	ImageURL  string
	IsInStock bool
}

// ListItem combines a wishlist entry with product summary data.
type ListItem struct {
	Item
	Product ProductSummary
}
