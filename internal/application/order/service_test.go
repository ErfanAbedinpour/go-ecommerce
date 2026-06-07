package order

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "app/internal/domain/order"
	"app/pkg/pagination"
)

type mockRepo struct {
	orders    map[uuid.UUID]*domain.Order
	history   []domain.StatusHistory
	restored  bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{orders: make(map[uuid.UUID]*domain.Order)}
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *o
	return &cp, nil
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter, page pagination.Params) ([]domain.ListItem, int64, error) {
	items := make([]domain.ListItem, 0, len(m.orders))
	for _, o := range m.orders {
		items = append(items, domain.ListItem{
			Summary: domain.Summary{
				ID: o.ID, OrderNumber: o.OrderNumber, Status: o.Status.String(),
				PaymentStatus: o.PaymentStatus.String(), Total: o.Total,
			},
			CustomerID: o.CustomerID,
		})
	}
	return items, int64(len(items)), nil
}

func (m *mockRepo) Update(_ context.Context, o *domain.Order) error {
	m.orders[o.ID] = o
	return nil
}

func (m *mockRepo) AddStatusHistory(_ context.Context, entry *domain.StatusHistory) error {
	m.history = append(m.history, *entry)
	return nil
}

func (m *mockRepo) RestoreInventory(_ context.Context, _ []domain.Item) error {
	m.restored = true
	return nil
}

func TestService_UpdateStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()
	adminID := uuid.New()

	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusPending, PaymentStatus: domain.PaymentUnpaid,
		Items: []domain.Item{{ProductID: uuid.New(), Quantity: 2}},
	}

	order, err := svc.UpdateStatus(context.Background(), id, UpdateStatusInput{
		Status: "processing", Note: "Started", ChangedBy: adminID,
	})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if order.Status != domain.StatusProcessing {
		t.Errorf("status = %q, want processing", order.Status)
	}
}

func TestService_UpdateStatus_InvalidTransition(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()

	repo.orders[id] = &domain.Order{ID: id, Status: domain.StatusShipped, PaymentStatus: domain.PaymentPaid}

	_, err := svc.UpdateStatus(context.Background(), id, UpdateStatusInput{
		Status: "cancelled", ChangedBy: uuid.New(),
	})
	if err != domain.ErrInvalidStatusTransition {
		t.Errorf("expected invalid transition, got %v", err)
	}
}

func TestService_Cancel_RestoresInventory(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()

	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusPending, PaymentStatus: domain.PaymentUnpaid,
		Items: []domain.Item{{ProductID: uuid.New(), Quantity: 1}},
	}

	_, err := svc.Cancel(context.Background(), id, uuid.New())
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if !repo.restored {
		t.Error("expected inventory restore")
	}
}

func TestService_Cancel_NotAllowed(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()
	repo.orders[id] = &domain.Order{ID: id, Status: domain.StatusShipped}

	_, err := svc.Cancel(context.Background(), id, uuid.New())
	if err != domain.ErrCannotCancel {
		t.Errorf("expected cannot cancel, got %v", err)
	}
}

func TestService_Refund(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()

	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusDelivered, PaymentStatus: domain.PaymentPaid, Total: 100,
	}

	order, err := svc.Refund(context.Background(), id, RefundInput{
		Amount: 50, Reason: "Partial refund", ChangedBy: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	if order.Status != domain.StatusRefunded {
		t.Errorf("status = %q, want refunded", order.Status)
	}
	if order.PaymentStatus != domain.PaymentRefunded {
		t.Errorf("payment = %q, want refunded", order.PaymentStatus)
	}
}

func TestService_Refund_InvalidAmount(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()
	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusDelivered, PaymentStatus: domain.PaymentPaid, Total: 100,
	}

	_, err := svc.Refund(context.Background(), id, RefundInput{Amount: 200, Reason: "Too much", ChangedBy: uuid.New()})
	if err != domain.ErrInvalidRefundAmount {
		t.Errorf("expected invalid amount, got %v", err)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetByID(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected not found, got %v", err)
	}
}
