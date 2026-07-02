//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestSettings_Unauthorized(t *testing.T) {
	endpoints := []string{
		"/api/v1/admin/settings/site",
		"/api/v1/admin/settings/contact",
		"/api/v1/admin/settings/social",
		"/api/v1/admin/settings/seo",
		"/api/v1/admin/navigation",
	}

	for _, path := range endpoints {
		t.Run("GET "+path, func(t *testing.T) {
			resp := testClient.do(http.MethodGet, path, nil, nil)
			resp.AssertStatus(t, http.StatusUnauthorized)
		})
	}
}

func TestSettings_Site_RoundTrip(t *testing.T) {
	c := adminClient(t)

	getResp := c.do(http.MethodGet, "/api/v1/admin/settings/site", nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	updateBody := map[string]string{
		"name":        fmt.Sprintf("E2E Shop %d", testRunID),
		"url":         "https://e2e-shop.example.com",
		"logo_url":    "https://e2e-shop.example.com/logo.png",
		"favicon_url": "https://e2e-shop.example.com/favicon.ico",
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/settings/site", updateBody, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var site struct {
		Name       string `json:"name"`
		URL        string `json:"url"`
		LogoURL    string `json:"logo_url"`
		FaviconURL string `json:"favicon_url"`
	}
	updateResp.JSON(t, &site)

	if site.Name != updateBody["name"] {
		t.Fatalf("name = %q, want %q", site.Name, updateBody["name"])
	}
	if site.URL != updateBody["url"] {
		t.Fatalf("url = %q, want %q", site.URL, updateBody["url"])
	}
	if site.LogoURL != updateBody["logo_url"] {
		t.Fatalf("logo_url = %q, want %q", site.LogoURL, updateBody["logo_url"])
	}
	if site.FaviconURL != updateBody["favicon_url"] {
		t.Fatalf("favicon_url = %q, want %q", site.FaviconURL, updateBody["favicon_url"])
	}

	verifyResp := c.do(http.MethodGet, "/api/v1/admin/settings/site", nil, nil)
	verifyResp.AssertStatus(t, http.StatusOK)

	var persisted struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	verifyResp.JSON(t, &persisted)
	if persisted.Name != updateBody["name"] || persisted.URL != updateBody["url"] {
		t.Fatalf("persisted site settings mismatch: %+v", persisted)
	}
}

func TestSettings_Site_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]string
	}{
		{
			name: "missing name",
			body: map[string]string{"url": "https://example.com"},
		},
		{
			name: "invalid url",
			body: map[string]string{"name": "Shop", "url": "not-a-url"},
		},
		{
			name: "invalid logo url",
			body: map[string]string{
				"name":     "Shop",
				"url":      "https://example.com",
				"logo_url": "bad-logo",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPut, "/api/v1/admin/settings/site", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestSettings_Contact_RoundTrip(t *testing.T) {
	c := adminClient(t)

	updateBody := map[string]string{
		"email":   fmt.Sprintf("support-%d@e2e.test", testRunID),
		"phone":   "+982112345678",
		"address": "123 E2E Street",
		"city":    "Tehran",
		"country": "Iran",
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/settings/contact", updateBody, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var contact struct {
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
		City    string `json:"city"`
		Country string `json:"country"`
	}
	updateResp.JSON(t, &contact)

	for key, want := range updateBody {
		got := map[string]string{
			"email":   contact.Email,
			"phone":   contact.Phone,
			"address": contact.Address,
			"city":    contact.City,
			"country": contact.Country,
		}[key]
		if got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestSettings_Contact_InvalidEmail(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodPut, "/api/v1/admin/settings/contact", map[string]string{
		"email": "not-an-email",
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestSettings_Social_RoundTrip(t *testing.T) {
	c := adminClient(t)

	updateBody := map[string]string{
		"facebook":  "https://facebook.com/e2eshop",
		"twitter":   "https://twitter.com/e2eshop",
		"instagram": "https://instagram.com/e2eshop",
		"linkedin":  "https://linkedin.com/company/e2eshop",
		"youtube":   "https://youtube.com/@e2eshop",
		"tiktok":    "https://tiktok.com/@e2eshop",
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/settings/social", updateBody, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	getResp := c.do(http.MethodGet, "/api/v1/admin/settings/social", nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var social struct {
		Instagram string `json:"instagram"`
		YouTube   string `json:"youtube"`
	}
	getResp.JSON(t, &social)
	if social.Instagram != updateBody["instagram"] {
		t.Fatalf("instagram = %q, want %q", social.Instagram, updateBody["instagram"])
	}
	if social.YouTube != updateBody["youtube"] {
		t.Fatalf("youtube = %q, want %q", social.YouTube, updateBody["youtube"])
	}
}

func TestSettings_Social_InvalidURL(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodPut, "/api/v1/admin/settings/social", map[string]string{
		"facebook": "not-a-url",
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestSettings_SEO_RoundTrip(t *testing.T) {
	c := adminClient(t)

	updateBody := map[string]any{
		"meta_title":          fmt.Sprintf("E2E Shop %d", testRunID),
		"meta_description":    "E2E meta description for SEO settings test",
		"meta_keywords":       "e2e,shop,test",
		"og_image_url":        "https://e2e-shop.example.com/og.png",
		"robots_txt":          "User-agent: *\nDisallow:",
		"google_analytics_id": "G-E2ETEST123",
		"sitemap_enabled":     true,
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/settings/seo", updateBody, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var seo struct {
		MetaTitle         string `json:"meta_title"`
		MetaDescription   string `json:"meta_description"`
		GoogleAnalyticsID string `json:"google_analytics_id"`
		SitemapEnabled    bool   `json:"sitemap_enabled"`
	}
	updateResp.JSON(t, &seo)

	if seo.MetaTitle != updateBody["meta_title"] {
		t.Fatalf("meta_title = %q", seo.MetaTitle)
	}
	if seo.MetaDescription != updateBody["meta_description"] {
		t.Fatalf("meta_description = %q", seo.MetaDescription)
	}
	if seo.GoogleAnalyticsID != updateBody["google_analytics_id"] {
		t.Fatalf("google_analytics_id = %q", seo.GoogleAnalyticsID)
	}
	if !seo.SitemapEnabled {
		t.Fatal("expected sitemap_enabled=true")
	}
}

func TestSettings_Navigation_RoundTrip(t *testing.T) {
	c := adminClient(t)

	updateBody := map[string]any{
		"items": []map[string]any{
			{
				"id":         "home",
				"label":      "Home",
				"url":        "/",
				"sort_order": 0,
				"is_active":  true,
				"children":   []map[string]any{},
			},
			{
				"id":         "shop",
				"label":      "Shop",
				"url":        "/products",
				"sort_order": 1,
				"is_active":  true,
				"children": []map[string]any{
					{
						"id":         "new-arrivals",
						"label":      "New Arrivals",
						"url":        "/products/new",
						"sort_order": 0,
						"is_active":  true,
					},
				},
			},
		},
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/navigation", updateBody, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	getResp := c.do(http.MethodGet, "/api/v1/admin/navigation", nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var nav struct {
		Items []struct {
			Label    string `json:"label"`
			URL      string `json:"url"`
			Children []struct {
				Label string `json:"label"`
				URL   string `json:"url"`
			} `json:"children"`
		} `json:"items"`
	}
	getResp.JSON(t, &nav)

	if len(nav.Items) != 2 {
		t.Fatalf("expected 2 nav items, got %d", len(nav.Items))
	}
	if nav.Items[0].Label != "Home" || nav.Items[0].URL != "/" {
		t.Fatalf("unexpected first nav item: %+v", nav.Items[0])
	}
	if len(nav.Items[1].Children) != 1 || nav.Items[1].Children[0].Label != "New Arrivals" {
		t.Fatalf("unexpected nested nav: %+v", nav.Items[1])
	}
}

func TestSettings_CustomerForbidden(t *testing.T) {
	email := uniqueEmail("settings-deny")
	signupResp := testClient.do(http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":      email,
		"password":   "CustomerPass1!",
		"first_name": "Settings",
		"last_name":  "Denied",
	}, nil)
	signupResp.AssertStatus(t, http.StatusCreated)

	var tokens tokenResponse
	signupResp.JSON(t, &tokens)

	resp := testClient.withToken(tokens.AccessToken).do(http.MethodGet, "/api/v1/admin/settings/site", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}
