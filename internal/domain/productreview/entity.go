package productreview

import (
	"time"

	"github.com/google/uuid"
)

// Status is the moderation state of a review.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

// Review represents a customer-submitted product review.
type Review struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	CustomerID  *uuid.UUID
	AuthorName  string
	Rating      int
	Title       string
	Content     string
	Status      Status
	CreatedAt   time.Time
}

// Summary holds aggregate review statistics for a product.
type Summary struct {
	AverageRating float64
	TotalCount    int64
	Distribution  map[int]int64 // rating → count
}

// AdminListItem includes product name for admin views.
type AdminListItem struct {
	Review
	ProductName string
}
