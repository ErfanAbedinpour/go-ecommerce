package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/storecontent"
	"app/internal/infrastructure/persistence/models"
)

const (
	storefrontHeroID = "a1000000-0000-0000-0000-000000000001"
	faqSectionID     = "c1000000-0000-0000-0000-000000000001"
)

// StoreContentRepository implements storecontent.Repository using GORM.
type StoreContentRepository struct {
	db *gorm.DB
}

// NewStoreContentRepository creates a new StoreContentRepository.
func NewStoreContentRepository(db *gorm.DB) *StoreContentRepository {
	return &StoreContentRepository{db: db}
}

func (r *StoreContentRepository) GetHero(ctx context.Context) (*storecontent.Hero, error) {
	id, err := uuid.Parse(storefrontHeroID)
	if err != nil {
		return nil, err
	}
	var m models.StorefrontHeroModel
	err = r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, storecontent.ErrHeroNotFound
		}
		return nil, err
	}
	return toHeroDomain(&m), nil
}

func (r *StoreContentRepository) UpdateHero(ctx context.Context, hero *storecontent.Hero) error {
	m := toHeroModel(hero)
	result := r.db.WithContext(ctx).Model(&models.StorefrontHeroModel{}).
		Where("id = ?", hero.ID).
		Updates(map[string]any{
			"video_url":           m.VideoURL,
			"title":               m.Title,
			"subtitle":            m.Subtitle,
			"cta_primary_text":    m.CTAPrimaryText,
			"cta_primary_url":     m.CTAPrimaryURL,
			"cta_secondary_text":  m.CTASecondaryText,
			"cta_secondary_url":   m.CTASecondaryURL,
			"is_active":           m.IsActive,
			"updated_at":          time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrHeroNotFound
	}
	return nil
}

func (r *StoreContentRepository) GetFAQSection(ctx context.Context) (*storecontent.FAQSection, error) {
	id, err := uuid.Parse(faqSectionID)
	if err != nil {
		return nil, err
	}
	var m models.FAQSectionModel
	err = r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, storecontent.ErrNotFound
		}
		return nil, err
	}
	return toFAQSectionDomain(&m), nil
}

func (r *StoreContentRepository) UpdateFAQSection(ctx context.Context, section *storecontent.FAQSection) error {
	m := toFAQSectionModel(section)
	result := r.db.WithContext(ctx).Model(&models.FAQSectionModel{}).
		Where("id = ?", section.ID).
		Updates(map[string]any{
			"image_url":  m.ImageURL,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrNotFound
	}
	return nil
}

func (r *StoreContentRepository) ListProductSlides(ctx context.Context) ([]storecontent.ProductSlide, error) {
	var items []models.ProductSlideModel
	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("sort_order ASC, slide_type ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toProductSlidesDomain(items), nil
}

func (r *StoreContentRepository) GetProductSlide(ctx context.Context, slideType storecontent.SlideType) (*storecontent.ProductSlide, error) {
	var m models.ProductSlideModel
	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("slide_type = ?", slideType.String()).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, storecontent.ErrSlideNotFound
		}
		return nil, err
	}
	return toProductSlideDomain(&m), nil
}

func (r *StoreContentRepository) UpdateProductSlide(ctx context.Context, slide *storecontent.ProductSlide) error {
	result := r.db.WithContext(ctx).Model(&models.ProductSlideModel{}).
		Where("id = ?", slide.ID).
		Updates(map[string]any{
			"title":                slide.Title,
			"autoplay_interval_ms": slide.AutoplayIntervalMs,
			"is_active":            slide.IsActive,
			"sort_order":           slide.SortOrder,
			"updated_at":           time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrSlideNotFound
	}
	return nil
}

func (r *StoreContentRepository) CreateSlideItem(ctx context.Context, item *storecontent.SlideItem) error {
	m := toSlideItemModel(item)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	item.ID = m.ID
	return nil
}

func (r *StoreContentRepository) UpdateSlideItem(ctx context.Context, item *storecontent.SlideItem) error {
	m := toSlideItemModel(item)
	result := r.db.WithContext(ctx).Model(&models.ProductSlideItemModel{}).
		Where("id = ?", item.ID).
		Updates(map[string]any{
			"slide_id":   m.SlideID,
			"product_id": m.ProductID,
			"sort_order": m.SortOrder,
			"tab_label":  m.TabLabel,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrSlideItemNotFound
	}
	return nil
}

func (r *StoreContentRepository) DeleteSlideItem(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.ProductSlideItemModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrSlideItemNotFound
	}
	return nil
}

func (r *StoreContentRepository) ListProBanners(ctx context.Context) ([]storecontent.ProBanner, error) {
	var items []models.ProBannerModel
	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toProBannersDomain(items), nil
}

func (r *StoreContentRepository) GetProBanner(ctx context.Context, id uuid.UUID) (*storecontent.ProBanner, error) {
	var m models.ProBannerModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, storecontent.ErrBannerNotFound
		}
		return nil, err
	}
	return toProBannerDomain(&m), nil
}

func (r *StoreContentRepository) CreateProBanner(ctx context.Context, banner *storecontent.ProBanner) error {
	m := toProBannerModel(banner)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	banner.ID = m.ID
	banner.CreatedAt = m.CreatedAt
	banner.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *StoreContentRepository) UpdateProBanner(ctx context.Context, banner *storecontent.ProBanner) error {
	m := toProBannerModel(banner)
	result := r.db.WithContext(ctx).Model(&models.ProBannerModel{}).
		Where("id = ?", banner.ID).
		Updates(map[string]any{
			"desktop_image_url": m.DesktopImageURL,
			"mobile_image_url":  m.MobileImageURL,
			"link_url":          m.LinkURL,
			"sort_order":        m.SortOrder,
			"is_active":         m.IsActive,
			"updated_at":        time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrBannerNotFound
	}
	return nil
}

func (r *StoreContentRepository) DeleteProBanner(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.ProBannerModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrBannerNotFound
	}
	return nil
}

func (r *StoreContentRepository) ListPartnerBrands(ctx context.Context) ([]storecontent.PartnerBrand, error) {
	var items []models.PartnerBrandModel
	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toPartnerBrandsDomain(items), nil
}

func (r *StoreContentRepository) GetPartnerBrand(ctx context.Context, id uuid.UUID) (*storecontent.PartnerBrand, error) {
	var m models.PartnerBrandModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, storecontent.ErrBrandNotFound
		}
		return nil, err
	}
	return toPartnerBrandDomain(&m), nil
}

func (r *StoreContentRepository) CreatePartnerBrand(ctx context.Context, brand *storecontent.PartnerBrand) error {
	m := toPartnerBrandModel(brand)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	brand.ID = m.ID
	brand.CreatedAt = m.CreatedAt
	brand.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *StoreContentRepository) UpdatePartnerBrand(ctx context.Context, brand *storecontent.PartnerBrand) error {
	m := toPartnerBrandModel(brand)
	result := r.db.WithContext(ctx).Model(&models.PartnerBrandModel{}).
		Where("id = ?", brand.ID).
		Updates(map[string]any{
			"title":       m.Title,
			"description": m.Description,
			"logo_url":    m.LogoURL,
			"link_url":    m.LinkURL,
			"sort_order":  m.SortOrder,
			"is_active":   m.IsActive,
			"updated_at":  time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrBrandNotFound
	}
	return nil
}

func (r *StoreContentRepository) DeletePartnerBrand(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.PartnerBrandModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrBrandNotFound
	}
	return nil
}

func (r *StoreContentRepository) ListHomepageReviews(ctx context.Context) ([]storecontent.HomepageReview, error) {
	var items []models.HomepageReviewModel
	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toHomepageReviewsDomain(items), nil
}

func (r *StoreContentRepository) GetHomepageReview(ctx context.Context, id uuid.UUID) (*storecontent.HomepageReview, error) {
	var m models.HomepageReviewModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, storecontent.ErrReviewNotFound
		}
		return nil, err
	}
	return toHomepageReviewDomain(&m), nil
}

func (r *StoreContentRepository) CreateHomepageReview(ctx context.Context, review *storecontent.HomepageReview) error {
	m := toHomepageReviewModel(review)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	review.ID = m.ID
	review.CreatedAt = m.CreatedAt
	return nil
}

func (r *StoreContentRepository) UpdateHomepageReview(ctx context.Context, review *storecontent.HomepageReview) error {
	m := toHomepageReviewModel(review)
	result := r.db.WithContext(ctx).Model(&models.HomepageReviewModel{}).
		Where("id = ?", review.ID).
		Updates(map[string]any{
			"customer_name": m.CustomerName,
			"photo_url":     m.PhotoURL,
			"review_text":   m.ReviewText,
			"rating":        m.Rating,
			"sort_order":    m.SortOrder,
			"is_active":     m.IsActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrReviewNotFound
	}
	return nil
}

func (r *StoreContentRepository) DeleteHomepageReview(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.HomepageReviewModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrReviewNotFound
	}
	return nil
}

func (r *StoreContentRepository) ListFAQItems(ctx context.Context) ([]storecontent.FAQItem, error) {
	var items []models.FAQItemModel
	err := r.db.WithContext(ctx).Order("sort_order ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toFAQItemsDomain(items), nil
}

func (r *StoreContentRepository) GetFAQItem(ctx context.Context, id uuid.UUID) (*storecontent.FAQItem, error) {
	var m models.FAQItemModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, storecontent.ErrFAQItemNotFound
		}
		return nil, err
	}
	return toFAQItemDomain(&m), nil
}

func (r *StoreContentRepository) CreateFAQItem(ctx context.Context, item *storecontent.FAQItem) error {
	m := toFAQItemModel(item)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	item.ID = m.ID
	return nil
}

func (r *StoreContentRepository) UpdateFAQItem(ctx context.Context, item *storecontent.FAQItem) error {
	m := toFAQItemModel(item)
	result := r.db.WithContext(ctx).Model(&models.FAQItemModel{}).
		Where("id = ?", item.ID).
		Updates(map[string]any{
			"question":   m.Question,
			"answer":     m.Answer,
			"sort_order": m.SortOrder,
			"is_active":  m.IsActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrFAQItemNotFound
	}
	return nil
}

func (r *StoreContentRepository) DeleteFAQItem(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.FAQItemModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return storecontent.ErrFAQItemNotFound
	}
	return nil
}

func (r *StoreContentRepository) GetHomepageData(ctx context.Context) (*storecontent.HomepageData, error) {
	data := &storecontent.HomepageData{}

	hero, err := r.GetHero(ctx)
	if err == nil && hero.IsActive {
		data.Hero = hero
	} else if err != nil && !errors.Is(err, storecontent.ErrHeroNotFound) {
		return nil, err
	}

	slides, err := r.ListProductSlides(ctx)
	if err != nil {
		return nil, err
	}
	for _, slide := range slides {
		if slide.IsActive {
			data.ProductSlides = append(data.ProductSlides, slide)
		}
	}

	banners, err := r.ListProBanners(ctx)
	if err != nil {
		return nil, err
	}
	for _, b := range banners {
		if b.IsActive {
			data.ProBanners = append(data.ProBanners, b)
		}
	}

	brands, err := r.ListPartnerBrands(ctx)
	if err != nil {
		return nil, err
	}
	for _, b := range brands {
		if b.IsActive {
			data.PartnerBrands = append(data.PartnerBrands, b)
		}
	}

	section, err := r.GetFAQSection(ctx)
	if err != nil && !errors.Is(err, storecontent.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		data.FAQSection = section
	}

	faqItems, err := r.ListFAQItems(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range faqItems {
		if item.IsActive {
			data.FAQItems = append(data.FAQItems, item)
		}
	}

	reviews, err := r.ListHomepageReviews(ctx)
	if err != nil {
		return nil, err
	}
	for _, review := range reviews {
		if review.IsActive {
			data.HomepageReviews = append(data.HomepageReviews, review)
		}
	}

	return data, nil
}

var _ storecontent.Repository = (*StoreContentRepository)(nil)
