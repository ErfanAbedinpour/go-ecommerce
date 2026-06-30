package customer

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	"app/pkg/pagination"
)

type mockRepo struct {
	customers map[uuid.UUID]*domain.Customer
	addresses map[uuid.UUID][]domain.Address
	orders    map[uuid.UUID][]domainorder.Summary
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		customers: make(map[uuid.UUID]*domain.Customer),
		addresses: make(map[uuid.UUID][]domain.Address),
		orders:    make(map[uuid.UUID][]domainorder.Summary),
	}
}

func (m *mockRepo) Create(_ context.Context, c *domain.Customer) error {
	cp := *c
	m.customers[c.ID] = &cp
	return nil
}

func (m *mockRepo) FindByUserID(_ context.Context, userID uuid.UUID) (*domain.Customer, error) {
	for _, c := range m.customers {
		if c.UserID != nil && *c.UserID == userID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Customer, error) {
	c, ok := m.customers[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter, page pagination.Params) ([]domain.Customer, int64, error) {
	items := make([]domain.Customer, 0, len(m.customers))
	for _, c := range m.customers {
		items = append(items, *c)
	}
	return items, int64(len(items)), nil
}

func (m *mockRepo) ListAddresses(_ context.Context, customerID uuid.UUID) ([]domain.Address, error) {
	return m.addresses[customerID], nil
}
func (m *mockRepo) ReplaceAddresses(_ context.Context, customerID uuid.UUID, addresses []domain.Address) error {
	m.addresses[customerID] = addresses
	return nil
}

func (m *mockRepo) ListOrders(_ context.Context, customerID uuid.UUID, page pagination.Params) ([]domainorder.Summary, int64, error) {
	orders := m.orders[customerID]
	return orders, int64(len(orders)), nil
}
func (m *mockRepo) Count(context.Context) (int64, error) {
	return int64(len(m.customers)), nil
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*domain.Customer, error) {
	for _, c := range m.customers {
		if c.Email == email {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockRepo) Update(_ context.Context, c *domain.Customer) error {
	m.customers[c.ID] = c
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.customers, id)
	return nil
}

func (m *mockRepo) HasOrders(_ context.Context, customerID uuid.UUID) (bool, error) {
	return len(m.orders[customerID]) > 0, nil
}

func (m *mockRepo) GetLastOrderAt(_ context.Context, customerID uuid.UUID) (*time.Time, error) {
	orders := m.orders[customerID]
	if len(orders) == 0 {
		return nil, nil
	}
	latest := orders[0].CreatedAt
	for _, o := range orders[1:] {
		if o.CreatedAt.After(latest) {
			latest = o.CreatedAt
		}
	}
	return &latest, nil
}

func TestService_GetByID_WithAddresses(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	id := uuid.New()
	repo.customers[id] = &domain.Customer{
		ID: id, Email: "john@example.com", FirstName: "John", LastName: "Doe",
		Type: domain.TypeRegistered,
	}
	repo.addresses[id] = []domain.Address{
		{ID: uuid.New(), CustomerID: id, Type: domain.AddressHome, Street: "123 Main", City: "NYC", PostalCode: "10001", Country: "US"},
	}

	customer, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if len(customer.Addresses) != 1 {
		t.Fatalf("addresses = %d, want 1", len(customer.Addresses))
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetByID(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestService_GetPurchaseHistory(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	id := uuid.New()
	repo.customers[id] = &domain.Customer{ID: id, Email: "jane@example.com", FirstName: "Jane", LastName: "Doe", Type: domain.TypeGuest}
	repo.orders[id] = []domainorder.Summary{
		{ID: uuid.New(), OrderNumber: "ORD-001", Status: "delivered", PaymentStatus: "paid", Total: 99.99, ItemCount: 2, CreatedAt: time.Now()},
	}

	result, err := svc.GetPurchaseHistory(context.Background(), id, pagination.Params{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("GetPurchaseHistory() error = %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("orders = %d, want 1", len(result.Data))
	}
	if result.Data[0].OrderNumber != "ORD-001" {
		t.Errorf("order number = %q, want ORD-001", result.Data[0].OrderNumber)
	}
}

func TestService_GetPurchaseHistory_CustomerNotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetPurchaseHistory(context.Background(), uuid.New(), pagination.Params{Page: 1, PerPage: 20})
	if err != domain.ErrNotFound {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestService_Update(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()
	repo.customers[id] = &domain.Customer{
		ID: id, Email: "john@example.com", FirstName: "John", LastName: "Doe", Type: domain.TypeRegistered,
	}

	firstName := "Jonathan"
	customer, err := svc.Update(context.Background(), id, UpdateInput{FirstName: &firstName})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if customer.FirstName != "Jonathan" {
		t.Errorf("first_name = %q, want Jonathan", customer.FirstName)
	}
}

func TestService_Delete_HasOrders(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()
	repo.customers[id] = &domain.Customer{ID: id, Email: "j@example.com", Type: domain.TypeGuest}
	repo.orders[id] = []domainorder.Summary{{ID: uuid.New()}}

	err := svc.Delete(context.Background(), id)
	if err != domain.ErrHasOrders {
		t.Errorf("expected has orders error, got %v", err)
	}
}

func TestService_GetByID_LastOrderAt(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	id := uuid.New()
	created := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	repo.customers[id] = &domain.Customer{ID: id, Email: "j@example.com", Type: domain.TypeRegistered}
	repo.orders[id] = []domainorder.Summary{{CreatedAt: created}}

	customer, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if customer.LastOrderAt == nil || !customer.LastOrderAt.Equal(created) {
		t.Errorf("last_order_at = %v, want %v", customer.LastOrderAt, created)
	}
}

func TestService_List(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	id := uuid.New()
	repo.customers[id] = &domain.Customer{
		ID: id, Email: "alice@example.com", FirstName: "Alice", LastName: "Smith", Type: domain.TypeRegistered,
	}

	result, err := svc.List(context.Background(), domain.ListFilter{}, pagination.Params{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("customers = %d, want 1", len(result.Data))
	}
}
