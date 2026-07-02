//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestStorefront_PublicContent(t *testing.T) {
	store := storeClient(t)

	endpoints := []string{
		"/api/v1/store/homepage",
		"/api/v1/store/settings",
		"/api/v1/store/settings/checkout",
		"/api/v1/store/navigation",
		"/api/v1/store/about",
		"/api/v1/store/theme",
		"/api/v1/store/checkout/shipping-methods?city=Tehran",
	}

	for _, path := range endpoints {
		t.Run(path, func(t *testing.T) {
			resp := store.do(http.MethodGet, path, nil, nil)
			resp.AssertStatus(t, http.StatusOK)
		})
	}
}

func TestStorefront_Products_ListSearchDetail(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)

	n := uniqueCounter.Add(1)
	name := fmt.Sprintf("E2E Storefront Product %d-%d", testRunID, n)
	slug := fmt.Sprintf("e2e-storefront-product-%d-%d", testRunID, n)

	createResp := admin.do(http.MethodPost, "/api/v1/admin/products", map[string]any{
		"name":        name,
		"slug":        slug,
		"price":       75000,
		"status":      "active",
		"description": "Visible on public storefront",
		"inventory": map[string]any{
			"quantity": 15,
		},
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID string `json:"id"`
	}
	createResp.JSON(t, &created)
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/products/"+created.ID, nil, nil)
	})

	listResp := store.do(http.MethodGet, "/api/v1/store/products?page=1&per_page=50", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID         string `json:"id"`
			Slug       string `json:"slug"`
			Name       string `json:"name"`
			PriceToman int64  `json:"price_toman"`
		} `json:"data"`
		Meta struct {
			Page int `json:"page"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Page != 1 {
		t.Fatalf("page = %d, want 1", list.Meta.Page)
	}

	foundInList := false
	for _, p := range list.Data {
		if p.ID == created.ID {
			foundInList = true
			if p.Slug != slug {
				t.Fatalf("slug = %q, want %q", p.Slug, slug)
			}
			if p.Name != name {
				t.Fatalf("name = %q, want %q", p.Name, name)
			}
			if p.PriceToman != 75000 {
				t.Fatalf("price_toman = %d, want 75000", p.PriceToman)
			}
			break
		}
	}
	if !foundInList {
		t.Fatalf("product %s not found in storefront list", created.ID)
	}

	searchResp := store.do(http.MethodGet, "/api/v1/store/products/search?q="+url.QueryEscape(name), nil, nil)
	searchResp.AssertStatus(t, http.StatusOK)

	var search struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	searchResp.JSON(t, &search)
	if len(search.Data) == 0 {
		t.Fatal("expected search hits")
	}

	bySlugResp := store.do(http.MethodGet, "/api/v1/store/products/"+slug, nil, nil)
	bySlugResp.AssertStatus(t, http.StatusOK)

	var bySlug struct {
		ID          string `json:"id"`
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	bySlugResp.JSON(t, &bySlug)
	if bySlug.ID != created.ID || bySlug.Slug != slug {
		t.Fatalf("get by slug mismatch: %+v", bySlug)
	}
	if bySlug.Description != "Visible on public storefront" {
		t.Fatalf("description = %q", bySlug.Description)
	}

	byIDResp := store.do(http.MethodGet, "/api/v1/store/products/"+created.ID, nil, nil)
	byIDResp.AssertStatus(t, http.StatusOK)

	var byID struct {
		ID string `json:"id"`
	}
	byIDResp.JSON(t, &byID)
	if byID.ID != created.ID {
		t.Fatalf("get by id = %q, want %q", byID.ID, created.ID)
	}

	relatedResp := store.do(http.MethodGet, "/api/v1/store/products/"+created.ID+"/related?limit=4", nil, nil)
	relatedResp.AssertStatus(t, http.StatusOK)

	var related struct {
		Data []any `json:"data"`
	}
	relatedResp.JSON(t, &related)
	if related.Data == nil {
		t.Fatal("expected related products data array")
	}
}

func TestStorefront_Products_CategoryFilter(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)

	categoryID := createCategory(t, admin, map[string]any{
		"name":      fmt.Sprintf("E2E Store Category %d", testRunID),
		"is_active": true,
	})
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/categories/"+categoryID, nil, nil)
	})

	n := uniqueCounter.Add(1)
	name := fmt.Sprintf("E2E Categorized Product %d-%d", testRunID, n)

	createResp := admin.do(http.MethodPost, "/api/v1/admin/products", map[string]any{
		"name":        name,
		"price":       60000,
		"status":      "active",
		"category_id": categoryID,
		"inventory": map[string]any{
			"quantity": 8,
		},
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var product struct {
		ID string `json:"id"`
	}
	createResp.JSON(t, &product)
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/products/"+product.ID, nil, nil)
	})

	filterResp := store.do(http.MethodGet, "/api/v1/store/products?category_id="+categoryID+"&page=1&per_page=20", nil, nil)
	filterResp.AssertStatus(t, http.StatusOK)

	var filtered struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	filterResp.JSON(t, &filtered)

	found := false
	for _, p := range filtered.Data {
		if p.ID == product.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("product %s not found when filtering by category %s", product.ID, categoryID)
	}
}

func TestStorefront_Products_DraftHidden(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)

	n := uniqueCounter.Add(1)
	name := fmt.Sprintf("E2E Draft Product %d-%d", testRunID, n)
	slug := fmt.Sprintf("e2e-draft-product-%d-%d", testRunID, n)

	createResp := admin.do(http.MethodPost, "/api/v1/admin/products", map[string]any{
		"name":   name,
		"slug":   slug,
		"price":  40000,
		"status": "draft",
		"inventory": map[string]any{
			"quantity": 5,
		},
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var draft struct {
		ID string `json:"id"`
	}
	createResp.JSON(t, &draft)
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/products/"+draft.ID, nil, nil)
	})

	getResp := store.do(http.MethodGet, "/api/v1/store/products/"+slug, nil, nil)
	getResp.AssertStatus(t, http.StatusNotFound)
	getResp.AssertErrorCode(t, "NOT_FOUND")

	searchResp := store.do(http.MethodGet, "/api/v1/store/products/search?q="+url.QueryEscape(name), nil, nil)
	searchResp.AssertStatus(t, http.StatusOK)

	var search struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	searchResp.JSON(t, &search)
	for _, hit := range search.Data {
		if hit.ID == draft.ID {
			t.Fatalf("draft product %s should not appear in search", draft.ID)
		}
	}
}

func TestStorefront_Categories_ActiveOnly(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)

	activeID := createCategory(t, admin, map[string]any{
		"name":      fmt.Sprintf("E2E Active Store Category %d", testRunID),
		"is_active": true,
	})
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/categories/"+activeID, nil, nil)
	})

	inactiveID := createCategory(t, admin, map[string]any{
		"name": fmt.Sprintf("E2E Inactive Store Category %d", testRunID),
	})
	deactivateResp := admin.do(http.MethodPut, "/api/v1/admin/categories/"+inactiveID, map[string]any{
		"is_active": false,
	}, nil)
	deactivateResp.AssertStatus(t, http.StatusOK)

	var deactivated struct {
		IsActive bool `json:"is_active"`
	}
	deactivateResp.JSON(t, &deactivated)
	if deactivated.IsActive {
		t.Fatal("expected is_active=false after update")
	}
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/categories/"+inactiveID, nil, nil)
	})

	resp := store.do(http.MethodGet, "/api/v1/store/categories", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var tree struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(t, &tree)

	foundActive := false
	foundInactive := false
	for _, cat := range tree.Data {
		if cat.ID == activeID {
			foundActive = true
		}
		if cat.ID == inactiveID {
			foundInactive = true
		}
	}
	if !foundActive {
		t.Fatalf("active category %s not in storefront tree", activeID)
	}
	if foundInactive {
		t.Fatalf("inactive category %s should not appear on storefront", inactiveID)
	}
}

func TestStorefront_Brands(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)

	slug := fmt.Sprintf("e2e-store-brand-%d", testRunID)
	createResp := admin.do(http.MethodPost, "/api/v1/admin/brands", map[string]any{
		"name":      "E2E Storefront Brand",
		"slug":      slug,
		"is_active": true,
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var brand struct {
		ID string `json:"id"`
	}
	createResp.JSON(t, &brand)
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/brands/"+brand.ID, nil, nil)
	})

	listResp := store.do(http.MethodGet, "/api/v1/store/brands", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
			Name string `json:"name"`
		} `json:"data"`
	}
	listResp.JSON(t, &list)

	found := false
	for _, b := range list.Data {
		if b.ID == brand.ID {
			found = true
			if b.Slug != slug {
				t.Fatalf("slug = %q, want %q", b.Slug, slug)
			}
			break
		}
	}
	if !found {
		t.Fatalf("active brand %s not found in storefront list", brand.ID)
	}
}

func TestStorefront_Search_ValidationError(t *testing.T) {
	store := storeClient(t)
	resp := store.do(http.MethodGet, "/api/v1/store/products/search", nil, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestStorefront_Product_NotFound(t *testing.T) {
	store := storeClient(t)
	resp := store.do(http.MethodGet, "/api/v1/store/products/does-not-exist-slug", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}
