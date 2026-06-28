package storecontent

import (
	"time"

	"github.com/google/uuid"
)

// SlideType identifies a homepage product carousel.
type SlideType string

const (
	SlideFeatured    SlideType = "featured"
	SlideBestseller  SlideType = "bestseller"
	SlideDiscounted  SlideType = "discounted"
)

// ParseSlideType validates and parses a slide type string.
func ParseSlideType(value string) (SlideType, error) {
	switch SlideType(value) {
	case SlideFeatured, SlideBestseller, SlideDiscounted:
		return SlideType(value), nil
	default:
		return "", ErrInvalidSlideType
	}
}

func (t SlideType) String() string { return string(t) }

// Hero is the singleton homepage hero configuration.
type Hero struct {
	ID               uuid.UUID
	VideoURL         string
	Title            string
	Subtitle         string
	CTAPrimaryText   string
	CTAPrimaryURL    string
	CTASecondaryText string
	CTASecondaryURL  string
	IsActive         bool
	UpdatedAt        time.Time
}

// ProductSlide is a homepage product carousel configuration.
type ProductSlide struct {
	ID                 uuid.UUID
	SlideType          SlideType
	Title              string
	AutoplayIntervalMs int
	IsActive           bool
	SortOrder          int
	Items              []SlideItem
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SlideItem links a product to a carousel slide.
type SlideItem struct {
	ID        uuid.UUID
	SlideID   uuid.UUID
	ProductID uuid.UUID
	SortOrder int
	TabLabel  string
}

// ProBanner is a promotional banner on the homepage.
type ProBanner struct {
	ID              uuid.UUID
	DesktopImageURL string
	MobileImageURL  string
	LinkURL         string
	SortOrder       int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PartnerBrand is a partner brand logo block on the homepage.
type PartnerBrand struct {
	ID          uuid.UUID
	Title       string
	Description string
	LogoURL     string
	LinkURL     string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// HomepageReview is a customer testimonial shown on the homepage.
type HomepageReview struct {
	ID           uuid.UUID
	CustomerName string
	PhotoURL     string
	ReviewText   string
	Rating       *int
	SortOrder    int
	IsActive     bool
	CreatedAt    time.Time
}

// FAQSection is the singleton FAQ section configuration.
type FAQSection struct {
	ID        uuid.UUID
	ImageURL  string
	UpdatedAt time.Time
}

// FAQItem is a single FAQ entry.
type FAQItem struct {
	ID        uuid.UUID
	Question  string
	Answer    string
	SortOrder int
	IsActive  bool
}

// HomepageData aggregates raw homepage content for projection.
type HomepageData struct {
	Hero            *Hero
	ProductSlides   []ProductSlide
	ProBanners      []ProBanner
	PartnerBrands   []PartnerBrand
	FAQSection      *FAQSection
	FAQItems        []FAQItem
	HomepageReviews []HomepageReview
}
