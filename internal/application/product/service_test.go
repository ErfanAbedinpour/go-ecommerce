package product

import (
	"context"
	"strconv"
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
	for _, sku := range p.SKUs {
		m.skus[sku.Code] = p.ID
	}
	return nil
}

func (m *mockRepo) Update(_ context.Context, p *domain.Product) error {
	m.products[p.ID] = p
	m.slugs[p.Slug] = p.ID
	for _, sku := range p.SKUs {
		m.skus[sku.Code] = p.ID
	}
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

func (m *mockRepo) ListStorefront(_ context.Context, _ domain.StoreListFilter, page pagination.Params) ([]domain.Product, int64, error) {
	return m.List(context.Background(), domain.ListFilter{Status: string(domain.StatusActive)}, page)
}
func (m *mockRepo) SearchStorefront(context.Context, string, int) ([]domain.Product, error) {
	return nil, nil
}
func (m *mockRepo) ListRelatedStorefront(context.Context, uuid.UUID, int) ([]domain.Product, error) {
	return nil, nil
}

func (m *mockRepo) CountActive(_ context.Context) (int64, error) {
	var count int64
	for _, p := range m.products {
		if p.Status == domain.StatusActive {
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) FindByIDs(_ context.Context, ids []uuid.UUID) ([]domain.Product, error) {
	out := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			out = append(out, *p)
		}
	}
	return out, nil
}

func TestService_Create(t *testing.T) {
	svc := NewService(newMockRepo())
	p, err := svc.Create(context.Background(), CreateInput{
		Name:  "Nike Air Max",
		Price: 129.99,
		Inventory: InventoryInput{Quantity: 50, LowStockThreshold: 10},
		Attributes: []AttributeInput{
			{Name: "Color", Values: []string{"Red", "Blue"}},
			{Name: "Size", Values: []string{"9", "10"}},
		},
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
	if len(p.SKUs) != 4 {
		t.Errorf("Expected 4 SKUs, got %d", len(p.SKUs))
	}
}

func TestService_Create_SlugConflict(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{Name: "Test", Price: 10})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = svc.Create(context.Background(), CreateInput{Name: "Other", Slug: "test", Price: 10})
	if err != domain.ErrSlugConflict {
		t.Errorf("expected slug conflict, got %v", err)
	}
}

func TestService_Delete_ActiveOrders(t *testing.T) {
	repo := newMockRepo()
	repo.activeOrder = true
	svc := NewService(repo)

	p, _ := svc.Create(context.Background(), CreateInput{Name: "Test", Price: 10})
	err := svc.Delete(context.Background(), p.ID)
	if err != domain.ErrHasActiveOrders {
		t.Errorf("expected ErrHasActiveOrders, got %v", err)
	}
}

func TestService_UpdateInventory(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	p, _ := svc.Create(context.Background(), CreateInput{
		Name: "Test", Price: 10,
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
		Name: "Test", Price: 100, SalePrice: &sale,
	})
	if err != domain.ErrInvalidSalePrice {
		t.Errorf("expected invalid sale price, got %v", err)
	}
}

func TestGenerateSKUs(t *testing.T) {
	productID := uuid.New()
	inputs := []AttributeInput{
		{Name: "Color", Values: []string{"Red", "Blue"}},
		{Name: "Size", Values: []string{"S", "M", "L"}},
	}

	attributes, skus, err := generateSKUs(productID, "test-slug", inputs)
	if err != nil {
		t.Fatalf("generateSKUs() error = %v", err)
	}

	if len(attributes) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(attributes))
	}
	if len(skus) != 6 {
		t.Errorf("Expected 6 SKUs, got %d", len(skus))
	}

	// Check if all combinations are generated
	expectedCodes := map[string]bool{
		"TEST-SLUG-RED-S": false, "TEST-SLUG-RED-M": false, "TEST-SLUG-RED-L": false,
		"TEST-SLUG-BLUE-S": false, "TEST-SLUG-BLUE-M": false, "TEST-SLUG-BLUE-L": false,
	}

	for _, sku := range skus {
		if _, ok := expectedCodes[sku.Code]; !ok {
			t.Errorf("Unexpected SKU code: %s", sku.Code)
		}
		expectedCodes[sku.Code] = true
	}

	for code, found := range expectedCodes {
		if !found {
			t.Errorf("Missing SKU code: %s", code)
		}
	}
}

func TestGenerateSKUs_Validation(t *testing.T) {
	productID := uuid.New()

	tests := []struct {
		name        string
		inputs      []AttributeInput
		expectedErr error
	}{
		{
			name: "Empty attribute name",
			inputs: []AttributeInput{
				{Name: "", Values: []string{"Red"}},
			},
			expectedErr: domain.ErrEmptyAttributeName,
		},
		{
			name: "Duplicate attribute name",
			inputs: []AttributeInput{
				{Name: "Color", Values: []string{"Red"}},
				{Name: "color", Values: []string{"Blue"}},
			},
			expectedErr: domain.ErrDuplicateAttributeName,
		},
		{
			name: "Empty attribute values",
			inputs: []AttributeInput{
				{Name: "Color", Values: []string{}},
			},
			expectedErr: domain.ErrEmptyAttributeValues,
		},
		{
			name: "Empty attribute value",
			inputs: []AttributeInput{
				{Name: "Color", Values: []string{"Red", ""}},
			},
			expectedErr: domain.ErrEmptyAttributeValue,
		},
		{
			name: "Duplicate attribute value",
			inputs: []AttributeInput{
				{Name: "Color", Values: []string{"Red", "red"}},
			},
			expectedErr: domain.ErrDuplicateAttributeValue,
		},
		{
			name: "Max variants exceeded",
			inputs: []AttributeInput{
				{Name: "A", Values: make([]string, 11)},
				{Name: "B", Values: make([]string, 100)},
			},
			expectedErr: domain.ErrMaxVariantsExceeded,
		},
	}

	// Fill the slices for the last test
	for i := 0; i < 11; i++ {
		tests[5].inputs[0].Values[i] = "A" + strconv.Itoa(i)
	}
	for i := 0; i < 100; i++ {
		tests[5].inputs[1].Values[i] = "B" + strconv.Itoa(i)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := generateSKUs(productID, "test-slug", tt.inputs)
			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
