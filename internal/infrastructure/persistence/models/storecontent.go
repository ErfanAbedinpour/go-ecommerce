package models

import (
	"time"

	"github.com/google/uuid"
)

// StorefrontHeroModel is the GORM model for storefront_hero table.
type StorefrontHeroModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	VideoURL         *string   `gorm:"type:varchar(500)"`
	Title            *string   `gorm:"type:varchar(255)"`
	Subtitle         *string   `gorm:"type:text"`
	CTAPrimaryText   *string   `gorm:"column:cta_primary_text;type:varchar(100)"`
	CTAPrimaryURL    *string   `gorm:"column:cta_primary_url;type:varchar(500)"`
	CTASecondaryText *string   `gorm:"column:cta_secondary_text;type:varchar(100)"`
	CTASecondaryURL  *string   `gorm:"column:cta_secondary_url;type:varchar(500)"`
	IsActive         bool      `gorm:"not null;default:true"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (StorefrontHeroModel) TableName() string { return "storefront_hero" }

// ProductSlideModel is the GORM model for product_slides table.
type ProductSlideModel struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SlideType          string              `gorm:"type:varchar(50);not null;uniqueIndex"`
	Title              *string             `gorm:"type:varchar(255)"`
	AutoplayIntervalMs int                 `gorm:"not null;default:4500"`
	IsActive           bool                `gorm:"not null;default:true"`
	SortOrder          int                 `gorm:"not null;default:0"`
	Items              []ProductSlideItemModel `gorm:"foreignKey:SlideID"`
	CreatedAt          time.Time           `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt          time.Time           `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (ProductSlideModel) TableName() string { return "product_slides" }

// ProductSlideItemModel is the GORM model for product_slide_items table.
type ProductSlideItemModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SlideID   uuid.UUID `gorm:"type:uuid;not null;index:idx_product_slide_items_slide,priority:1"`
	ProductID uuid.UUID `gorm:"type:uuid;not null"`
	SortOrder int       `gorm:"not null;default:0;index:idx_product_slide_items_slide,priority:2"`
	TabLabel  *string   `gorm:"type:varchar(100)"`
}

func (ProductSlideItemModel) TableName() string { return "product_slide_items" }

// ProBannerModel is the GORM model for pro_banners table.
type ProBannerModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DesktopImageURL string    `gorm:"type:varchar(500);not null"`
	MobileImageURL  *string   `gorm:"type:varchar(500)"`
	LinkURL         *string   `gorm:"type:varchar(500)"`
	SortOrder       int       `gorm:"not null;default:0"`
	IsActive        bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (ProBannerModel) TableName() string { return "pro_banners" }

// PartnerBrandModel is the GORM model for partner_brands table.
type PartnerBrandModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description *string   `gorm:"type:text"`
	LogoURL     string    `gorm:"type:varchar(500);not null"`
	LinkURL     *string   `gorm:"type:varchar(500)"`
	SortOrder   int       `gorm:"not null;default:0"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (PartnerBrandModel) TableName() string { return "partner_brands" }

// HomepageReviewModel is the GORM model for homepage_reviews table.
type HomepageReviewModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerName string    `gorm:"type:varchar(255);not null"`
	PhotoURL     *string   `gorm:"type:varchar(500)"`
	ReviewText   string    `gorm:"type:text;not null"`
	Rating       *int16    `gorm:"type:smallint"`
	SortOrder    int       `gorm:"not null;default:0"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (HomepageReviewModel) TableName() string { return "homepage_reviews" }

// FAQSectionModel is the GORM model for faq_sections table.
type FAQSectionModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ImageURL  *string   `gorm:"type:varchar(500)"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (FAQSectionModel) TableName() string { return "faq_sections" }

// FAQItemModel is the GORM model for faq_items table.
type FAQItemModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Question  string    `gorm:"type:text;not null"`
	Answer    string    `gorm:"type:text;not null"`
	SortOrder int       `gorm:"not null;default:0"`
	IsActive  bool      `gorm:"not null;default:true"`
}

func (FAQItemModel) TableName() string { return "faq_items" }
