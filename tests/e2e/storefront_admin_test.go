//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestStorefrontAdmin_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/storefront/hero", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestStorefrontAdmin_Hero_RoundTrip(t *testing.T) {
	c := adminClient(t)

	getResp := c.do(http.MethodGet, "/api/v1/admin/storefront/hero", nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	updateBody := map[string]any{
		"title":              fmt.Sprintf("E2E Hero Title %d", testRunID),
		"subtitle":           "E2E hero subtitle for storefront admin test",
		"video_url":          "https://example.com/hero.mp4",
		"cta_primary_text":   "Shop Now",
		"cta_primary_url":    "/products",
		"cta_secondary_text": "Learn More",
		"cta_secondary_url":  "/about",
		"is_active":          true,
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/storefront/hero", updateBody, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var hero struct {
		Title            string `json:"title"`
		Subtitle         string `json:"subtitle"`
		CTAPrimaryText   string `json:"cta_primary_text"`
		CTAPrimaryURL    string `json:"cta_primary_url"`
		CTASecondaryText string `json:"cta_secondary_text"`
		IsActive         bool   `json:"is_active"`
	}
	updateResp.JSON(t, &hero)

	if hero.Title != updateBody["title"] {
		t.Fatalf("title = %q", hero.Title)
	}
	if hero.Subtitle != updateBody["subtitle"] {
		t.Fatalf("subtitle = %q", hero.Subtitle)
	}
	if hero.CTAPrimaryText != "Shop Now" || hero.CTAPrimaryURL != "/products" {
		t.Fatalf("unexpected CTA primary: %+v", hero)
	}
	if hero.CTASecondaryText != "Learn More" {
		t.Fatalf("cta_secondary_text = %q", hero.CTASecondaryText)
	}
	if !hero.IsActive {
		t.Fatal("expected is_active=true")
	}
}

func TestStorefrontAdmin_ProductSlides(t *testing.T) {
	c := adminClient(t)
	productID := createActiveProduct(t, c)

	listResp := c.do(http.MethodGet, "/api/v1/admin/storefront/product-slides", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var slides struct {
		Data []struct {
			SlideType string `json:"slide_type"`
		} `json:"data"`
	}
	listResp.JSON(t, &slides)
	if len(slides.Data) < 3 {
		t.Fatalf("expected seeded slides, got %d", len(slides.Data))
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/storefront/product-slides/featured", map[string]any{
		"title":                fmt.Sprintf("E2E Featured %d", testRunID),
		"autoplay_interval_ms": 5000,
		"is_active":            true,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var slide struct {
		Title              string `json:"title"`
		AutoplayIntervalMs int    `json:"autoplay_interval_ms"`
		IsActive           bool   `json:"is_active"`
	}
	updateResp.JSON(t, &slide)
	if slide.AutoplayIntervalMs != 5000 {
		t.Fatalf("autoplay_interval_ms = %d, want 5000", slide.AutoplayIntervalMs)
	}
	if !slide.IsActive {
		t.Fatal("expected is_active=true")
	}

	itemResp := c.do(http.MethodPost, "/api/v1/admin/storefront/product-slides/featured/items", map[string]any{
		"product_id": productID,
		"sort_order": 1,
		"tab_label":  "Featured pick",
	}, nil)
	itemResp.AssertStatus(t, http.StatusCreated)

	var item struct {
		ID        string `json:"id"`
		ProductID string `json:"product_id"`
		TabLabel  string `json:"tab_label"`
	}
	itemResp.JSON(t, &item)
	if item.ID == "" || item.ProductID != productID {
		t.Fatalf("unexpected slide item: %+v", item)
	}
	if item.TabLabel != "Featured pick" {
		t.Fatalf("tab_label = %q", item.TabLabel)
	}

	updateItemResp := c.do(http.MethodPut, "/api/v1/admin/storefront/product-slide-items/"+item.ID, map[string]any{
		"product_id": productID,
		"sort_order": 2,
		"tab_label":  "Updated tab",
	}, nil)
	updateItemResp.AssertStatus(t, http.StatusOK)

	deleteItemResp := c.do(http.MethodDelete, "/api/v1/admin/storefront/product-slide-items/"+item.ID, nil, nil)
	deleteItemResp.AssertStatus(t, http.StatusNoContent)

	invalidSlideResp := c.do(http.MethodPut, "/api/v1/admin/storefront/product-slides/not-a-type", map[string]any{
		"title": "Bad",
	}, nil)
	invalidSlideResp.AssertStatus(t, http.StatusBadRequest)
	invalidSlideResp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestStorefrontAdmin_ProBanners_CRUD(t *testing.T) {
	c := adminClient(t)

	createResp := c.do(http.MethodPost, "/api/v1/admin/storefront/pro-banners", map[string]any{
		"desktop_image_url": "https://example.com/e2e-banner-desktop.jpg",
		"mobile_image_url":  "https://example.com/e2e-banner-mobile.jpg",
		"link_url":          "https://example.com/sale",
		"sort_order":        1,
		"is_active":         true,
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID              string `json:"id"`
		DesktopImageURL string `json:"desktop_image_url"`
		LinkURL         string `json:"link_url"`
		IsActive        bool   `json:"is_active"`
	}
	createResp.JSON(t, &created)
	if created.ID == "" {
		t.Fatal("expected banner id")
	}
	if created.DesktopImageURL != "https://example.com/e2e-banner-desktop.jpg" {
		t.Fatalf("desktop_image_url = %q", created.DesktopImageURL)
	}
	if !created.IsActive {
		t.Fatal("expected is_active=true")
	}

	listResp := c.do(http.MethodGet, "/api/v1/admin/storefront/pro-banners", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	updateResp := c.do(http.MethodPut, "/api/v1/admin/storefront/pro-banners/"+created.ID, map[string]any{
		"desktop_image_url": "https://example.com/e2e-banner-updated.jpg",
		"link_url":          "https://example.com/updated",
		"sort_order":        2,
		"is_active":         false,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		DesktopImageURL string `json:"desktop_image_url"`
		IsActive        bool   `json:"is_active"`
	}
	updateResp.JSON(t, &updated)
	if updated.DesktopImageURL != "https://example.com/e2e-banner-updated.jpg" {
		t.Fatalf("desktop_image_url = %q", updated.DesktopImageURL)
	}
	if updated.IsActive {
		t.Fatal("expected is_active=false after update")
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/storefront/pro-banners/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)
}

func TestStorefrontAdmin_FAQ_Lifecycle(t *testing.T) {
	c := adminClient(t)

	sectionGetResp := c.do(http.MethodGet, "/api/v1/admin/storefront/faq", nil, nil)
	sectionGetResp.AssertStatus(t, http.StatusOK)

	sectionUpdateResp := c.do(http.MethodPut, "/api/v1/admin/storefront/faq", map[string]any{
		"image_url": "https://example.com/faq-section.jpg",
	}, nil)
	sectionUpdateResp.AssertStatus(t, http.StatusOK)

	var section struct {
		ImageURL string `json:"image_url"`
	}
	sectionUpdateResp.JSON(t, &section)
	if section.ImageURL != "https://example.com/faq-section.jpg" {
		t.Fatalf("image_url = %q", section.ImageURL)
	}

	createItemResp := c.do(http.MethodPost, "/api/v1/admin/storefront/faq/items", map[string]any{
		"question":   "E2E FAQ question?",
		"answer":     "E2E FAQ answer for storefront admin test.",
		"sort_order": 1,
		"is_active":  true,
	}, nil)
	createItemResp.AssertStatus(t, http.StatusCreated)

	var item struct {
		ID       string `json:"id"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
		IsActive bool   `json:"is_active"`
	}
	createItemResp.JSON(t, &item)
	if item.ID == "" || item.Question != "E2E FAQ question?" {
		t.Fatalf("unexpected faq item: %+v", item)
	}
	if !item.IsActive {
		t.Fatal("expected is_active=true")
	}

	listResp := c.do(http.MethodGet, "/api/v1/admin/storefront/faq/items", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	updateItemResp := c.do(http.MethodPut, "/api/v1/admin/storefront/faq/items/"+item.ID, map[string]any{
		"question":   "Updated E2E question?",
		"answer":     "Updated E2E answer.",
		"sort_order": 2,
		"is_active":  false,
	}, nil)
	updateItemResp.AssertStatus(t, http.StatusOK)

	deleteItemResp := c.do(http.MethodDelete, "/api/v1/admin/storefront/faq/items/"+item.ID, nil, nil)
	deleteItemResp.AssertStatus(t, http.StatusNoContent)
}

func TestStorefrontAdmin_HomepageReviews_CRUD(t *testing.T) {
	c := adminClient(t)

	createResp := c.do(http.MethodPost, "/api/v1/admin/storefront/homepage-reviews", map[string]any{
		"customer_name": "E2E Testimonial",
		"photo_url":     "https://example.com/customer.jpg",
		"review_text":   "Excellent products and fast delivery.",
		"rating":        5,
		"sort_order":    1,
		"is_active":     true,
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID           string `json:"id"`
		CustomerName string `json:"customer_name"`
		ReviewText   string `json:"review_text"`
		Rating       int    `json:"rating"`
	}
	createResp.JSON(t, &created)
	if created.ID == "" || created.CustomerName != "E2E Testimonial" {
		t.Fatalf("unexpected review: %+v", created)
	}
	if created.Rating != 5 {
		t.Fatalf("rating = %d, want 5", created.Rating)
	}

	listResp := c.do(http.MethodGet, "/api/v1/admin/storefront/homepage-reviews", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	updateResp := c.do(http.MethodPut, "/api/v1/admin/storefront/homepage-reviews/"+created.ID, map[string]any{
		"customer_name": "Updated Testimonial",
		"review_text":   "Updated review text.",
		"rating":        4,
		"is_active":     false,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/storefront/homepage-reviews/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)
}

func TestStorefrontAdmin_ContactSection_RoundTrip(t *testing.T) {
	c := adminClient(t)

	getResp := c.do(http.MethodGet, "/api/v1/admin/storefront/contact-section", nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	updateResp := c.do(http.MethodPut, "/api/v1/admin/storefront/contact-section", map[string]any{
		"image_url": "https://example.com/contact-section.jpg",
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var section struct {
		ImageURL string `json:"image_url"`
	}
	updateResp.JSON(t, &section)
	if section.ImageURL != "https://example.com/contact-section.jpg" {
		t.Fatalf("image_url = %q", section.ImageURL)
	}
}

func TestStorefrontAdmin_Navigation_RoundTrip(t *testing.T) {
	c := adminClient(t)

	getResp := c.do(http.MethodGet, "/api/v1/admin/storefront/navigation", nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	updateResp := c.do(http.MethodPut, "/api/v1/admin/storefront/navigation", map[string]any{
		"items": []map[string]any{
			{
				"id":         "e2e-home",
				"label":      "E2E Home",
				"url":        "/",
				"sort_order": 0,
				"is_active":  true,
				"children":   []map[string]any{},
			},
		},
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var nav struct {
		Items []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"items"`
	}
	updateResp.JSON(t, &nav)
	if len(nav.Items) != 1 || nav.Items[0].Label != "E2E Home" {
		t.Fatalf("unexpected navigation: %+v", nav)
	}
}

func TestStorefrontAdmin_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name   string
		method string
		path   string
		body   map[string]any
	}{
		{
			name:   "pro banner missing desktop image",
			method: http.MethodPost,
			path:   "/api/v1/admin/storefront/pro-banners",
			body:   map[string]any{"link_url": "https://example.com"},
		},
		{
			name:   "homepage review missing customer name",
			method: http.MethodPost,
			path:   "/api/v1/admin/storefront/homepage-reviews",
			body:   map[string]any{"review_text": "Missing name"},
		},
		{
			name:   "faq item missing answer",
			method: http.MethodPost,
			path:   "/api/v1/admin/storefront/faq/items",
			body:   map[string]any{"question": "Only question?"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(tc.method, tc.path, tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestStorefrontAdmin_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/storefront/hero", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}
