package productquestion

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for admin question listing.
type ListFilter struct {
	Status    string
	ProductID *uuid.UUID
	Query     string
}

// Repository defines the port for product question persistence.
type Repository interface {
	Create(ctx context.Context, q *Question) error
	FindByID(ctx context.Context, id uuid.UUID) (*Question, error)
	ListByProduct(ctx context.Context, productID uuid.UUID, page pagination.Params) ([]Question, int64, error)
	ListAdmin(ctx context.Context, filter ListFilter, page pagination.Params) ([]AdminListItem, int64, error)
	Answer(ctx context.Context, id uuid.UUID, answer string, answeredBy uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
