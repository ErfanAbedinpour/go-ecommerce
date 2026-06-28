package storecontent

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines persistence for storefront CMS content.
type Repository interface {
	// Singletons
	GetHero(ctx context.Context) (*Hero, error)
	UpdateHero(ctx context.Context, hero *Hero) error
	GetFAQSection(ctx context.Context) (*FAQSection, error)
	UpdateFAQSection(ctx context.Context, section *FAQSection) error

	// Product slides
	ListProductSlides(ctx context.Context) ([]ProductSlide, error)
	GetProductSlide(ctx context.Context, slideType SlideType) (*ProductSlide, error)
	UpdateProductSlide(ctx context.Context, slide *ProductSlide) error
	CreateSlideItem(ctx context.Context, item *SlideItem) error
	UpdateSlideItem(ctx context.Context, item *SlideItem) error
	DeleteSlideItem(ctx context.Context, id uuid.UUID) error

	// Pro banners
	ListProBanners(ctx context.Context) ([]ProBanner, error)
	GetProBanner(ctx context.Context, id uuid.UUID) (*ProBanner, error)
	CreateProBanner(ctx context.Context, banner *ProBanner) error
	UpdateProBanner(ctx context.Context, banner *ProBanner) error
	DeleteProBanner(ctx context.Context, id uuid.UUID) error

	// Partner brands
	ListPartnerBrands(ctx context.Context) ([]PartnerBrand, error)
	GetPartnerBrand(ctx context.Context, id uuid.UUID) (*PartnerBrand, error)
	CreatePartnerBrand(ctx context.Context, brand *PartnerBrand) error
	UpdatePartnerBrand(ctx context.Context, brand *PartnerBrand) error
	DeletePartnerBrand(ctx context.Context, id uuid.UUID) error

	// Homepage reviews
	ListHomepageReviews(ctx context.Context) ([]HomepageReview, error)
	GetHomepageReview(ctx context.Context, id uuid.UUID) (*HomepageReview, error)
	CreateHomepageReview(ctx context.Context, review *HomepageReview) error
	UpdateHomepageReview(ctx context.Context, review *HomepageReview) error
	DeleteHomepageReview(ctx context.Context, id uuid.UUID) error

	// FAQ items
	ListFAQItems(ctx context.Context) ([]FAQItem, error)
	GetFAQItem(ctx context.Context, id uuid.UUID) (*FAQItem, error)
	CreateFAQItem(ctx context.Context, item *FAQItem) error
	UpdateFAQItem(ctx context.Context, item *FAQItem) error
	DeleteFAQItem(ctx context.Context, id uuid.UUID) error

	// Homepage aggregate read
	GetHomepageData(ctx context.Context) (*HomepageData, error)
}
