package storecontent

import (
	"context"
	"math"

	"github.com/google/uuid"

	domainblog "app/internal/domain/blog"
	domaincategory "app/internal/domain/category"
	domainorder "app/internal/domain/order"
	domain "app/internal/domain/storecontent"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
	"app/pkg/pagination"
)

// HomepageHero is the public hero projection.
type HomepageHero struct {
	VideoURL      string          `json:"video_url,omitempty"`
	Title         string          `json:"title,omitempty"`
	Subtitle      string          `json:"subtitle,omitempty"`
	CTAPrimary    *HomepageCTA    `json:"cta_primary,omitempty"`
	CTASecondary  *HomepageCTA    `json:"cta_secondary,omitempty"`
}

// HomepageCTA is a call-to-action button on the hero.
type HomepageCTA struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// HomepageProductSlide is a public product carousel projection.
type HomepageProductSlide struct {
	SlideType          string              `json:"slide_type"`
	Title              string              `json:"title"`
	TabLabel           string              `json:"tab_label,omitempty"`
	AutoplayIntervalMs int                 `json:"autoplay_interval_ms"`
	Products           []HomepageProductCard `json:"products"`
}

// HomepageProductCard is a product card on the homepage.
type HomepageProductCard struct {
	ID              uuid.UUID `json:"id"`
	Slug            string    `json:"slug"`
	Name            string    `json:"name"`
	ThumbnailURL    string    `json:"thumbnail_url,omitempty"`
	PriceToman      int64     `json:"price_toman"`
	SalePriceToman  *int64    `json:"sale_price_toman,omitempty"`
	DiscountPercent int       `json:"discount_percent,omitempty"`
	IsOnSale        bool      `json:"is_on_sale"`
	IsOutOfStock    bool      `json:"is_out_of_stock"`
	Brand           string    `json:"brand,omitempty"`
}

// HomepageFAQ is the public FAQ projection.
type HomepageFAQ struct {
	ImageURL string            `json:"image_url,omitempty"`
	Items    []domain.FAQItem  `json:"items"`
}

// HomepageContactSection holds the contact section image.
type HomepageContactSection struct {
	ImageURL string `json:"image_url,omitempty"`
}

// HomepageStats holds computed homepage counters.
type HomepageStats struct {
	ProductsCount         int64 `json:"products_count"`
	CustomersCount        int64 `json:"customers_count"`
	DeliveredOrdersCount  int64 `json:"delivered_orders_count"`
	YearsExperience       int   `json:"years_experience"`
}

// HomepageCategory is a category node on the homepage grid.
type HomepageCategory struct {
	ID            uuid.UUID          `json:"id"`
	Name          string             `json:"name"`
	Slug          string             `json:"slug"`
	ProductsCount int64              `json:"products_count"`
	Children      []HomepageCategory `json:"children,omitempty"`
}

// HomepageBlogPost is a teaser card for the homepage blog section.
type HomepageBlogPost struct {
	ID            uuid.UUID `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Excerpt       string    `json:"excerpt"`
	CoverImageURL string    `json:"cover_image_url,omitempty"`
}

// HomepageBlogTeaser holds latest published posts for the homepage.
type HomepageBlogTeaser struct {
	Posts []HomepageBlogPost `json:"posts"`
}

// HomepageProjection is the aggregated public homepage payload.
type HomepageProjection struct {
	Hero            *HomepageHero            `json:"hero,omitempty"`
	Categories      []HomepageCategory       `json:"categories"`
	ProductSlides   []HomepageProductSlide   `json:"product_slides"`
	ProBanners      []domain.ProBanner       `json:"pro_banners"`
	PartnerBrands   []domain.PartnerBrand    `json:"partner_brands"`
	FAQ             HomepageFAQ              `json:"faq"`
	ContactSection  HomepageContactSection   `json:"contact_section"`
	Testimonials    []domain.HomepageReview  `json:"testimonials"`
	BlogTeaser      HomepageBlogTeaser       `json:"blog_teaser"`
	Stats           HomepageStats            `json:"stats"`
}

// BuildHomepage loads homepage content and projects it for the public storefront.
func (s *Service) BuildHomepage(ctx context.Context) (*HomepageProjection, error) {
	data, err := s.repo.GetHomepageData(ctx)
	if err != nil {
		return nil, err
	}

	projection := &HomepageProjection{
		ProBanners:    data.ProBanners,
		PartnerBrands: data.PartnerBrands,
		Testimonials:  data.HomepageReviews,
		Categories:    []HomepageCategory{},
		BlogTeaser:    HomepageBlogTeaser{Posts: []HomepageBlogPost{}},
	}

	if data.Hero != nil {
		projection.Hero = projectHero(data.Hero)
	}

	for _, slide := range data.ProductSlides {
		projected, err := s.projectProductSlide(ctx, slide)
		if err != nil {
			return nil, err
		}
		if projected != nil {
			projection.ProductSlides = append(projection.ProductSlides, *projected)
		}
	}

	if data.FAQSection != nil {
		projection.FAQ.ImageURL = data.FAQSection.ImageURL
	}
	projection.FAQ.Items = data.FAQItems

	storeSettings, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	projection.ContactSection.ImageURL = storeSettings.ContactSectionImageURL

	count, err := s.products.CountActive(ctx)
	if err != nil {
		return nil, err
	}
	projection.Stats.ProductsCount = count

	if s.customers != nil {
		if customersCount, err := s.customers.Count(ctx); err == nil {
			projection.Stats.CustomersCount = customersCount
		}
	}
	if s.orders != nil {
		if deliveredCount, err := s.orders.CountByStatus(ctx, domainorder.StatusDelivered); err == nil {
			projection.Stats.DeliveredOrdersCount = deliveredCount
		}
	}
	if storeSettings.Site.URL != "" {
		projection.Stats.YearsExperience = yearsSinceSiteLaunch(*storeSettings)
	}

	if s.categories != nil {
		categories, err := s.buildHomepageCategories(ctx)
		if err != nil {
			return nil, err
		}
		projection.Categories = categories
	}

	if s.blogs != nil {
		posts, err := s.buildBlogTeaser(ctx)
		if err != nil {
			return nil, err
		}
		projection.BlogTeaser = HomepageBlogTeaser{Posts: posts}
	}

	return projection, nil
}

func yearsSinceSiteLaunch(settings domainsettings.StoreSettings) int {
	if settings.Site.URL == "" {
		return 5
	}
	// Stable public default until CMS stores an explicit founding year.
	return 5
}

func (s *Service) buildHomepageCategories(ctx context.Context) ([]HomepageCategory, error) {
	active := true
	items, err := s.categories.ListAll(ctx, domaincategory.ListFilter{IsActive: &active})
	if err != nil {
		return nil, err
	}
	counts, err := s.categories.ProductCounts(ctx)
	if err != nil {
		return nil, err
	}
	return buildHomepageCategoryTree(items, counts), nil
}

func buildHomepageCategoryTree(items []domaincategory.Category, counts map[uuid.UUID]int64) []HomepageCategory {
	byID := make(map[uuid.UUID]*HomepageCategory, len(items))
	roots := make([]HomepageCategory, 0)

	for _, item := range items {
		node := HomepageCategory{
			ID:            item.ID,
			Name:          item.Name,
			Slug:          item.Slug,
			ProductsCount: counts[item.ID],
		}
		byID[item.ID] = &node
	}

	for _, item := range items {
		node := byID[item.ID]
		if item.ParentID == nil {
			roots = append(roots, *node)
			continue
		}
		if parent, ok := byID[*item.ParentID]; ok {
			parent.Children = append(parent.Children, *node)
		} else {
			roots = append(roots, *node)
		}
	}

	result := make([]HomepageCategory, 0, len(roots))
	for _, root := range roots {
		if built, ok := byID[root.ID]; ok {
			result = append(result, *built)
		}
	}
	return result
}

func (s *Service) buildBlogTeaser(ctx context.Context) ([]HomepageBlogPost, error) {
	items, _, err := s.blogs.ListPosts(ctx, domainblog.PostFilter{
		Status: string(domainblog.PostStatusPublished),
	}, pagination.Params{Page: 1, PerPage: 3})
	if err != nil {
		return nil, err
	}
	posts := make([]HomepageBlogPost, 0, len(items))
	for _, item := range items {
		posts = append(posts, HomepageBlogPost{
			ID:            item.ID,
			Slug:          item.Slug,
			Title:         item.Title,
			Excerpt:       item.Summary,
			CoverImageURL: item.FeaturedImage,
		})
	}
	return posts, nil
}

func projectHero(hero *domain.Hero) *HomepageHero {
	h := &HomepageHero{
		VideoURL: hero.VideoURL,
		Title:    hero.Title,
		Subtitle: hero.Subtitle,
	}
	if hero.CTAPrimaryText != "" || hero.CTAPrimaryURL != "" {
		h.CTAPrimary = &HomepageCTA{Text: hero.CTAPrimaryText, URL: hero.CTAPrimaryURL}
	}
	if hero.CTASecondaryText != "" || hero.CTASecondaryURL != "" {
		h.CTASecondary = &HomepageCTA{Text: hero.CTASecondaryText, URL: hero.CTASecondaryURL}
	}
	return h
}

func (s *Service) projectProductSlide(ctx context.Context, slide domain.ProductSlide) (*HomepageProductSlide, error) {
	if len(slide.Items) == 0 {
		return &HomepageProductSlide{
			SlideType:          mapSlideTypeForStorefront(slide.SlideType.String()),
			Title:              slide.Title,
			AutoplayIntervalMs: slide.AutoplayIntervalMs,
			Products:           []HomepageProductCard{},
		}, nil
	}

	ids := make([]uuid.UUID, 0, len(slide.Items))
	for _, item := range slide.Items {
		ids = append(ids, item.ProductID)
	}
	products, err := s.products.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]domainproduct.Product, len(products))
	for _, product := range products {
		byID[product.ID] = product
	}

	cards := make([]HomepageProductCard, 0, len(slide.Items))
	for _, item := range slide.Items {
		product, ok := byID[item.ProductID]
		if !ok {
			continue
		}
		cards = append(cards, toHomepageProductCard(&product))
	}

	tabLabel := slide.Title
	if len(slide.Items) > 0 && slide.Items[0].TabLabel != "" {
		tabLabel = slide.Items[0].TabLabel
	}

	return &HomepageProductSlide{
		SlideType:          mapSlideTypeForStorefront(slide.SlideType.String()),
		Title:              slide.Title,
		TabLabel:           tabLabel,
		AutoplayIntervalMs: slide.AutoplayIntervalMs,
		Products:           cards,
	}, nil
}

func mapSlideTypeForStorefront(slideType string) string {
	if slideType == "featured" {
		return "new"
	}
	return slideType
}

func toHomepageProductCard(p *domainproduct.Product) HomepageProductCard {
	card := HomepageProductCard{
		ID:           p.ID,
		Slug:         p.Slug,
		Name:         p.Name,
		PriceToman:   int64(math.Round(p.Price)),
		IsOutOfStock: p.Inventory.IsOutOfStock(),
		Brand:        p.Brand,
	}
	if len(p.Images) > 0 {
		card.ThumbnailURL = p.Images[0].URL
	}
	if p.SalePrice != nil && *p.SalePrice >= 0 && *p.SalePrice < p.Price {
		sale := int64(math.Round(*p.SalePrice))
		card.SalePriceToman = &sale
		card.IsOnSale = true
		if p.Price > 0 {
			card.DiscountPercent = int(math.Round((1 - *p.SalePrice/p.Price) * 100))
		}
	}
	return card
}
