package contact

import (
	"context"
	"time"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

// ListFilter holds optional filters for contact message listing.
type ListFilter struct {
	Status string
	Source string
	Query  string
	From   *time.Time
	To     *time.Time
}

// Repository defines the port for contact message persistence.
type Repository interface {
	Create(ctx context.Context, m *Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	List(ctx context.Context, filter ListFilter, page pagination.Params) ([]Message, int64, error)
	CountStats(ctx context.Context) (*InboxStats, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	Delete(ctx context.Context, id uuid.UUID) error
}
