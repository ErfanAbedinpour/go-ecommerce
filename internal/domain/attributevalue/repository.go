package attributevalue

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for attribute value listing.
type ListFilter struct {
	AttributeID *uuid.UUID
	IsActive    *bool
}

// Repository defines the port for attribute value persistence.
type Repository interface {
	Create(ctx context.Context, v *Value) error
	Update(ctx context.Context, v *Value) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Value, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Value, int64, error)
	AttributeExists(ctx context.Context, attributeID uuid.UUID) (bool, error)
}
