package storecontent

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/storecontent"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
	"app/pkg/pagination"
)

type mockStoreContentRepo struct {
	slides map[domain.SlideType]*domain.ProductSlide
}

func (m *mockStoreContentRepo) GetHero(context.Context) (*domain.Hero, error) {
	return &domain.Hero{}, nil
}
func (m *mockStoreContentRepo) UpdateHero(context.Context, *domain.Hero) error { return nil }
func (m *mockStoreContentRepo) GetFAQSection(context.Context) (*domain.FAQSection, error) {
	return &domain.FAQSection{}, nil
}
func (m *mockStoreContentRepo) UpdateFAQSection(context.Context, *domain.FAQSection) error {
	return nil
}
func (m *mockStoreContentRepo) ListProductSlides(ctx context.Context) ([]domain.ProductSlide, error) {
	items := make([]domain.ProductSlide, 0, len(m.slides))
	for _, s := range m.slides {
		items = append(items, *s)
	}
	return items, nil
}
func (m *mockStoreContentRepo) GetProductSlide(_ context.Context, slideType domain.SlideType) (*domain.ProductSlide, error) {
	s, ok := m.slides[slideType]
	if !ok {
		return nil, domain.ErrSlideNotFound
	}
	return s, nil
}
func (m *mockStoreContentRepo) UpdateProductSlide(context.Context, *domain.ProductSlide) error {
	return nil
}
func (m *mockStoreContentRepo) CreateSlideItem(_ context.Context, item *domain.SlideItem) error {
	slide, ok := m.slides[domain.SlideFeatured]
	if !ok {
		return domain.ErrSlideNotFound
	}
	slide.Items = append(slide.Items, *item)
	return nil
}
func (m *mockStoreContentRepo) UpdateSlideItem(context.Context, *domain.SlideItem) error { return nil }
func (m *mockStoreContentRepo) DeleteSlideItem(context.Context, uuid.UUID) error         { return nil }
func (m *mockStoreContentRepo) ListProBanners(context.Context) ([]domain.ProBanner, error) {
	return nil, nil
}
func (m *mockStoreContentRepo) GetProBanner(context.Context, uuid.UUID) (*domain.ProBanner, error) {
	return nil, domain.ErrBannerNotFound
}
func (m *mockStoreContentRepo) CreateProBanner(context.Context, *domain.ProBanner) error { return nil }
func (m *mockStoreContentRepo) UpdateProBanner(context.Context, *domain.ProBanner) error { return nil }
func (m *mockStoreContentRepo) DeleteProBanner(context.Context, uuid.UUID) error         { return nil }
func (m *mockStoreContentRepo) ListPartnerBrands(context.Context) ([]domain.PartnerBrand, error) {
	return nil, nil
}
func (m *mockStoreContentRepo) GetPartnerBrand(context.Context, uuid.UUID) (*domain.PartnerBrand, error) {
	return nil, domain.ErrBrandNotFound
}
func (m *mockStoreContentRepo) CreatePartnerBrand(context.Context, *domain.PartnerBrand) error {
	return nil
}
func (m *mockStoreContentRepo) UpdatePartnerBrand(context.Context, *domain.PartnerBrand) error {
	return nil
}
func (m *mockStoreContentRepo) DeletePartnerBrand(context.Context, uuid.UUID) error { return nil }
func (m *mockStoreContentRepo) ListHomepageReviews(context.Context) ([]domain.HomepageReview, error) {
	return nil, nil
}
func (m *mockStoreContentRepo) GetHomepageReview(context.Context, uuid.UUID) (*domain.HomepageReview, error) {
	return nil, domain.ErrReviewNotFound
}
func (m *mockStoreContentRepo) CreateHomepageReview(context.Context, *domain.HomepageReview) error {
	return nil
}
func (m *mockStoreContentRepo) UpdateHomepageReview(context.Context, *domain.HomepageReview) error {
	return nil
}
func (m *mockStoreContentRepo) DeleteHomepageReview(context.Context, uuid.UUID) error { return nil }
func (m *mockStoreContentRepo) ListFAQItems(context.Context) ([]domain.FAQItem, error) {
	return nil, nil
}
func (m *mockStoreContentRepo) GetFAQItem(context.Context, uuid.UUID) (*domain.FAQItem, error) {
	return nil, domain.ErrFAQItemNotFound
}
func (m *mockStoreContentRepo) CreateFAQItem(context.Context, *domain.FAQItem) error { return nil }
func (m *mockStoreContentRepo) UpdateFAQItem(context.Context, *domain.FAQItem) error { return nil }
func (m *mockStoreContentRepo) DeleteFAQItem(context.Context, uuid.UUID) error       { return nil }
func (m *mockStoreContentRepo) GetHomepageData(context.Context) (*domain.HomepageData, error) {
	return &domain.HomepageData{}, nil
}

type mockSlideProductRepo struct {
	products map[uuid.UUID]*domainproduct.Product
}

func (m *mockSlideProductRepo) Create(context.Context, *domainproduct.Product) error { return nil }
func (m *mockSlideProductRepo) Update(context.Context, *domainproduct.Product) error { return nil }
func (m *mockSlideProductRepo) SoftDelete(context.Context, uuid.UUID) error          { return nil }
func (m *mockSlideProductRepo) FindBySlug(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *mockSlideProductRepo) FindBySKU(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *mockSlideProductRepo) List(context.Context, domainproduct.ListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockSlideProductRepo) ListStorefront(context.Context, domainproduct.StoreListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockSlideProductRepo) SearchStorefront(context.Context, string, int) ([]domainproduct.Product, error) {
	return nil, nil
}
func (m *mockSlideProductRepo) ListRelatedStorefront(context.Context, uuid.UUID, int) ([]domainproduct.Product, error) {
	return nil, nil
}
func (m *mockSlideProductRepo) Search(context.Context, string, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockSlideProductRepo) CountActive(context.Context) (int64, error) { return 0, nil }
func (m *mockSlideProductRepo) UpdateInventory(context.Context, uuid.UUID, domainproduct.Inventory) error {
	return nil
}
func (m *mockSlideProductRepo) ExistsInActiveOrders(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockSlideProductRepo) CategoryExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (m *mockSlideProductRepo) GetStats(context.Context) (*domainproduct.Stats, error) {
	return nil, nil
}
func (m *mockSlideProductRepo) FindByID(_ context.Context, id uuid.UUID) (*domainproduct.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domainproduct.ErrNotFound
	}
	return p, nil
}

type mockStoreContentSettingsRepo struct{}

func (mockStoreContentSettingsRepo) Get(context.Context) (*domainsettings.StoreSettings, error) {
	return &domainsettings.StoreSettings{}, nil
}
func (mockStoreContentSettingsRepo) UpdateSite(context.Context, domainsettings.Site) (*domainsettings.Site, error) {
	return nil, nil
}
func (mockStoreContentSettingsRepo) UpdateContact(context.Context, domainsettings.Contact) (*domainsettings.Contact, error) {
	return nil, nil
}
func (mockStoreContentSettingsRepo) UpdateSocial(context.Context, domainsettings.Social) (*domainsettings.Social, error) {
	return nil, nil
}
func (mockStoreContentSettingsRepo) UpdateSEO(context.Context, domainsettings.SEO) (*domainsettings.SEO, error) {
	return nil, nil
}
func (mockStoreContentSettingsRepo) UpdateNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (mockStoreContentSettingsRepo) UpdateStorefrontNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (mockStoreContentSettingsRepo) UpdateContactSectionImage(context.Context, string) (string, error) {
	return "", nil
}

func TestCreateSlideItem_InvalidProduct(t *testing.T) {
	slideID := uuid.New()
	repo := &mockStoreContentRepo{
		slides: map[domain.SlideType]*domain.ProductSlide{
			domain.SlideFeatured: {ID: slideID, SlideType: domain.SlideFeatured},
		},
	}
	svc := NewService(repo, &mockSlideProductRepo{products: map[uuid.UUID]*domainproduct.Product{}}, mockStoreContentSettingsRepo{})

	_, err := svc.CreateSlideItem(context.Background(), "featured", CreateSlideItemInput{
		ProductID: uuid.New(),
	})
	if err != domainproduct.ErrNotFound {
		t.Fatalf("CreateSlideItem() error = %v, want ErrNotFound", err)
	}
}

func TestCreateSlideItem_Success(t *testing.T) {
	slideID := uuid.New()
	productID := uuid.New()
	repo := &mockStoreContentRepo{
		slides: map[domain.SlideType]*domain.ProductSlide{
			domain.SlideFeatured: {ID: slideID, SlideType: domain.SlideFeatured, CreatedAt: time.Now().UTC()},
		},
	}
	products := map[uuid.UUID]*domainproduct.Product{
		productID: {ID: productID, Name: "Featured Product", Status: domainproduct.StatusActive},
	}
	svc := NewService(repo, &mockSlideProductRepo{products: products}, mockStoreContentSettingsRepo{})

	item, err := svc.CreateSlideItem(context.Background(), "featured", CreateSlideItemInput{
		ProductID: productID,
		SortOrder: 1,
		TabLabel:  "New",
	})
	if err != nil {
		t.Fatalf("CreateSlideItem() error = %v", err)
	}
	if item.ProductID != productID || item.TabLabel != "New" {
		t.Fatalf("unexpected slide item: %+v", item)
	}
}
