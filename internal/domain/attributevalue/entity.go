package attributevalue

import (
	"time"

	"github.com/google/uuid"
)

// Value is an allowed value for a global product attribute definition.
type Value struct {
	ID          uuid.UUID
	AttributeID uuid.UUID
	Value       string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
