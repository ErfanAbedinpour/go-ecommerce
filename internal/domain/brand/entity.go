package brand

import (
	"time"

	"github.com/google/uuid"
)

// Brand is the aggregate root for product brands.
type Brand struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
