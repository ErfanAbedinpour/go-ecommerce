package category

import (
	"time"

	"github.com/google/uuid"
)

// Category is the aggregate root for hierarchical product categories.
type Category struct {
	ID          uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Description string
	ImageURL    string
	SortOrder   int
	IsActive    bool
	Children    []Category
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
