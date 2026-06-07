package response

import domain "app/internal/domain/settings"

// SiteSettingsResponse is the site identity settings payload.
type SiteSettingsResponse struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	LogoURL    string `json:"logo_url,omitempty"`
	FaviconURL string `json:"favicon_url,omitempty"`
}

// ContactSettingsResponse is the contact settings payload.
type ContactSettingsResponse struct {
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
}

// SocialSettingsResponse is the social media settings payload.
type SocialSettingsResponse struct {
	Facebook  string `json:"facebook,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
	TikTok    string `json:"tiktok,omitempty"`
}

// SEOSettingsResponse is the SEO settings payload.
type SEOSettingsResponse struct {
	MetaTitle         string `json:"meta_title,omitempty"`
	MetaDescription   string `json:"meta_description,omitempty"`
	MetaKeywords      string `json:"meta_keywords,omitempty"`
	OGImageURL        string `json:"og_image_url,omitempty"`
	RobotsTxt         string `json:"robots_txt,omitempty"`
	GoogleAnalyticsID string `json:"google_analytics_id,omitempty"`
	SitemapEnabled    bool   `json:"sitemap_enabled"`
}

// NavItemResponse is a navigation menu entry.
type NavItemResponse struct {
	ID        string            `json:"id"`
	Label     string            `json:"label"`
	URL       string            `json:"url"`
	SortOrder int               `json:"sort_order"`
	IsActive  bool              `json:"is_active"`
	Children  []NavItemResponse `json:"children"`
}

// NavigationResponse is the navigation menu tree.
type NavigationResponse struct {
	Items []NavItemResponse `json:"items"`
}

func ToSiteSettingsResponse(s *domain.Site) SiteSettingsResponse {
	return SiteSettingsResponse{
		Name:       s.Name,
		URL:        s.URL,
		LogoURL:    s.LogoURL,
		FaviconURL: s.FaviconURL,
	}
}

func ToContactSettingsResponse(c *domain.Contact) ContactSettingsResponse {
	return ContactSettingsResponse{
		Email:   c.Email,
		Phone:   c.Phone,
		Address: c.Address,
		City:    c.City,
		Country: c.Country,
	}
}

func ToSocialSettingsResponse(s *domain.Social) SocialSettingsResponse {
	return SocialSettingsResponse{
		Facebook:  s.Facebook,
		Twitter:   s.Twitter,
		Instagram: s.Instagram,
		LinkedIn:  s.LinkedIn,
		YouTube:   s.YouTube,
		TikTok:    s.TikTok,
	}
}

func ToSEOSettingsResponse(s *domain.SEO) SEOSettingsResponse {
	return SEOSettingsResponse{
		MetaTitle:         s.MetaTitle,
		MetaDescription:   s.MetaDescription,
		MetaKeywords:      s.MetaKeywords,
		OGImageURL:        s.OGImageURL,
		RobotsTxt:         s.RobotsTxt,
		GoogleAnalyticsID: s.GoogleAnalyticsID,
		SitemapEnabled:    s.SitemapEnabled,
	}
}

func ToNavigationResponse(items []domain.NavItem) NavigationResponse {
	return NavigationResponse{Items: toNavItemResponses(items)}
}

func toNavItemResponses(items []domain.NavItem) []NavItemResponse {
	if items == nil {
		return []NavItemResponse{}
	}
	result := make([]NavItemResponse, len(items))
	for i, item := range items {
		result[i] = NavItemResponse{
			ID:        item.ID,
			Label:     item.Label,
			URL:       item.URL,
			SortOrder: item.SortOrder,
			IsActive:  item.IsActive,
			Children:  toNavItemResponses(item.Children),
		}
	}
	return result
}
