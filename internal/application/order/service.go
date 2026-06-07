package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/order"
	"app/pkg/pagination"
)

// Service handles order management use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new order Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// List returns a paginated order list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.ListItem], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.ListItem]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetByID returns full order detail.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateStatusInput holds data for a manual status update.
type UpdateStatusInput struct {
	Status    string
	Note      string
	ChangedBy uuid.UUID
}

// UpdateStatus transitions an order through the fulfillment workflow.
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateStatusInput) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newStatus, err := domain.ParseStatus(input.Status)
	if err != nil {
		return nil, err
	}

	fromStatus := order.Status
	if err := order.TransitionTo(newStatus); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	order.UpdatedAt = now

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	if err := s.recordHistory(ctx, order.ID, &fromStatus, newStatus, input.Note, input.ChangedBy, now); err != nil {
		return nil, err
	}

	if newStatus == domain.StatusCancelled {
		if err := s.repo.RestoreInventory(ctx, order.Items); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(ctx, id)
}

// Cancel cancels a pending or processing order and restores inventory.
func (s *Service) Cancel(ctx context.Context, id uuid.UUID, changedBy uuid.UUID) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !domain.CanCancel(order.Status) {
		return nil, domain.ErrCannotCancel
	}

	fromStatus := order.Status
	order.Status = domain.StatusCancelled
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	if err := s.recordHistory(ctx, order.ID, &fromStatus, domain.StatusCancelled, "Order cancelled", changedBy, order.UpdatedAt); err != nil {
		return nil, err
	}

	if err := s.repo.RestoreInventory(ctx, order.Items); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

// RefundInput holds data for refunding an order.
type RefundInput struct {
	Amount    float64
	Reason    string
	ChangedBy uuid.UUID
}

// Refund processes a full or partial refund on a delivered, paid order.
func (s *Service) Refund(ctx context.Context, id uuid.UUID, input RefundInput) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !domain.CanRefund(order.Status, order.PaymentStatus) {
		return nil, domain.ErrCannotRefund
	}

	if input.Amount <= 0 || input.Amount > order.Total {
		return nil, domain.ErrInvalidRefundAmount
	}

	fromStatus := order.Status
	order.Status = domain.StatusRefunded
	order.PaymentStatus = domain.PaymentRefunded
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	note := fmt.Sprintf("Refund of %.2f: %s", input.Amount, input.Reason)
	if err := s.recordHistory(ctx, order.ID, &fromStatus, domain.StatusRefunded, note, input.ChangedBy, order.UpdatedAt); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *Service) recordHistory(ctx context.Context, orderID uuid.UUID, from *domain.Status, to domain.Status, note string, changedBy uuid.UUID, at time.Time) error {
	entry := &domain.StatusHistory{
		ID:         uuid.New(),
		OrderID:    orderID,
		FromStatus: from,
		ToStatus:   to,
		Note:       note,
		ChangedBy:  &changedBy,
		CreatedAt:  at,
	}
	return s.repo.AddStatusHistory(ctx, entry)
}
