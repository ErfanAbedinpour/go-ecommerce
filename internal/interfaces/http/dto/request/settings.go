package request

// UpdateSiteSettingsRequest updates storefront identity settings.
type UpdateSiteSettingsRequest struct {
	Name       string `json:"name" validate:"required,max=200"`
	URL        string `json:"url" validate:"required,url,max=500"`
	LogoURL    string `json:"logo_url" validate:"omitempty,url,max=500"`
	FaviconURL string `json:"favicon_url" validate:"omitempty,url,max=500"`
}

// UpdateContactSettingsRequest updates public contact information.
type UpdateContactSettingsRequest struct {
	Email   string `json:"email" validate:"omitempty,email,max=255"`
	Phone   string `json:"phone" validate:"omitempty,max=30"`
	Address string `json:"address" validate:"omitempty,max=500"`
	City    string `json:"city" validate:"omitempty,max=100"`
	Country string `json:"country" validate:"omitempty,max=100"`
}

// UpdateSocialSettingsRequest updates social media links.
type UpdateSocialSettingsRequest struct {
	Facebook  string `json:"facebook" validate:"omitempty,url,max=500"`
	Twitter   string `json:"twitter" validate:"omitempty,url,max=500"`
	Instagram string `json:"instagram" validate:"omitempty,url,max=500"`
	LinkedIn  string `json:"linkedin" validate:"omitempty,url,max=500"`
	YouTube   string `json:"youtube" validate:"omitempty,url,max=500"`
	TikTok    string `json:"tiktok" validate:"omitempty,url,max=500"`
}

// UpdateSEOSettingsRequest updates SEO configuration.
type UpdateSEOSettingsRequest struct {
	MetaTitle         string `json:"meta_title" validate:"omitempty,max=200"`
	MetaDescription   string `json:"meta_description" validate:"omitempty,max=500"`
	MetaKeywords      string `json:"meta_keywords" validate:"omitempty,max=500"`
	OGImageURL        string `json:"og_image_url" validate:"omitempty,url,max=500"`
	RobotsTxt         string `json:"robots_txt" validate:"omitempty,max=5000"`
	GoogleAnalyticsID string `json:"google_analytics_id" validate:"omitempty,max=100"`
	SitemapEnabled    bool   `json:"sitemap_enabled"`
}

// NavItemRequest is a navigation menu entry in update requests.
type NavItemRequest struct {
	ID        string           `json:"id" validate:"omitempty,max=100"`
	Label     string           `json:"label" validate:"required,max=200"`
	URL       string           `json:"url" validate:"required,max=500"`
	SortOrder int              `json:"sort_order" validate:"gte=0"`
	IsActive  bool             `json:"is_active"`
	Children  []NavItemRequest `json:"children" validate:"dive"`
}

// UpdateNavigationRequest replaces the navigation menu tree.
type UpdateNavigationRequest struct {
	Items []NavItemRequest `json:"items" validate:"dive"`
}
