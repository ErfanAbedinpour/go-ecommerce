package storefront

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domainbrand "app/internal/domain/brand"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

type catalogExtraProductRepo struct {
	products map[uuid.UUID]*domainproduct.Product
	bySlug   map[string]*domainproduct.Product
	related  []domainproduct.Product
	search   []domainproduct.Product
}

func newCatalogExtraProductRepo() *catalogExtraProductRepo {
	return &catalogExtraProductRepo{
		products: make(map[uuid.UUID]*domainproduct.Product),
		bySlug:   make(map[string]*domainproduct.Product),
	}
}

func (m *catalogExtraProductRepo) Create(context.Context, *domainproduct.Product) error { return nil }
func (m *catalogExtraProductRepo) Update(context.Context, *domainproduct.Product) error { return nil }
func (m *catalogExtraProductRepo) SoftDelete(context.Context, uuid.UUID) error          { return nil }
func (m *catalogExtraProductRepo) FindBySKU(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *catalogExtraProductRepo) List(context.Context, domainproduct.ListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *catalogExtraProductRepo) ListStorefront(context.Context, domainproduct.StoreListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *catalogExtraProductRepo) Search(context.Context, string, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *catalogExtraProductRepo) CountActive(context.Context) (int64, error) { return 0, nil }
func (m *catalogExtraProductRepo) UpdateInventory(context.Context, uuid.UUID, domainproduct.Inventory) error {
	return nil
}
func (m *catalogExtraProductRepo) ExistsInActiveOrders(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *catalogExtraProductRepo) CategoryExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (m *catalogExtraProductRepo) GetStats(context.Context) (*domainproduct.Stats, error) {
	return nil, nil
}
func (m *catalogExtraProductRepo) FindByID(_ context.Context, id uuid.UUID) (*domainproduct.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domainproduct.ErrNotFound
	}
	return p, nil
}
func (m *catalogExtraProductRepo) FindBySlug(_ context.Context, slug string) (*domainproduct.Product, error) {
	p, ok := m.bySlug[slug]
	if !ok {
		return nil, domainproduct.ErrNotFound
	}
	return p, nil
}
func (m *catalogExtraProductRepo) SearchStorefront(_ context.Context, _ string, _ int) ([]domainproduct.Product, error) {
	return m.search, nil
}
func (m *catalogExtraProductRepo) ListRelatedStorefront(_ context.Context, _ uuid.UUID, _ int) ([]domainproduct.Product, error) {
	return m.related, nil
}
func (m *catalogExtraProductRepo) FindByIDs(_ context.Context, ids []uuid.UUID) ([]domainproduct.Product, error) {
	out := make([]domainproduct.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			out = append(out, *p)
		}
	}
	return out, nil
}

type catalogExtraBrandRepo struct {
	brands []domainbrand.Brand
}

func (m *catalogExtraBrandRepo) Create(context.Context, *domainbrand.Brand) error { return nil }
func (m *catalogExtraBrandRepo) Update(context.Context, *domainbrand.Brand) error { return nil }
func (m *catalogExtraBrandRepo) SoftDelete(context.Context, uuid.UUID) error      { return nil }
func (m *catalogExtraBrandRepo) FindByID(context.Context, uuid.UUID) (*domainbrand.Brand, error) {
	return nil, domainbrand.ErrNotFound
}
func (m *catalogExtraBrandRepo) FindBySlug(context.Context, string) (*domainbrand.Brand, error) {
	return nil, domainbrand.ErrNotFound
}
func (m *catalogExtraBrandRepo) List(context.Context, domainbrand.ListFilter, pagination.Params) ([]domainbrand.Brand, int64, error) {
	return m.brands, int64(len(m.brands)), nil
}
func (m *catalogExtraBrandRepo) HasProducts(context.Context, string) (bool, error) {
	return false, nil
}

func TestSearchProducts(t *testing.T) {
	repo := newCatalogExtraProductRepo()
	id := uuid.New()
	repo.search = []domainproduct.Product{
		{ID: id, Slug: "tile-60", Name: "Tile 60x60", Price: 399000, Status: domainproduct.StatusActive},
	}

	svc := NewService(repo, catalogCategoryRepo{}, &catalogExtraBrandRepo{}, nil, nil, nil, nil, nil, nil, noopMailer{})
	result, err := svc.SearchProducts(context.Background(), "tile", 5)
	if err != nil {
		t.Fatalf("SearchProducts() error = %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].Name != "Tile 60x60" {
		t.Fatalf("unexpected search result: %+v", result.Data)
	}
}

func TestListRelatedProducts(t *testing.T) {
	repo := newCatalogExtraProductRepo()
	id := uuid.New()
	repo.products[id] = &domainproduct.Product{ID: id, Slug: "main-tile", Status: domainproduct.StatusActive}
	repo.related = []domainproduct.Product{
		{ID: uuid.New(), Slug: "related-tile", Name: "Related Tile", Price: 250000, Status: domainproduct.StatusActive},
	}

	svc := NewService(repo, catalogCategoryRepo{}, &catalogExtraBrandRepo{}, nil, nil, nil, nil, nil, nil, noopMailer{})
	result, err := svc.ListRelatedProducts(context.Background(), id.String(), 4)
	if err != nil {
		t.Fatalf("ListRelatedProducts() error = %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 related product, got %d", len(result.Data))
	}
}

func TestListBrands(t *testing.T) {
	brandID := uuid.New()
	brands := &catalogExtraBrandRepo{
		brands: []domainbrand.Brand{{ID: brandID, Name: "Paryab", Slug: "paryab", IsActive: true}},
	}
	svc := NewService(newCatalogExtraProductRepo(), catalogCategoryRepo{}, brands, nil, nil, nil, nil, nil, nil, noopMailer{})

	result, err := svc.ListBrands(context.Background())
	if err != nil {
		t.Fatalf("ListBrands() error = %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].Slug != "paryab" {
		t.Fatalf("unexpected brands: %+v", result.Data)
	}
}

func TestGetShippingMethods(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, noopMailer{})

	result, err := svc.GetShippingMethods(context.Background(), "Tehran")
	if err != nil {
		t.Fatalf("GetShippingMethods() error = %v", err)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected courier + post in Tehran, got %d", len(result.Data))
	}

	result, err = svc.GetShippingMethods(context.Background(), "Shiraz")
	if err != nil {
		t.Fatalf("GetShippingMethods() error = %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].Code != "post" {
		t.Fatalf("expected post only outside courier cities, got %+v", result.Data)
	}
}
