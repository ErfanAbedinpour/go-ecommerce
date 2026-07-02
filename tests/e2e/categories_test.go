//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func createCategory(t *testing.T, admin *client, body map[string]any) string {
	t.Helper()
	resp := admin.do(http.MethodPost, "/api/v1/admin/categories", body, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var cat struct {
		ID string `json:"id"`
	}
	resp.JSON(t, &cat)
	if cat.ID == "" {
		t.Fatal("expected category id")
	}
	return cat.ID
}

func TestCategories_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/categories", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestCategories_Create_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing name",
			body: map[string]any{
				"slug": "e2e-category",
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
				"name": fmt.Sprintf("%0201d", 0),
			},
		},
		{
			name: "invalid parent_id",
			body: map[string]any{
				"name":      "Child",
				"parent_id": "not-a-uuid",
			},
		},
		{
			name: "invalid image_url",
			body: map[string]any{
				"name":      "With Image",
				"image_url": "not-a-url",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPost, "/api/v1/admin/categories", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestCategories_CRUD_Lifecycle(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-category-%d", testRunID)

	createBody := map[string]any{
		"name":        "E2E Test Category",
		"slug":        slug,
		"description": "Category created by e2e test",
		"sort_order":  10,
		"is_active":   true,
	}

	createResp := c.do(http.MethodPost, "/api/v1/admin/categories", createBody, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
		IsActive    bool   `json:"is_active"`
	}
	createResp.JSON(t, &created)

	if created.ID == "" {
		t.Fatal("expected category id")
	}
	if created.Name != "E2E Test Category" {
		t.Fatalf("name = %q", created.Name)
	}
	if created.Slug != slug {
		t.Fatalf("slug = %q, want %q", created.Slug, slug)
	}
	if created.SortOrder != 10 {
		t.Fatalf("sort_order = %d, want 10", created.SortOrder)
	}
	if !created.IsActive {
		t.Fatal("expected is_active=true")
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/categories/"+created.ID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	listResp := c.do(http.MethodGet, "/api/v1/admin/categories?page=1&per_page=10&is_active=true", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Page int `json:"page"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Page != 1 {
		t.Fatalf("unexpected pagination meta: %+v", list.Meta)
	}
	if len(list.Data) == 0 {
		t.Fatal("expected at least one active category")
	}

	newSlug := fmt.Sprintf("e2e-category-updated-%d", testRunID)
	updateResp := c.do(http.MethodPut, "/api/v1/admin/categories/"+created.ID, map[string]any{
		"name":        "E2E Updated Category",
		"slug":        newSlug,
		"description": "Updated description",
		"is_active":   false,
		"sort_order":  20,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		Name      string `json:"name"`
		Slug      string `json:"slug"`
		IsActive  bool   `json:"is_active"`
		SortOrder int    `json:"sort_order"`
	}
	updateResp.JSON(t, &updated)
	if updated.Name != "E2E Updated Category" {
		t.Fatalf("updated name = %q", updated.Name)
	}
	if updated.Slug != newSlug {
		t.Fatalf("updated slug = %q", updated.Slug)
	}
	if updated.IsActive {
		t.Fatal("expected is_active=false after update")
	}
	if updated.SortOrder != 20 {
		t.Fatalf("sort_order = %d, want 20", updated.SortOrder)
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/categories/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := c.do(http.MethodGet, "/api/v1/admin/categories/"+created.ID, nil, nil)
	notFoundResp.AssertStatus(t, http.StatusNotFound)
	notFoundResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestCategories_ParentChild_AndTree(t *testing.T) {
	c := adminClient(t)

	parentID := createCategory(t, c, map[string]any{
		"name":      fmt.Sprintf("E2E Parent %d", testRunID),
		"is_active": true,
	})
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/categories/"+parentID, nil, nil)
	})

	childID := createCategory(t, c, map[string]any{
		"name":      fmt.Sprintf("E2E Child %d", testRunID),
		"parent_id": parentID,
		"is_active": true,
	})
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/categories/"+childID, nil, nil)
	})

	treeResp := c.do(http.MethodGet, "/api/v1/admin/categories?tree=true", nil, nil)
	treeResp.AssertStatus(t, http.StatusOK)

	var tree struct {
		Data []struct {
			ID       string `json:"id"`
			Children []struct {
				ID string `json:"id"`
			} `json:"children"`
		} `json:"data"`
	}
	treeResp.JSON(t, &tree)

	foundParent := false
	for _, node := range tree.Data {
		if node.ID != parentID {
			continue
		}
		foundParent = true
		foundChild := false
		for _, ch := range node.Children {
			if ch.ID == childID {
				foundChild = true
				break
			}
		}
		if !foundChild {
			t.Fatalf("expected child %s under parent %s in tree", childID, parentID)
		}
	}
	if !foundParent {
		t.Fatalf("parent %s not found in category tree", parentID)
	}

	deleteParentResp := c.do(http.MethodDelete, "/api/v1/admin/categories/"+parentID, nil, nil)
	deleteParentResp.AssertStatus(t, http.StatusUnprocessableEntity)
	deleteParentResp.AssertErrorCode(t, "UNPROCESSABLE_ENTITY")

	deleteChildResp := c.do(http.MethodDelete, "/api/v1/admin/categories/"+childID, nil, nil)
	deleteChildResp.AssertStatus(t, http.StatusNoContent)

	deleteParentAfterResp := c.do(http.MethodDelete, "/api/v1/admin/categories/"+parentID, nil, nil)
	deleteParentAfterResp.AssertStatus(t, http.StatusNoContent)
}

func TestCategories_SlugConflict(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-category-conflict-%d", testRunID)

	firstID := createCategory(t, c, map[string]any{
		"name": "First Category",
		"slug": slug,
	})
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/categories/"+firstID, nil, nil)
	})

	secondResp := c.do(http.MethodPost, "/api/v1/admin/categories", map[string]any{
		"name": "Second Category",
		"slug": slug,
	}, nil)
	secondResp.AssertStatus(t, http.StatusConflict)
	secondResp.AssertErrorCode(t, "CONFLICT")
}

func TestCategories_Delete_WithProducts(t *testing.T) {
	c := adminClient(t)

	categoryID := createCategory(t, c, map[string]any{
		"name":      fmt.Sprintf("E2E Category With Product %d", testRunID),
		"is_active": true,
	})

	productResp := c.do(http.MethodPost, "/api/v1/admin/products", map[string]any{
		"name":        fmt.Sprintf("E2E Product In Category %d", testRunID),
		"price":       50000,
		"status":      "active",
		"category_id": categoryID,
		"inventory": map[string]any{
			"quantity": 5,
		},
	}, nil)
	productResp.AssertStatus(t, http.StatusCreated)

	var product struct {
		ID string `json:"id"`
	}
	productResp.JSON(t, &product)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/products/"+product.ID, nil, nil)
		_ = c.do(http.MethodDelete, "/api/v1/admin/categories/"+categoryID, nil, nil)
	})

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/categories/"+categoryID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusUnprocessableEntity)
	deleteResp.AssertErrorCode(t, "UNPROCESSABLE_ENTITY")
}

func TestCategories_ParentNotFound(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodPost, "/api/v1/admin/categories", map[string]any{
		"name":      "Orphan Category",
		"parent_id": "00000000-0000-0000-0000-000000000099",
	}, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestCategories_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/categories", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestCategories_Get_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/categories/00000000-0000-0000-0000-000000000099", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}
