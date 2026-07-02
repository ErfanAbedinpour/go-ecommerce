//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestProduct_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/products", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestProduct_Create_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing name",
			body: map[string]any{
				"price": 100000,
				"inventory": map[string]any{"quantity": 5},
			},
		},
		{
			name: "negative price",
			body: map[string]any{
				"name":  "Invalid Price Product",
				"price": -1,
				"inventory": map[string]any{"quantity": 1},
			},
		},
		{
			name: "invalid image url",
			body: map[string]any{
				"name":  "Bad Image Product",
				"price": 50000,
				"images": []map[string]any{
					{"url": "not-a-url", "sort_order": 0},
				},
				"inventory": map[string]any{"quantity": 1},
			},
		},
		{
			name: "invalid status",
			body: map[string]any{
				"name":     "Bad Status Product",
				"price":    50000,
				"status":   "published",
				"inventory": map[string]any{"quantity": 1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPost, "/api/v1/admin/products", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestProduct_CRUD_Lifecycle(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-product-%d", testRunID)

	createBody := map[string]any{
		"name":              "E2E Test Product",
		"slug":              slug,
		"description":       "Full description for e2e test",
		"short_description": "Short e2e description",
		"price":             250000,
		"sale_price":        199000,
		"brand":             "E2E Brand",
		"is_featured":       true,
		"status":            "active",
		"images": []map[string]any{
			{
				"url":        "https://example.com/images/e2e-product.jpg",
				"alt_text":   "E2E product image",
				"sort_order": 0,
			},
		},
		"attributes": []map[string]any{
			{"name": "Color", "values": []string{"Red", "Blue"}},
			{"name": "Size", "values": []string{"M", "L"}},
		},
		"inventory": map[string]any{
			"quantity":            25,
			"low_stock_threshold": 5,
		},
	}

	createResp := c.do(http.MethodPost, "/api/v1/admin/products", createBody, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID               string   `json:"id"`
		Name             string   `json:"name"`
		Slug             string   `json:"slug"`
		Price            float64  `json:"price"`
		SalePrice        *float64 `json:"sale_price"`
		Brand            string   `json:"brand"`
		Status           string   `json:"status"`
		IsFeatured       bool     `json:"is_featured"`
		ShortDescription string   `json:"short_description"`
		Inventory        struct {
			Quantity          int `json:"quantity"`
			LowStockThreshold int `json:"low_stock_threshold"`
		} `json:"inventory"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
		Attributes []struct {
			Name   string   `json:"name"`
			Values []string `json:"values"`
		} `json:"attributes"`
		SKUs []struct {
			Code string `json:"code"`
		} `json:"skus"`
	}
	createResp.JSON(t, &created)

	if created.ID == "" {
		t.Fatal("expected product id")
	}
	if created.Name != "E2E Test Product" {
		t.Fatalf("name = %q", created.Name)
	}
	if created.Slug != slug {
		t.Fatalf("slug = %q, want %q", created.Slug, slug)
	}
	if created.Price != 250000 {
		t.Fatalf("price = %v, want 250000", created.Price)
	}
	if created.SalePrice == nil || *created.SalePrice != 199000 {
		t.Fatalf("sale_price = %v, want 199000", created.SalePrice)
	}
	if created.Status != "active" {
		t.Fatalf("status = %q, want active", created.Status)
	}
	if !created.IsFeatured {
		t.Fatal("expected is_featured=true")
	}
	if created.Inventory.Quantity != 25 {
		t.Fatalf("inventory quantity = %d, want 25", created.Inventory.Quantity)
	}
	if len(created.Images) != 1 || created.Images[0].URL != "https://example.com/images/e2e-product.jpg" {
		t.Fatalf("unexpected images: %+v", created.Images)
	}
	if len(created.Attributes) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(created.Attributes))
	}
	if len(created.SKUs) != 4 {
		t.Fatalf("expected 4 SKUs (2x2 variants), got %d", len(created.SKUs))
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/products/"+created.ID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var fetched struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	getResp.JSON(t, &fetched)
	if fetched.ID != created.ID || fetched.Name != created.Name {
		t.Fatalf("get mismatch: %+v", fetched)
	}

	listResp := c.do(http.MethodGet, "/api/v1/admin/products?page=1&per_page=5&status=active", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Page    int   `json:"page"`
			PerPage int   `json:"per_page"`
			Total   int64 `json:"total"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Page != 1 || list.Meta.PerPage != 5 {
		t.Fatalf("unexpected pagination meta: %+v", list.Meta)
	}
	if len(list.Data) == 0 {
		t.Fatal("expected at least one product in list")
	}

	searchResp := c.do(http.MethodGet, "/api/v1/admin/products/search?q=E2E+Test", nil, nil)
	searchResp.AssertStatus(t, http.StatusOK)

	var search struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	searchResp.JSON(t, &search)
	if len(search.Data) == 0 {
		t.Fatal("expected search results for E2E Test")
	}

	updateResp := c.do(http.MethodPut, "/api/v1/admin/products/"+created.ID, map[string]any{
		"name":        "E2E Updated Product",
		"description": "Updated description",
		"price":       275000,
		"is_featured": false,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		Name       string  `json:"name"`
		Price      float64 `json:"price"`
		IsFeatured bool    `json:"is_featured"`
	}
	updateResp.JSON(t, &updated)
	if updated.Name != "E2E Updated Product" {
		t.Fatalf("updated name = %q", updated.Name)
	}
	if updated.Price != 275000 {
		t.Fatalf("updated price = %v", updated.Price)
	}
	if updated.IsFeatured {
		t.Fatal("expected is_featured=false after update")
	}

	invResp := c.do(http.MethodPatch, "/api/v1/admin/products/"+created.ID+"/inventory", map[string]any{
		"quantity":           10,
		"low_stock_threshold": 3,
		"adjustment_reason":  "e2e stock adjustment",
	}, nil)
	invResp.AssertStatus(t, http.StatusOK)

	var inventory struct {
		Quantity          int  `json:"quantity"`
		LowStockThreshold int  `json:"low_stock_threshold"`
		IsLowStock        bool `json:"is_low_stock"`
		IsOutOfStock      bool `json:"is_out_of_stock"`
	}
	invResp.JSON(t, &inventory)
	if inventory.Quantity != 10 {
		t.Fatalf("inventory quantity = %d, want 10", inventory.Quantity)
	}
	if inventory.LowStockThreshold != 3 {
		t.Fatalf("low_stock_threshold = %d, want 3", inventory.LowStockThreshold)
	}
	if inventory.IsOutOfStock {
		t.Fatal("expected product to be in stock")
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/products/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := c.do(http.MethodGet, "/api/v1/admin/products/"+created.ID, nil, nil)
	notFoundResp.AssertStatus(t, http.StatusNotFound)
	notFoundResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestProduct_SlugConflict(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-slug-conflict-%d", testRunID)

	firstBody := map[string]any{
		"name":      "First Product",
		"slug":      slug,
		"price":     10000,
		"inventory": map[string]any{"quantity": 1},
	}
	firstResp := c.do(http.MethodPost, "/api/v1/admin/products", firstBody, nil)
	firstResp.AssertStatus(t, http.StatusCreated)

	var first struct {
		ID string `json:"id"`
	}
	firstResp.JSON(t, &first)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/products/"+first.ID, nil, nil)
	})

	secondBody := map[string]any{
		"name":      "Second Product",
		"slug":      slug,
		"price":     20000,
		"inventory": map[string]any{"quantity": 1},
	}
	secondResp := c.do(http.MethodPost, "/api/v1/admin/products", secondBody, nil)
	secondResp.AssertStatus(t, http.StatusConflict)
	secondResp.AssertErrorCode(t, "CONFLICT")
}

func TestProduct_Get_InvalidAndMissing(t *testing.T) {
	c := adminClient(t)

	invalidResp := c.do(http.MethodGet, "/api/v1/admin/products/not-a-uuid", nil, nil)
	// uuid.Parse failure is currently surfaced as 500; documents current API behavior.
	invalidResp.AssertStatus(t, http.StatusInternalServerError)

	missingResp := c.do(http.MethodGet, "/api/v1/admin/products/00000000-0000-0000-0000-000000000099", nil, nil)
	missingResp.AssertStatus(t, http.StatusNotFound)
	missingResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestProduct_Create_MinimalPayload(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodPost, "/api/v1/admin/products", map[string]any{
		"name":  fmt.Sprintf("Minimal %d", testRunID),
		"price": 99000,
		"inventory": map[string]any{
			"quantity": 0,
		},
	}, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var product struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
		Status string  `json:"status"`
		Slug   string  `json:"slug"`
	}
	resp.JSON(t, &product)

	if product.ID == "" || product.Slug == "" {
		t.Fatalf("expected generated id and slug: %+v", product)
	}
	if product.Price != 99000 {
		t.Fatalf("price = %v, want 99000", product.Price)
	}

	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/products/"+product.ID, nil, nil)
	})
}
