//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestCart_EmptyGet(t *testing.T) {
	store := storeClient(t)

	resp := store.do(http.MethodGet, "/api/v1/store/cart", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var cart struct {
		Items []any `json:"items"`
		Summary struct {
			SubtotalToman int64  `json:"subtotal_toman"`
			Currency      string `json:"currency"`
		} `json:"summary"`
	}
	resp.JSON(t, &cart)
	if len(cart.Items) != 0 {
		t.Fatalf("expected empty cart, got %d items", len(cart.Items))
	}
	if cart.Summary.Currency != "IRT" {
		t.Fatalf("currency = %q, want IRT", cart.Summary.Currency)
	}
}

func TestCart_Add_ValidationErrors(t *testing.T) {
	store := storeClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing product_id",
			body: map[string]any{"quantity": 1},
		},
		{
			name: "invalid product_id",
			body: map[string]any{"product_id": "not-a-uuid", "quantity": 1},
		},
		{
			name: "zero quantity",
			body: map[string]any{"product_id": "00000000-0000-0000-0000-000000000001", "quantity": 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := store.do(http.MethodPost, "/api/v1/store/cart/items", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestCart_Lifecycle(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)

	addResp := store.do(http.MethodPost, "/api/v1/store/cart/items", map[string]any{
		"product_id": productID,
		"quantity":   2,
	}, nil)
	addResp.AssertStatus(t, http.StatusOK)

	var added struct {
		Items []struct {
			ProductID      string `json:"product_id"`
			Quantity       int    `json:"quantity"`
			UnitPriceToman int64  `json:"unit_price_toman"`
			LineTotalToman int64  `json:"line_total_toman"`
			IsAvailable    bool   `json:"is_available"`
		} `json:"items"`
		Summary struct {
			SubtotalToman int64 `json:"subtotal_toman"`
			TotalToman    int64 `json:"total_toman"`
		} `json:"summary"`
	}
	addResp.JSON(t, &added)
	if len(added.Items) != 1 {
		t.Fatalf("expected 1 cart item, got %d", len(added.Items))
	}
	if added.Items[0].ProductID != productID {
		t.Fatalf("product_id = %q", added.Items[0].ProductID)
	}
	if added.Items[0].Quantity != 2 {
		t.Fatalf("quantity = %d, want 2", added.Items[0].Quantity)
	}
	if !added.Items[0].IsAvailable {
		t.Fatal("expected item to be available")
	}
	if added.Items[0].LineTotalToman != added.Items[0].UnitPriceToman*2 {
		t.Fatalf("line total mismatch: unit=%d line=%d", added.Items[0].UnitPriceToman, added.Items[0].LineTotalToman)
	}
	if added.Summary.SubtotalToman != added.Items[0].LineTotalToman {
		t.Fatalf("subtotal = %d, want %d", added.Summary.SubtotalToman, added.Items[0].LineTotalToman)
	}

	updateResp := store.do(http.MethodPatch, "/api/v1/store/cart/items/"+productID, map[string]any{
		"quantity": 3,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		Items []struct {
			Quantity int `json:"quantity"`
		} `json:"items"`
	}
	updateResp.JSON(t, &updated)
	if len(updated.Items) != 1 || updated.Items[0].Quantity != 3 {
		t.Fatalf("expected quantity 3 after update: %+v", updated.Items)
	}

	dupAddResp := store.do(http.MethodPost, "/api/v1/store/cart/items", map[string]any{
		"product_id": productID,
		"quantity":   1,
	}, nil)
	dupAddResp.AssertStatus(t, http.StatusOK)

	var merged struct {
		Items []struct {
			Quantity int `json:"quantity"`
		} `json:"items"`
	}
	dupAddResp.JSON(t, &merged)
	if len(merged.Items) != 1 || merged.Items[0].Quantity != 4 {
		t.Fatalf("expected merged quantity 4: %+v", merged.Items)
	}

	removeResp := store.do(http.MethodDelete, "/api/v1/store/cart/items/"+productID, nil, nil)
	removeResp.AssertStatus(t, http.StatusOK)

	var removed struct {
		Items []any `json:"items"`
	}
	removeResp.JSON(t, &removed)
	if len(removed.Items) != 0 {
		t.Fatalf("expected empty cart after remove, got %d items", len(removed.Items))
	}

	addProductToCart(t, store, productID, 1)
	clearResp := store.do(http.MethodDelete, "/api/v1/store/cart", nil, nil)
	clearResp.AssertStatus(t, http.StatusNoContent)

	emptyResp := store.do(http.MethodGet, "/api/v1/store/cart", nil, nil)
	emptyResp.AssertStatus(t, http.StatusOK)

	var empty struct {
		Items []any `json:"items"`
	}
	emptyResp.JSON(t, &empty)
	if len(empty.Items) != 0 {
		t.Fatal("expected empty cart after clear")
	}
}

func TestCart_ProductNotFound(t *testing.T) {
	store := storeClient(t)

	resp := store.do(http.MethodPost, "/api/v1/store/cart/items", map[string]any{
		"product_id": "00000000-0000-0000-0000-000000000099",
		"quantity":   1,
	}, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestCart_Remove_NotInCart(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)

	resp := store.do(http.MethodDelete, "/api/v1/store/cart/items/"+productID, nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestCart_DefaultQuantityIsOne(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)

	resp := store.do(http.MethodPost, "/api/v1/store/cart/items", map[string]any{
		"product_id": productID,
	}, nil)
	resp.AssertStatus(t, http.StatusOK)

	var cart struct {
		Items []struct {
			Quantity int `json:"quantity"`
		} `json:"items"`
	}
	resp.JSON(t, &cart)
	if len(cart.Items) != 1 || cart.Items[0].Quantity != 1 {
		t.Fatalf("expected default quantity 1: %+v", cart.Items)
	}
}

func TestCart_AuthenticatedCustomer(t *testing.T) {
	admin := adminClient(t)
	store := customerStoreClient(t)
	productID := createCheckoutProduct(t, admin)

	addProductToCart(t, store, productID, 1)

	resp := store.do(http.MethodGet, "/api/v1/store/cart", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var cart struct {
		Items []struct {
			ProductID string `json:"product_id"`
		} `json:"items"`
	}
	resp.JSON(t, &cart)
	if len(cart.Items) != 1 || cart.Items[0].ProductID != productID {
		t.Fatalf("unexpected authenticated cart: %+v", cart.Items)
	}
}

func TestCart_MergesGuestCartAfterSignup(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	addProductToCart(t, store, productID, 2)

	email := uniqueEmail("cart-merge")
	signupResp := store.do(http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":      email,
		"password":   "CustomerPass1!",
		"first_name": "Merge",
		"last_name":  "Guest",
	}, nil)
	signupResp.AssertStatus(t, http.StatusCreated)

	var tokens tokenResponse
	signupResp.JSON(t, &tokens)
	store = store.withToken(tokens.AccessToken)

	resp := store.do(http.MethodGet, "/api/v1/store/cart", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var cart struct {
		Items []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}
	resp.JSON(t, &cart)
	if len(cart.Items) != 1 {
		t.Fatalf("expected merged cart with 1 item, got %d", len(cart.Items))
	}
	if cart.Items[0].ProductID != productID || cart.Items[0].Quantity != 2 {
		t.Fatalf("unexpected merged cart: %+v", cart.Items)
	}
}
