package attributedef

import (
	"time"

	"github.com/google/uuid"
)

// Definition is the aggregate root for global product attribute definitions.
type Definition struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
