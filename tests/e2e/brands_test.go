//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestBrands_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/brands", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestBrands_Create_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing name",
			body: map[string]any{
				"slug": "e2e-brand",
			},
		},
		{
			name: "empty name",
			body: map[string]any{
				"name": "",
			},
		},
		{
			name: "name too long",
			body: map[string]any{
				"name": fmt.Sprintf("%0101d", 0),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPost, "/api/v1/admin/brands", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestBrands_CRUD_Lifecycle(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-brand-%d", testRunID)

	createBody := map[string]any{
		"name":        "E2E Test Brand",
		"slug":        slug,
		"description": "Brand created by e2e test",
		"is_active":   true,
	}

	createResp := c.do(http.MethodPost, "/api/v1/admin/brands", createBody, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}
	createResp.JSON(t, &created)

	if created.ID == "" {
		t.Fatal("expected brand id")
	}
	if created.Name != "E2E Test Brand" {
		t.Fatalf("name = %q", created.Name)
	}
	if created.Slug != slug {
		t.Fatalf("slug = %q, want %q", created.Slug, slug)
	}
	if created.Description != "Brand created by e2e test" {
		t.Fatalf("description = %q", created.Description)
	}
	if !created.IsActive {
		t.Fatal("expected is_active=true")
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/brands/"+created.ID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var fetched struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	getResp.JSON(t, &fetched)
	if fetched.ID != created.ID || fetched.Name != created.Name {
		t.Fatalf("get mismatch: %+v", fetched)
	}

	listResp := c.do(http.MethodGet, "/api/v1/admin/brands?page=1&per_page=10&is_active=true", nil, nil)
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
	if list.Meta.Page != 1 {
		t.Fatalf("unexpected pagination meta: %+v", list.Meta)
	}
	if len(list.Data) == 0 {
		t.Fatal("expected at least one active brand")
	}

	newSlug := fmt.Sprintf("e2e-brand-updated-%d", testRunID)
	updateResp := c.do(http.MethodPut, "/api/v1/admin/brands/"+created.ID, map[string]any{
		"name":        "E2E Updated Brand",
		"slug":        newSlug,
		"description": "Updated description",
		"is_active":   false,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}
	updateResp.JSON(t, &updated)
	if updated.Name != "E2E Updated Brand" {
		t.Fatalf("updated name = %q", updated.Name)
	}
	if updated.Slug != newSlug {
		t.Fatalf("updated slug = %q", updated.Slug)
	}
	if updated.IsActive {
		t.Fatal("expected is_active=false after update")
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/brands/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := c.do(http.MethodGet, "/api/v1/admin/brands/"+created.ID, nil, nil)
	notFoundResp.AssertStatus(t, http.StatusNotFound)
	notFoundResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestBrands_SlugConflict(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-brand-conflict-%d", testRunID)

	firstResp := c.do(http.MethodPost, "/api/v1/admin/brands", map[string]any{
		"name": "First Brand",
		"slug": slug,
	}, nil)
	firstResp.AssertStatus(t, http.StatusCreated)

	var first struct {
		ID string `json:"id"`
	}
	firstResp.JSON(t, &first)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/brands/"+first.ID, nil, nil)
	})

	secondResp := c.do(http.MethodPost, "/api/v1/admin/brands", map[string]any{
		"name": "Second Brand",
		"slug": slug,
	}, nil)
	secondResp.AssertStatus(t, http.StatusConflict)
	secondResp.AssertErrorCode(t, "CONFLICT")
}

func TestBrands_Create_MinimalPayload(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodPost, "/api/v1/admin/brands", map[string]any{
		"name": fmt.Sprintf("Minimal Brand %d", testRunID),
	}, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var brand struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Slug   string `json:"slug"`
		IsActive bool `json:"is_active"`
	}
	resp.JSON(t, &brand)

	if brand.ID == "" || brand.Slug == "" {
		t.Fatalf("expected generated id and slug: %+v", brand)
	}
	if brand.IsActive {
		t.Fatal("expected default is_active=false when omitted")
	}

	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/brands/"+brand.ID, nil, nil)
	})
}

func TestBrands_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/brands", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestBrands_Get_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/brands/00000000-0000-0000-0000-000000000099", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}
