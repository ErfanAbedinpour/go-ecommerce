//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestWishlist_Unauthorized(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodGet, "/api/v1/store/account/wishlist", nil},
		{http.MethodPost, "/api/v1/store/account/wishlist", map[string]string{"product_id": "00000000-0000-0000-0000-000000000001"}},
		{http.MethodGet, "/api/v1/store/account/wishlist/count", nil},
		{http.MethodGet, "/api/v1/store/account/wishlist/ids", nil},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			resp := testClient.do(ep.method, ep.path, ep.body, nil)
			resp.AssertStatus(t, http.StatusUnauthorized)
		})
	}
}

func TestWishlist_AdminForbidden(t *testing.T) {
	admin := adminClient(t)
	resp := admin.do(http.MethodGet, "/api/v1/store/account/wishlist", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestWishlist_Add_ValidationErrors(t *testing.T) {
	customer := customerClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing product_id",
			body: map[string]any{},
		},
		{
			name: "invalid uuid",
			body: map[string]any{"product_id": "not-a-uuid"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := customer.do(http.MethodPost, "/api/v1/store/account/wishlist", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestWishlist_Lifecycle(t *testing.T) {
	admin := adminClient(t)
	customer := customerClient(t)
	productID := createActiveProduct(t, admin)

	addResp := customer.do(http.MethodPost, "/api/v1/store/account/wishlist", map[string]string{
		"product_id": productID,
	}, nil)
	addResp.AssertStatus(t, http.StatusCreated)

	var added struct {
		ID        string `json:"id"`
		ProductID string `json:"product_id"`
	}
	addResp.JSON(t, &added)
	if added.ProductID != productID {
		t.Fatalf("product_id = %q, want %q", added.ProductID, productID)
	}

	dupResp := customer.do(http.MethodPost, "/api/v1/store/account/wishlist", map[string]string{
		"product_id": productID,
	}, nil)
	dupResp.AssertStatus(t, http.StatusOK)

	countResp := customer.do(http.MethodGet, "/api/v1/store/account/wishlist/count", nil, nil)
	countResp.AssertStatus(t, http.StatusOK)

	var count struct {
		Count int `json:"count"`
	}
	countResp.JSON(t, &count)
	if count.Count != 1 {
		t.Fatalf("count = %d, want 1", count.Count)
	}

	idsResp := customer.do(http.MethodGet, "/api/v1/store/account/wishlist/ids", nil, nil)
	idsResp.AssertStatus(t, http.StatusOK)

	var ids struct {
		ProductIDs []string `json:"product_ids"`
	}
	idsResp.JSON(t, &ids)
	if len(ids.ProductIDs) != 1 || ids.ProductIDs[0] != productID {
		t.Fatalf("product_ids = %v, want [%s]", ids.ProductIDs, productID)
	}

	listResp := customer.do(http.MethodGet, "/api/v1/store/account/wishlist?page=1&per_page=10", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ProductID string `json:"product_id"`
			Product   struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"product"`
		} `json:"data"`
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Total != 1 {
		t.Fatalf("total = %d, want 1", list.Meta.Total)
	}
	if len(list.Data) != 1 {
		t.Fatalf("expected 1 wishlist item, got %d", len(list.Data))
	}
	if list.Data[0].ProductID != productID {
		t.Fatalf("product_id = %q, want %q", list.Data[0].ProductID, productID)
	}
	if list.Data[0].Product.Name == "" || list.Data[0].Product.Slug == "" {
		t.Fatalf("expected product summary fields: %+v", list.Data[0].Product)
	}

	removeResp := customer.do(http.MethodDelete, "/api/v1/store/account/wishlist/"+productID, nil, nil)
	removeResp.AssertStatus(t, http.StatusNoContent)

	emptyCountResp := customer.do(http.MethodGet, "/api/v1/store/account/wishlist/count", nil, nil)
	emptyCountResp.AssertStatus(t, http.StatusOK)

	var emptyCount struct {
		Count int `json:"count"`
	}
	emptyCountResp.JSON(t, &emptyCount)
	if emptyCount.Count != 0 {
		t.Fatalf("count after remove = %d, want 0", emptyCount.Count)
	}
}

func TestWishlist_Add_ProductNotFound(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodPost, "/api/v1/store/account/wishlist", map[string]string{
		"product_id": "00000000-0000-0000-0000-000000000099",
	}, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestWishlist_Remove_NotInList(t *testing.T) {
	admin := adminClient(t)
	customer := customerClient(t)
	productID := createActiveProduct(t, admin)

	resp := customer.do(http.MethodDelete, "/api/v1/store/account/wishlist/"+productID, nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}
