package storefront

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domaincategory "app/internal/domain/category"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

type catalogProductRepo struct {
	byID   map[uuid.UUID]*domainproduct.Product
	bySlug map[string]*domainproduct.Product
	active []domainproduct.Product
}

func (m *catalogProductRepo) Create(context.Context, *domainproduct.Product) error { return nil }
func (m *catalogProductRepo) Update(context.Context, *domainproduct.Product) error { return nil }
func (m *catalogProductRepo) SoftDelete(context.Context, uuid.UUID) error          { return nil }
func (m *catalogProductRepo) FindBySKU(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *catalogProductRepo) List(context.Context, domainproduct.ListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *catalogProductRepo) Search(context.Context, string, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *catalogProductRepo) CountActive(context.Context) (int64, error) { return 0, nil }
func (m *catalogProductRepo) UpdateInventory(context.Context, uuid.UUID, domainproduct.Inventory) error {
	return nil
}
func (m *catalogProductRepo) ExistsInActiveOrders(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *catalogProductRepo) CategoryExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (m *catalogProductRepo) GetStats(context.Context) (*domainproduct.Stats, error) {
	return nil, nil
}
func (m *catalogProductRepo) FindByID(_ context.Context, id uuid.UUID) (*domainproduct.Product, error) {
	p, ok := m.byID[id]
	if !ok {
		return nil, domainproduct.ErrNotFound
	}
	return p, nil
}
func (m *catalogProductRepo) FindBySlug(_ context.Context, slug string) (*domainproduct.Product, error) {
	p, ok := m.bySlug[slug]
	if !ok {
		return nil, domainproduct.ErrNotFound
	}
	return p, nil
}
func (m *catalogProductRepo) ListStorefront(_ context.Context, _ domainproduct.StoreListFilter, page pagination.Params) ([]domainproduct.Product, int64, error) {
	return m.active, int64(len(m.active)), nil
}
func (m *catalogProductRepo) SearchStorefront(context.Context, string, int) ([]domainproduct.Product, error) {
	return nil, nil
}
func (m *catalogProductRepo) ListRelatedStorefront(context.Context, uuid.UUID, int) ([]domainproduct.Product, error) {
	return nil, nil
}
func (m *catalogProductRepo) FindByIDs(_ context.Context, ids []uuid.UUID) ([]domainproduct.Product, error) {
	out := make([]domainproduct.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := m.byID[id]; ok {
			out = append(out, *p)
		}
	}
	return out, nil
}

type catalogCategoryRepo struct{}

func (catalogCategoryRepo) Create(context.Context, *domaincategory.Category) error { return nil }
func (catalogCategoryRepo) Update(context.Context, *domaincategory.Category) error { return nil }
func (catalogCategoryRepo) SoftDelete(context.Context, uuid.UUID) error          { return nil }
func (catalogCategoryRepo) FindByID(context.Context, uuid.UUID) (*domaincategory.Category, error) {
	return nil, domaincategory.ErrNotFound
}
func (catalogCategoryRepo) FindBySlug(context.Context, string) (*domaincategory.Category, error) {
	return nil, domaincategory.ErrNotFound
}
func (catalogCategoryRepo) ListAll(context.Context, domaincategory.ListFilter) ([]domaincategory.Category, error) {
	return nil, nil
}
func (catalogCategoryRepo) IsDescendant(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}
func (catalogCategoryRepo) List(context.Context, domaincategory.ListFilter, pagination.Params) ([]domaincategory.Category, int64, error) {
	return nil, 0, nil
}
func (catalogCategoryRepo) ProductCounts(context.Context) (map[uuid.UUID]int64, error) {
	return map[uuid.UUID]int64{}, nil
}
func (catalogCategoryRepo) HasChildren(context.Context, uuid.UUID) (bool, error) { return false, nil }
func (catalogCategoryRepo) HasProducts(context.Context, uuid.UUID) (bool, error)   { return false, nil }

func TestGetProduct_ActiveOnly(t *testing.T) {
	id := uuid.New()
	repo := &catalogProductRepo{
		byID: map[uuid.UUID]*domainproduct.Product{
			id: {ID: id, Slug: "draft-item", Status: domainproduct.StatusDraft, Price: 1000},
		},
	}
	svc := NewService(repo, catalogCategoryRepo{}, nil, nil, nil, nil, nil, nil, noopMailer{})

	_, err := svc.GetProduct(context.Background(), id.String())
	if err != domainproduct.ErrNotFound {
		t.Fatalf("GetProduct() error = %v, want ErrNotFound for draft product", err)
	}
}

func TestGetProduct_BySlug(t *testing.T) {
	id := uuid.New()
	repo := &catalogProductRepo{
		bySlug: map[string]*domainproduct.Product{
			"blue-shirt": {ID: id, Slug: "blue-shirt", Name: "Blue Shirt", Status: domainproduct.StatusActive, Price: 120000},
		},
	}
	svc := NewService(repo, catalogCategoryRepo{}, nil, nil, nil, nil, nil, nil, noopMailer{})

	out, err := svc.GetProduct(context.Background(), "blue-shirt")
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}
	if out.Name != "Blue Shirt" || out.PriceToman != 120000 {
		t.Fatalf("unexpected product detail: %+v", out)
	}
}

func TestListProducts_ReturnsActiveOnly(t *testing.T) {
	activeID := uuid.New()
	repo := &catalogProductRepo{
		active: []domainproduct.Product{
			{ID: activeID, Slug: "active-one", Name: "Active", Status: domainproduct.StatusActive, Price: 50000},
		},
	}
	svc := NewService(repo, catalogCategoryRepo{}, nil, nil, nil, nil, nil, nil, noopMailer{})

	out, err := svc.ListProducts(context.Background(), domainproduct.StoreListFilter{}, pagination.Params{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	if len(out.Data) != 1 || out.Data[0].ID != activeID {
		t.Fatalf("unexpected list result: %+v", out)
	}
}
