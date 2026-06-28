package storecontent

import (
	"context"
	"math"

	"github.com/google/uuid"

	domain "app/internal/domain/storecontent"
	domainproduct "app/internal/domain/product"
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

// HomepageProjection is the aggregated public homepage payload.
type HomepageProjection struct {
	Hero            *HomepageHero            `json:"hero,omitempty"`
	ProductSlides   []HomepageProductSlide   `json:"product_slides"`
	ProBanners      []domain.ProBanner       `json:"pro_banners"`
	PartnerBrands   []domain.PartnerBrand    `json:"partner_brands"`
	FAQ             HomepageFAQ              `json:"faq"`
	ContactSection  HomepageContactSection   `json:"contact_section"`
	Testimonials    []domain.HomepageReview  `json:"testimonials"`
	Stats           HomepageStats            `json:"stats"`
}

// HomepageStats holds computed homepage counters.
type HomepageStats struct {
	ProductsCount int64 `json:"products_count"`
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

	return projection, nil
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
			SlideType:          slide.SlideType.String(),
			Title:              slide.Title,
			AutoplayIntervalMs: slide.AutoplayIntervalMs,
			Products:           []HomepageProductCard{},
		}, nil
	}

	cards := make([]HomepageProductCard, 0, len(slide.Items))
	for _, item := range slide.Items {
		product, err := s.products.FindByID(ctx, item.ProductID)
		if err != nil {
			if err == domainproduct.ErrNotFound {
				continue
			}
			return nil, err
		}
		if product.Status != domainproduct.StatusActive {
			continue
		}
		cards = append(cards, toHomepageProductCard(product))
	}

	tabLabel := slide.Title
	if len(slide.Items) > 0 && slide.Items[0].TabLabel != "" {
		tabLabel = slide.Items[0].TabLabel
	}

	return &HomepageProductSlide{
		SlideType:          slide.SlideType.String(),
		Title:              slide.Title,
		TabLabel:           tabLabel,
		AutoplayIntervalMs: slide.AutoplayIntervalMs,
		Products:           cards,
	}, nil
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
