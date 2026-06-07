package settings

import "time"

// Site holds storefront identity settings.
type Site struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	LogoURL    string `json:"logo_url"`
	FaviconURL string `json:"favicon_url"`
}

// Contact holds public contact information.
type Contact struct {
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	City    string `json:"city"`
	Country string `json:"country"`
}

// Social holds social media profile URLs.
type Social struct {
	Facebook  string `json:"facebook"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	LinkedIn  string `json:"linkedin"`
	YouTube   string `json:"youtube"`
	TikTok    string `json:"tiktok"`
}

// SEO holds search-engine and metadata configuration.
type SEO struct {
	MetaTitle         string `json:"meta_title"`
	MetaDescription   string `json:"meta_description"`
	MetaKeywords      string `json:"meta_keywords"`
	OGImageURL        string `json:"og_image_url"`
	RobotsTxt         string `json:"robots_txt"`
	GoogleAnalyticsID string `json:"google_analytics_id"`
	SitemapEnabled    bool   `json:"sitemap_enabled"`
}

// NavItem is a navigation menu entry with optional children.
type NavItem struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	Children  []NavItem `json:"children"`
}

// StoreSettings is the aggregate of all store configuration sections.
type StoreSettings struct {
	Site       Site
	Contact    Contact
	Social     Social
	SEO        SEO
	Navigation []NavItem
	UpdatedAt  time.Time
}
