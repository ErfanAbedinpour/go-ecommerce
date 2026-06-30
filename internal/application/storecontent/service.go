package storecontent

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainblog "app/internal/domain/blog"
	domaincategory "app/internal/domain/category"
	domaincustomer "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	domain "app/internal/domain/storecontent"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
)

// Service handles storefront CMS content use cases.
type Service struct {
	repo       domain.Repository
	products   domainproduct.Repository
	settings   domainsettings.Repository
	categories domaincategory.Repository
	blogs      domainblog.Repository
	customers  domaincustomer.Repository
	orders     domainorder.Repository
}

// NewService creates a new storecontent Service.
func NewService(
	repo domain.Repository,
	products domainproduct.Repository,
	settings domainsettings.Repository,
	categories domaincategory.Repository,
	blogs domainblog.Repository,
	customers domaincustomer.Repository,
	orders domainorder.Repository,
) *Service {
	return &Service{
		repo:       repo,
		products:   products,
		settings:   settings,
		categories: categories,
		blogs:      blogs,
		customers:  customers,
		orders:     orders,
	}
}

// GetHero returns the singleton hero configuration.
func (s *Service) GetHero(ctx context.Context) (*domain.Hero, error) {
	return s.repo.GetHero(ctx)
}

// UpdateHeroInput holds hero update data.
type UpdateHeroInput struct {
	VideoURL         string
	Title            string
	Subtitle         string
	CTAPrimaryText   string
	CTAPrimaryURL    string
	CTASecondaryText string
	CTASecondaryURL  string
	IsActive         bool
}

// UpdateHero updates the singleton hero configuration.
func (s *Service) UpdateHero(ctx context.Context, input UpdateHeroInput) (*domain.Hero, error) {
	hero, err := s.repo.GetHero(ctx)
	if err != nil {
		return nil, err
	}

	hero.VideoURL = input.VideoURL
	hero.Title = input.Title
	hero.Subtitle = input.Subtitle
	hero.CTAPrimaryText = input.CTAPrimaryText
	hero.CTAPrimaryURL = input.CTAPrimaryURL
	hero.CTASecondaryText = input.CTASecondaryText
	hero.CTASecondaryURL = input.CTASecondaryURL
	hero.IsActive = input.IsActive
	hero.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateHero(ctx, hero); err != nil {
		return nil, err
	}
	return hero, nil
}

// ListProductSlides returns all product slide configurations.
func (s *Service) ListProductSlides(ctx context.Context) ([]domain.ProductSlide, error) {
	return s.repo.ListProductSlides(ctx)
}

// GetProductSlide returns a slide by type.
func (s *Service) GetProductSlide(ctx context.Context, slideType string) (*domain.ProductSlide, error) {
	parsed, err := domain.ParseSlideType(slideType)
	if err != nil {
		return nil, err
	}
	return s.repo.GetProductSlide(ctx, parsed)
}

// UpdateProductSlideInput holds slide update data.
type UpdateProductSlideInput struct {
	Title              string
	AutoplayIntervalMs int
	IsActive           bool
	SortOrder          int
}

// UpdateProductSlide updates a product slide configuration.
func (s *Service) UpdateProductSlide(ctx context.Context, slideType string, input UpdateProductSlideInput) (*domain.ProductSlide, error) {
	parsed, err := domain.ParseSlideType(slideType)
	if err != nil {
		return nil, err
	}

	slide, err := s.repo.GetProductSlide(ctx, parsed)
	if err != nil {
		return nil, err
	}

	slide.Title = input.Title
	slide.AutoplayIntervalMs = input.AutoplayIntervalMs
	slide.IsActive = input.IsActive
	slide.SortOrder = input.SortOrder
	slide.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateProductSlide(ctx, slide); err != nil {
		return nil, err
	}
	return s.repo.GetProductSlide(ctx, parsed)
}

// CreateSlideItemInput holds data for adding a product to a slide.
type CreateSlideItemInput struct {
	ProductID uuid.UUID
	SortOrder int
	TabLabel  string
}

// CreateSlideItem adds a product to a slide.
func (s *Service) CreateSlideItem(ctx context.Context, slideType string, input CreateSlideItemInput) (*domain.SlideItem, error) {
	parsed, err := domain.ParseSlideType(slideType)
	if err != nil {
		return nil, err
	}

	slide, err := s.repo.GetProductSlide(ctx, parsed)
	if err != nil {
		return nil, err
	}

	if _, err := s.products.FindByID(ctx, input.ProductID); err != nil {
		return nil, err
	}

	item := &domain.SlideItem{
		ID:        uuid.New(),
		SlideID:   slide.ID,
		ProductID: input.ProductID,
		SortOrder: input.SortOrder,
		TabLabel:  input.TabLabel,
	}
	if err := s.repo.CreateSlideItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateSlideItemInput holds slide item update data.
type UpdateSlideItemInput struct {
	SortOrder int
	TabLabel  string
}

// UpdateSlideItem updates a slide item.
func (s *Service) UpdateSlideItem(ctx context.Context, id uuid.UUID, input UpdateSlideItemInput) (*domain.SlideItem, error) {
	slides, err := s.repo.ListProductSlides(ctx)
	if err != nil {
		return nil, err
	}

	var target *domain.SlideItem
	for _, slide := range slides {
		for i := range slide.Items {
			if slide.Items[i].ID == id {
				target = &slide.Items[i]
				break
			}
		}
		if target != nil {
			break
		}
	}
	if target == nil {
		return nil, domain.ErrSlideItemNotFound
	}

	target.SortOrder = input.SortOrder
	target.TabLabel = input.TabLabel
	if err := s.repo.UpdateSlideItem(ctx, target); err != nil {
		return nil, err
	}
	return target, nil
}

// DeleteSlideItem removes a product from a slide.
func (s *Service) DeleteSlideItem(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteSlideItem(ctx, id)
}

// ListProBanners returns all pro banners.
func (s *Service) ListProBanners(ctx context.Context) ([]domain.ProBanner, error) {
	return s.repo.ListProBanners(ctx)
}

// CreateProBannerInput holds pro banner creation data.
type CreateProBannerInput struct {
	DesktopImageURL string
	MobileImageURL  string
	LinkURL         string
	SortOrder       int
	IsActive        bool
}

// CreateProBanner creates a pro banner.
func (s *Service) CreateProBanner(ctx context.Context, input CreateProBannerInput) (*domain.ProBanner, error) {
	now := time.Now().UTC()
	banner := &domain.ProBanner{
		ID:              uuid.New(),
		DesktopImageURL: input.DesktopImageURL,
		MobileImageURL:  input.MobileImageURL,
		LinkURL:         input.LinkURL,
		SortOrder:       input.SortOrder,
		IsActive:        input.IsActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateProBanner(ctx, banner); err != nil {
		return nil, err
	}
	return banner, nil
}

// UpdateProBannerInput holds pro banner update data.
type UpdateProBannerInput struct {
	DesktopImageURL *string
	MobileImageURL  *string
	LinkURL         *string
	SortOrder       *int
	IsActive        *bool
}

// UpdateProBanner updates a pro banner.
func (s *Service) UpdateProBanner(ctx context.Context, id uuid.UUID, input UpdateProBannerInput) (*domain.ProBanner, error) {
	banner, err := s.repo.GetProBanner(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.DesktopImageURL != nil {
		banner.DesktopImageURL = *input.DesktopImageURL
	}
	if input.MobileImageURL != nil {
		banner.MobileImageURL = *input.MobileImageURL
	}
	if input.LinkURL != nil {
		banner.LinkURL = *input.LinkURL
	}
	if input.SortOrder != nil {
		banner.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		banner.IsActive = *input.IsActive
	}
	banner.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateProBanner(ctx, banner); err != nil {
		return nil, err
	}
	return banner, nil
}

// DeleteProBanner deletes a pro banner.
func (s *Service) DeleteProBanner(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteProBanner(ctx, id)
}

// ListPartnerBrands returns all partner brands.
func (s *Service) ListPartnerBrands(ctx context.Context) ([]domain.PartnerBrand, error) {
	return s.repo.ListPartnerBrands(ctx)
}

// CreatePartnerBrandInput holds partner brand creation data.
type CreatePartnerBrandInput struct {
	Title       string
	Description string
	LogoURL     string
	LinkURL     string
	SortOrder   int
	IsActive    bool
}

// CreatePartnerBrand creates a partner brand.
func (s *Service) CreatePartnerBrand(ctx context.Context, input CreatePartnerBrandInput) (*domain.PartnerBrand, error) {
	now := time.Now().UTC()
	brand := &domain.PartnerBrand{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		LogoURL:     input.LogoURL,
		LinkURL:     input.LinkURL,
		SortOrder:   input.SortOrder,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreatePartnerBrand(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}

// UpdatePartnerBrandInput holds partner brand update data.
type UpdatePartnerBrandInput struct {
	Title       *string
	Description *string
	LogoURL     *string
	LinkURL     *string
	SortOrder   *int
	IsActive    *bool
}

// UpdatePartnerBrand updates a partner brand.
func (s *Service) UpdatePartnerBrand(ctx context.Context, id uuid.UUID, input UpdatePartnerBrandInput) (*domain.PartnerBrand, error) {
	brand, err := s.repo.GetPartnerBrand(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		brand.Title = *input.Title
	}
	if input.Description != nil {
		brand.Description = *input.Description
	}
	if input.LogoURL != nil {
		brand.LogoURL = *input.LogoURL
	}
	if input.LinkURL != nil {
		brand.LinkURL = *input.LinkURL
	}
	if input.SortOrder != nil {
		brand.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		brand.IsActive = *input.IsActive
	}
	brand.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdatePartnerBrand(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}

// DeletePartnerBrand deletes a partner brand.
func (s *Service) DeletePartnerBrand(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePartnerBrand(ctx, id)
}

// ListHomepageReviews returns all homepage reviews.
func (s *Service) ListHomepageReviews(ctx context.Context) ([]domain.HomepageReview, error) {
	return s.repo.ListHomepageReviews(ctx)
}

// CreateHomepageReviewInput holds homepage review creation data.
type CreateHomepageReviewInput struct {
	CustomerName string
	PhotoURL     string
	ReviewText   string
	Rating       *int
	SortOrder    int
	IsActive     bool
}

// CreateHomepageReview creates a homepage review.
func (s *Service) CreateHomepageReview(ctx context.Context, input CreateHomepageReviewInput) (*domain.HomepageReview, error) {
	review := &domain.HomepageReview{
		ID:           uuid.New(),
		CustomerName: input.CustomerName,
		PhotoURL:     input.PhotoURL,
		ReviewText:   input.ReviewText,
		Rating:       input.Rating,
		SortOrder:    input.SortOrder,
		IsActive:     input.IsActive,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.CreateHomepageReview(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

// UpdateHomepageReviewInput holds homepage review update data.
type UpdateHomepageReviewInput struct {
	CustomerName *string
	PhotoURL     *string
	ReviewText   *string
	Rating       *int
	SortOrder    *int
	IsActive     *bool
}

// UpdateHomepageReview updates a homepage review.
func (s *Service) UpdateHomepageReview(ctx context.Context, id uuid.UUID, input UpdateHomepageReviewInput) (*domain.HomepageReview, error) {
	review, err := s.repo.GetHomepageReview(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.CustomerName != nil {
		review.CustomerName = *input.CustomerName
	}
	if input.PhotoURL != nil {
		review.PhotoURL = *input.PhotoURL
	}
	if input.ReviewText != nil {
		review.ReviewText = *input.ReviewText
	}
	if input.Rating != nil {
		review.Rating = input.Rating
	}
	if input.SortOrder != nil {
		review.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		review.IsActive = *input.IsActive
	}

	if err := s.repo.UpdateHomepageReview(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

// DeleteHomepageReview deletes a homepage review.
func (s *Service) DeleteHomepageReview(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteHomepageReview(ctx, id)
}

// GetFAQSection returns the singleton FAQ section configuration.
func (s *Service) GetFAQSection(ctx context.Context) (*domain.FAQSection, error) {
	return s.repo.GetFAQSection(ctx)
}

// UpdateFAQSectionInput holds FAQ section update data.
type UpdateFAQSectionInput struct {
	ImageURL string
}

// UpdateFAQSection updates the singleton FAQ section.
func (s *Service) UpdateFAQSection(ctx context.Context, input UpdateFAQSectionInput) (*domain.FAQSection, error) {
	section, err := s.repo.GetFAQSection(ctx)
	if err != nil {
		return nil, err
	}
	section.ImageURL = input.ImageURL
	section.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateFAQSection(ctx, section); err != nil {
		return nil, err
	}
	return section, nil
}

// ListFAQItems returns all FAQ items.
func (s *Service) ListFAQItems(ctx context.Context) ([]domain.FAQItem, error) {
	return s.repo.ListFAQItems(ctx)
}

// CreateFAQItemInput holds FAQ item creation data.
type CreateFAQItemInput struct {
	Question  string
	Answer    string
	SortOrder int
	IsActive  bool
}

// CreateFAQItem creates an FAQ item.
func (s *Service) CreateFAQItem(ctx context.Context, input CreateFAQItemInput) (*domain.FAQItem, error) {
	item := &domain.FAQItem{
		ID:        uuid.New(),
		Question:  input.Question,
		Answer:    input.Answer,
		SortOrder: input.SortOrder,
		IsActive:  input.IsActive,
	}
	if err := s.repo.CreateFAQItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateFAQItemInput holds FAQ item update data.
type UpdateFAQItemInput struct {
	Question  *string
	Answer    *string
	SortOrder *int
	IsActive  *bool
}

// UpdateFAQItem updates an FAQ item.
func (s *Service) UpdateFAQItem(ctx context.Context, id uuid.UUID, input UpdateFAQItemInput) (*domain.FAQItem, error) {
	item, err := s.repo.GetFAQItem(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Question != nil {
		item.Question = *input.Question
	}
	if input.Answer != nil {
		item.Answer = *input.Answer
	}
	if input.SortOrder != nil {
		item.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		item.IsActive = *input.IsActive
	}

	if err := s.repo.UpdateFAQItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteFAQItem deletes an FAQ item.
func (s *Service) DeleteFAQItem(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteFAQItem(ctx, id)
}
