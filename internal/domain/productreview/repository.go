package productreview

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for admin review listing.
type ListFilter struct {
	Status    string
	ProductID *uuid.UUID
	Rating    *int
	Query     string
}

// Repository defines the port for product review persistence.
type Repository interface {
	Create(ctx context.Context, r *Review) error
	FindByID(ctx context.Context, id uuid.UUID) (*Review, error)
	ListByProduct(ctx context.Context, productID uuid.UUID, sort string, page pagination.Params) ([]Review, int64, error)
	ListAdmin(ctx context.Context, filter ListFilter, page pagination.Params) ([]AdminListItem, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetSummary(ctx context.Context, productID uuid.UUID) (*Summary, error)
	ExistsByCustomer(ctx context.Context, productID, customerID uuid.UUID) (bool, error)
}
