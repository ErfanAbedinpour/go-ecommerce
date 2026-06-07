package product

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "app/internal/domain/product"
	"app/pkg/pagination"
)

type mockRepo struct {
	products    map[uuid.UUID]*domain.Product
	slugs       map[string]uuid.UUID
	skus        map[string]uuid.UUID
	activeOrder bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		products: make(map[uuid.UUID]*domain.Product),
		slugs:    make(map[string]uuid.UUID),
		skus:     make(map[string]uuid.UUID),
	}
}

func (m *mockRepo) Create(_ context.Context, p *domain.Product) error {
	cp := *p
	m.products[p.ID] = &cp
	m.slugs[p.Slug] = p.ID
	m.skus[p.SKU] = p.ID
	return nil
}

func (m *mockRepo) Update(_ context.Context, p *domain.Product) error {
	m.products[p.ID] = p
	m.slugs[p.Slug] = p.ID
	m.skus[p.SKU] = p.ID
	return nil
}

func (m *mockRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	delete(m.products, id)
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockRepo) FindBySlug(_ context.Context, slug string) (*domain.Product, error) {
	id, ok := m.slugs[slug]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m.FindByID(context.Background(), id)
}

func (m *mockRepo) FindBySKU(_ context.Context, sku string) (*domain.Product, error) {
	id, ok := m.skus[sku]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m.FindByID(context.Background(), id)
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter, page pagination.Params) ([]domain.Product, int64, error) {
	items := make([]domain.Product, 0, len(m.products))
	for _, p := range m.products {
		items = append(items, *p)
	}
	return items, int64(len(items)), nil
}

func (m *mockRepo) Search(_ context.Context, _ string, page pagination.Params) ([]domain.Product, int64, error) {
	return m.List(context.Background(), domain.ListFilter{}, page)
}

func (m *mockRepo) UpdateInventory(_ context.Context, productID uuid.UUID, inv domain.Inventory) error {
	p, ok := m.products[productID]
	if !ok {
		return domain.ErrNotFound
	}
	p.Inventory = inv
	return nil
}

func (m *mockRepo) ExistsInActiveOrders(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.activeOrder, nil
}

func (m *mockRepo) CategoryExists(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}

func (m *mockRepo) GetStats(_ context.Context) (*domain.Stats, error) {
	return &domain.Stats{Total: 10, Active: 7, Draft: 2, OutOfStock: 1}, nil
}

func TestService_Create(t *testing.T) {
	svc := NewService(newMockRepo())
	p, err := svc.Create(context.Background(), CreateInput{
		Name:  "Nike Air Max",
		SKU:   "PROD-001",
		Price: 129.99,
		Inventory: InventoryInput{Quantity: 50, LowStockThreshold: 10},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if p.Slug != "nike-air-max" {
		t.Errorf("Slug = %q, want %q", p.Slug, "nike-air-max")
	}
	if p.Status != domain.StatusDraft {
		t.Errorf("Status = %q, want draft", p.Status)
	}
}

func TestService_Create_SlugConflict(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{Name: "Test", SKU: "SKU-1", Price: 10})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = svc.Create(context.Background(), CreateInput{Name: "Other", Slug: "test", SKU: "SKU-2", Price: 10})
	if err != domain.ErrSlugConflict {
		t.Errorf("expected slug conflict, got %v", err)
	}
}

func TestService_Delete_ActiveOrders(t *testing.T) {
	repo := newMockRepo()
	repo.activeOrder = true
	svc := NewService(repo)

	p, _ := svc.Create(context.Background(), CreateInput{Name: "Test", SKU: "SKU-1", Price: 10})
	err := svc.Delete(context.Background(), p.ID)
	if err != domain.ErrHasActiveOrders {
		t.Errorf("expected ErrHasActiveOrders, got %v", err)
	}
}

func TestService_UpdateInventory(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	p, _ := svc.Create(context.Background(), CreateInput{
		Name: "Test", SKU: "SKU-1", Price: 10,
		Inventory: InventoryInput{Quantity: 5},
	})

	inv, err := svc.UpdateInventory(context.Background(), p.ID, InventoryUpdateInput{Quantity: 100})
	if err != nil {
		t.Fatalf("UpdateInventory() error = %v", err)
	}
	if inv.Quantity != 100 {
		t.Errorf("Quantity = %d, want 100", inv.Quantity)
	}
}

func TestService_InvalidSalePrice(t *testing.T) {
	svc := NewService(newMockRepo())
	sale := 150.0
	_, err := svc.Create(context.Background(), CreateInput{
		Name: "Test", SKU: "SKU-1", Price: 100, SalePrice: &sale,
	})
	if err != domain.ErrInvalidSalePrice {
		t.Errorf("expected invalid sale price, got %v", err)
	}
}
