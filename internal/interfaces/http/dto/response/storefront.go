package response

import (
	"time"

	apptheme "app/internal/application/theme"
	domain "app/internal/domain/storecontent"
	domaintheme "app/internal/domain/theme"
)

// HeroResponse is the admin hero configuration payload.
type HeroResponse struct {
	ID               string    `json:"id"`
	VideoURL         string    `json:"video_url,omitempty"`
	Title            string    `json:"title,omitempty"`
	Subtitle         string    `json:"subtitle,omitempty"`
	CTAPrimaryText   string    `json:"cta_primary_text,omitempty"`
	CTAPrimaryURL    string    `json:"cta_primary_url,omitempty"`
	CTASecondaryText string    `json:"cta_secondary_text,omitempty"`
	CTASecondaryURL  string    `json:"cta_secondary_url,omitempty"`
	IsActive         bool      `json:"is_active"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SlideItemResponse is a product slide item.
type SlideItemResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	SortOrder int    `json:"sort_order"`
	TabLabel  string `json:"tab_label,omitempty"`
}

// ProductSlideResponse is a product carousel configuration.
type ProductSlideResponse struct {
	ID                 string              `json:"id"`
	SlideType          string              `json:"slide_type"`
	Title              string              `json:"title,omitempty"`
	AutoplayIntervalMs int                 `json:"autoplay_interval_ms"`
	IsActive           bool                `json:"is_active"`
	SortOrder          int                 `json:"sort_order"`
	Items              []SlideItemResponse `json:"items"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// ProductSlideListResponse is a list of product slides.
type ProductSlideListResponse struct {
	Data []ProductSlideResponse `json:"data"`
}

// ProBannerResponse is a promotional banner.
type ProBannerResponse struct {
	ID              string    `json:"id"`
	DesktopImageURL string    `json:"desktop_image_url"`
	MobileImageURL  string    `json:"mobile_image_url,omitempty"`
	LinkURL         string    `json:"link_url,omitempty"`
	SortOrder       int       `json:"sort_order"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ProBannerListResponse is a list of pro banners.
type ProBannerListResponse struct {
	Data []ProBannerResponse `json:"data"`
}

// PartnerBrandResponse is a partner brand block.
type PartnerBrandResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	LogoURL     string    `json:"logo_url"`
	LinkURL     string    `json:"link_url,omitempty"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PartnerBrandListResponse is a list of partner brands.
type PartnerBrandListResponse struct {
	Data []PartnerBrandResponse `json:"data"`
}

// HomepageReviewResponse is a homepage testimonial.
type HomepageReviewResponse struct {
	ID           string    `json:"id"`
	CustomerName string    `json:"customer_name"`
	PhotoURL     string    `json:"photo_url,omitempty"`
	ReviewText   string    `json:"review_text"`
	Rating       *int      `json:"rating,omitempty"`
	SortOrder    int       `json:"sort_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// HomepageReviewListResponse is a list of homepage reviews.
type HomepageReviewListResponse struct {
	Data []HomepageReviewResponse `json:"data"`
}

// FAQSectionResponse is the FAQ section configuration.
type FAQSectionResponse struct {
	ID        string    `json:"id"`
	ImageURL  string    `json:"image_url,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FAQItemResponse is an FAQ Q&A item.
type FAQItemResponse struct {
	ID        string `json:"id"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

// FAQItemListResponse is a list of FAQ items.
type FAQItemListResponse struct {
	Data []FAQItemResponse `json:"data"`
}

// ContactSectionResponse is the contact section image payload.
type ContactSectionResponse struct {
	ImageURL string `json:"image_url,omitempty"`
}

// ThemeResponse is a theme catalog entry for admin listing.
type ThemeResponse struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	Slug            string                  `json:"slug"`
	Description     string                  `json:"description,omitempty"`
	PreviewImageURL string                  `json:"preview_image_url,omitempty"`
	Price           float64                 `json:"price"`
	IsActive        bool                    `json:"is_active"`
	DefaultColors   domaintheme.ColorTokens `json:"default_colors"`
	DefaultFont     string                  `json:"default_font,omitempty"`
	IsPurchased     bool                    `json:"is_purchased"`
	IsActiveTheme   bool                    `json:"is_active_theme"`
	CreatedAt       time.Time               `json:"created_at"`
}

// ThemeListResponse is a list of themes.
type ThemeListResponse struct {
	Data []ThemeResponse `json:"data"`
}

// ThemePurchaseResponse is a theme purchase confirmation.
type ThemePurchaseResponse struct {
	ThemeID     string    `json:"theme_id"`
	PurchasedAt time.Time `json:"purchased_at"`
}

// ThemeSummaryResponse is a compact theme reference.
type ThemeSummaryResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	PreviewImageURL string `json:"preview_image_url,omitempty"`
}

// StoreStyleResponse is the admin store style payload.
type StoreStyleResponse struct {
	ID            string                  `json:"id"`
	ActiveThemeID *string                 `json:"active_theme_id,omitempty"`
	ActiveTheme   *ThemeSummaryResponse   `json:"active_theme,omitempty"`
	Colors        domaintheme.ColorTokens `json:"colors"`
	FontFamily    string                  `json:"font_family,omitempty"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

func ToHeroResponse(h *domain.Hero) HeroResponse {
	return HeroResponse{
		ID:               h.ID.String(),
		VideoURL:         h.VideoURL,
		Title:            h.Title,
		Subtitle:         h.Subtitle,
		CTAPrimaryText:   h.CTAPrimaryText,
		CTAPrimaryURL:    h.CTAPrimaryURL,
		CTASecondaryText: h.CTASecondaryText,
		CTASecondaryURL:  h.CTASecondaryURL,
		IsActive:         h.IsActive,
		UpdatedAt:        h.UpdatedAt,
	}
}

func ToSlideItemResponse(item domain.SlideItem) SlideItemResponse {
	return SlideItemResponse{
		ID:        item.ID.String(),
		ProductID: item.ProductID.String(),
		SortOrder: item.SortOrder,
		TabLabel:  item.TabLabel,
	}
}

func ToProductSlideResponse(slide *domain.ProductSlide) ProductSlideResponse {
	items := make([]SlideItemResponse, len(slide.Items))
	for i, item := range slide.Items {
		items[i] = ToSlideItemResponse(item)
	}
	return ProductSlideResponse{
		ID:                 slide.ID.String(),
		SlideType:          slide.SlideType.String(),
		Title:              slide.Title,
		AutoplayIntervalMs: slide.AutoplayIntervalMs,
		IsActive:           slide.IsActive,
		SortOrder:          slide.SortOrder,
		Items:              items,
		CreatedAt:          slide.CreatedAt,
		UpdatedAt:          slide.UpdatedAt,
	}
}

func ToProductSlideListResponse(slides []domain.ProductSlide) ProductSlideListResponse {
	data := make([]ProductSlideResponse, len(slides))
	for i := range slides {
		data[i] = ToProductSlideResponse(&slides[i])
	}
	return ProductSlideListResponse{Data: data}
}

func ToProBannerResponse(b *domain.ProBanner) ProBannerResponse {
	return ProBannerResponse{
		ID:              b.ID.String(),
		DesktopImageURL: b.DesktopImageURL,
		MobileImageURL:  b.MobileImageURL,
		LinkURL:         b.LinkURL,
		SortOrder:       b.SortOrder,
		IsActive:        b.IsActive,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

func ToProBannerListResponse(banners []domain.ProBanner) ProBannerListResponse {
	data := make([]ProBannerResponse, len(banners))
	for i := range banners {
		data[i] = ToProBannerResponse(&banners[i])
	}
	return ProBannerListResponse{Data: data}
}

func ToPartnerBrandResponse(b *domain.PartnerBrand) PartnerBrandResponse {
	return PartnerBrandResponse{
		ID:          b.ID.String(),
		Title:       b.Title,
		Description: b.Description,
		LogoURL:     b.LogoURL,
		LinkURL:     b.LinkURL,
		SortOrder:   b.SortOrder,
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func ToPartnerBrandListResponse(brands []domain.PartnerBrand) PartnerBrandListResponse {
	data := make([]PartnerBrandResponse, len(brands))
	for i := range brands {
		data[i] = ToPartnerBrandResponse(&brands[i])
	}
	return PartnerBrandListResponse{Data: data}
}

func ToHomepageReviewResponse(r *domain.HomepageReview) HomepageReviewResponse {
	return HomepageReviewResponse{
		ID:           r.ID.String(),
		CustomerName: r.CustomerName,
		PhotoURL:     r.PhotoURL,
		ReviewText:   r.ReviewText,
		Rating:       r.Rating,
		SortOrder:    r.SortOrder,
		IsActive:     r.IsActive,
		CreatedAt:    r.CreatedAt,
	}
}

func ToHomepageReviewListResponse(reviews []domain.HomepageReview) HomepageReviewListResponse {
	data := make([]HomepageReviewResponse, len(reviews))
	for i := range reviews {
		data[i] = ToHomepageReviewResponse(&reviews[i])
	}
	return HomepageReviewListResponse{Data: data}
}

func ToFAQSectionResponse(s *domain.FAQSection) FAQSectionResponse {
	return FAQSectionResponse{
		ID:        s.ID.String(),
		ImageURL:  s.ImageURL,
		UpdatedAt: s.UpdatedAt,
	}
}

func ToFAQItemResponse(item *domain.FAQItem) FAQItemResponse {
	return FAQItemResponse{
		ID:        item.ID.String(),
		Question:  item.Question,
		Answer:    item.Answer,
		SortOrder: item.SortOrder,
		IsActive:  item.IsActive,
	}
}

func ToFAQItemListResponse(items []domain.FAQItem) FAQItemListResponse {
	data := make([]FAQItemResponse, len(items))
	for i := range items {
		data[i] = ToFAQItemResponse(&items[i])
	}
	return FAQItemListResponse{Data: data}
}

func ToContactSectionResponse(imageURL string) ContactSectionResponse {
	return ContactSectionResponse{ImageURL: imageURL}
}

func ToThemeResponse(item apptheme.ThemeListItem) ThemeResponse {
	return ThemeResponse{
		ID:              item.ID.String(),
		Name:            item.Name,
		Slug:            item.Slug,
		Description:     item.Description,
		PreviewImageURL: item.PreviewImageURL,
		Price:           item.Price,
		IsActive:        item.IsActive,
		DefaultColors:   item.DefaultColors,
		DefaultFont:     item.DefaultFont,
		IsPurchased:     item.IsPurchased,
		IsActiveTheme:   item.IsActiveTheme,
		CreatedAt:       item.CreatedAt,
	}
}

func ToThemeListResponse(items []apptheme.ThemeListItem) ThemeListResponse {
	data := make([]ThemeResponse, len(items))
	for i, item := range items {
		data[i] = ToThemeResponse(item)
	}
	return ThemeListResponse{Data: data}
}

func ToThemePurchaseResponse(p *domaintheme.ThemePurchase) ThemePurchaseResponse {
	return ThemePurchaseResponse{
		ThemeID:     p.ThemeID.String(),
		PurchasedAt: p.PurchasedAt,
	}
}

func ToStoreStyleResponse(output *apptheme.StyleOutput) StoreStyleResponse {
	resp := StoreStyleResponse{
		Colors:     output.Colors,
		FontFamily: output.FontFamily,
	}
	if output.Style != nil {
		resp.ID = output.Style.ID.String()
		resp.UpdatedAt = output.Style.UpdatedAt
		if output.Style.ActiveThemeID != nil {
			id := output.Style.ActiveThemeID.String()
			resp.ActiveThemeID = &id
		}
	}
	if output.ActiveTheme != nil {
		resp.ActiveTheme = &ThemeSummaryResponse{
			ID:              output.ActiveTheme.ID.String(),
			Name:            output.ActiveTheme.Name,
			Slug:            output.ActiveTheme.Slug,
			PreviewImageURL: output.ActiveTheme.PreviewImageURL,
		}
	}
	return resp
}
