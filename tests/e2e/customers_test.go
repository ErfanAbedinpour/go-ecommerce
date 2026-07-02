//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func customerIDFromProfile(t *testing.T, store *client) (id, email string) {
	t.Helper()
	resp := store.do(http.MethodGet, "/api/v1/store/account/profile", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var profile struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	resp.JSON(t, &profile)
	if profile.ID == "" || profile.Email == "" {
		t.Fatalf("expected profile id and email: %+v", profile)
	}
	return profile.ID, profile.Email
}

func TestCustomers_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/customers", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestCustomers_List_AndSearch(t *testing.T) {
	store := customerStoreClient(t)
	_, email := customerIDFromProfile(t, store)

	c := adminClient(t)
	listResp := c.do(http.MethodGet, "/api/v1/admin/customers?page=1&per_page=20&type=registered", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Type  string `json:"type"`
		} `json:"data"`
		Meta struct {
			Page  int   `json:"page"`
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Page != 1 {
		t.Fatalf("page = %d, want 1", list.Meta.Page)
	}
	if list.Meta.Total == 0 {
		t.Fatal("expected at least one registered customer")
	}

	searchResp := c.do(http.MethodGet, "/api/v1/admin/customers?q="+email, nil, nil)
	searchResp.AssertStatus(t, http.StatusOK)

	var search struct {
		Data []struct {
			Email string `json:"email"`
		} `json:"data"`
	}
	searchResp.JSON(t, &search)
	if len(search.Data) == 0 {
		t.Fatalf("expected search hit for email %q", email)
	}
	found := false
	for _, row := range search.Data {
		if row.Email == email {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("search results do not include %q", email)
	}
}

func TestCustomers_Get_Update_Delete(t *testing.T) {
	store := customerStoreClient(t)
	customerID, email := customerIDFromProfile(t, store)

	c := adminClient(t)

	getResp := c.do(http.MethodGet, "/api/v1/admin/customers/"+customerID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var detail struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Type      string `json:"type"`
		Stats     struct {
			TotalOrders int `json:"total_orders"`
		} `json:"stats"`
		Addresses []any `json:"addresses"`
	}
	getResp.JSON(t, &detail)
	if detail.ID != customerID {
		t.Fatalf("id = %q, want %q", detail.ID, customerID)
	}
	if detail.Email != email {
		t.Fatalf("email = %q, want %q", detail.Email, email)
	}
	if detail.Type != "registered" {
		t.Fatalf("type = %q, want registered", detail.Type)
	}
	if detail.Stats.TotalOrders != 0 {
		t.Fatalf("total_orders = %d, want 0", detail.Stats.TotalOrders)
	}

	newPhone := "+989131112233"
	updateResp := c.do(http.MethodPut, "/api/v1/admin/customers/"+customerID, map[string]any{
		"first_name": "Updated",
		"last_name":  "Customer",
		"phone":      newPhone,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		FullName  string `json:"full_name"`
		Phone     string `json:"phone"`
	}
	updateResp.JSON(t, &updated)
	if updated.FirstName != "Updated" || updated.LastName != "Customer" {
		t.Fatalf("name mismatch: %+v", updated)
	}
	if updated.FullName != "Updated Customer" {
		t.Fatalf("full_name = %q, want Updated Customer", updated.FullName)
	}
	if updated.Phone != newPhone {
		t.Fatalf("phone = %q, want %q", updated.Phone, newPhone)
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/customers/"+customerID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := c.do(http.MethodGet, "/api/v1/admin/customers/"+customerID, nil, nil)
	notFoundResp.AssertStatus(t, http.StatusNotFound)
	notFoundResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestCustomers_Update_ValidationErrors(t *testing.T) {
	store := customerStoreClient(t)
	customerID, _ := customerIDFromProfile(t, store)

	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "invalid email",
			body: map[string]any{"email": "not-an-email"},
		},
		{
			name: "invalid type",
			body: map[string]any{"type": "invalid"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPut, "/api/v1/admin/customers/"+customerID, tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestCustomers_Update_DuplicateEmail(t *testing.T) {
	first := customerStoreClient(t)
	_, firstEmail := customerIDFromProfile(t, first)

	second := customerStoreClient(t)
	secondID, _ := customerIDFromProfile(t, second)

	c := adminClient(t)
	resp := c.do(http.MethodPut, "/api/v1/admin/customers/"+secondID, map[string]any{
		"email": firstEmail,
	}, nil)
	resp.AssertStatus(t, http.StatusConflict)
	resp.AssertErrorCode(t, "CONFLICT")
}

func TestCustomers_OrdersHistory(t *testing.T) {

	admin := adminClient(t)
	store := customerStoreClient(t)
	customerID, email := customerIDFromProfile(t, store)
	productID := createCheckoutProduct(t, admin)
	orderID := placeStoreOrder(t, store, productID, email, "")

	c := adminClient(t)
	ordersResp := c.do(http.MethodGet, "/api/v1/admin/customers/"+customerID+"/orders?page=1&per_page=10", nil, nil)
	ordersResp.AssertStatus(t, http.StatusOK)

	var orders struct {
		Data []struct {
			ID          string  `json:"id"`
			OrderNumber string  `json:"order_number"`
			Status      string  `json:"status"`
			Total       float64 `json:"total"`
			ItemCount   int     `json:"item_count"`
		} `json:"data"`
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	ordersResp.JSON(t, &orders)
	if orders.Meta.Total == 0 {
		t.Fatal("expected customer order history")
	}

	found := false
	for _, o := range orders.Data {
		if o.ID == orderID {
			found = true
			if o.OrderNumber == "" {
				t.Fatal("expected order_number")
			}
			if o.Status != "pending" {
				t.Fatalf("status = %q, want pending", o.Status)
			}
			if o.ItemCount != 1 {
				t.Fatalf("item_count = %d, want 1", o.ItemCount)
			}
			break
		}
	}
	if !found {
		t.Fatalf("order %s not found in customer history", orderID)
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/customers/"+customerID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var detail struct {
		TotalOrders int     `json:"total_orders"`
		TotalSpent  float64 `json:"total_spent"`
		Stats       struct {
			TotalOrders int `json:"total_orders"`
		} `json:"stats"`
	}
	getResp.JSON(t, &detail)
	if detail.TotalOrders < 1 || detail.Stats.TotalOrders < 1 {
		t.Fatalf("expected updated order stats: %+v", detail)
	}
}

func TestCustomers_Delete_WithOrders(t *testing.T) {

	admin := adminClient(t)
	store := customerStoreClient(t)
	customerID, email := customerIDFromProfile(t, store)
	productID := createCheckoutProduct(t, admin)
	_ = placeStoreOrder(t, store, productID, email, "")

	c := adminClient(t)
	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/customers/"+customerID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusUnprocessableEntity)
	deleteResp.AssertErrorCode(t, "UNPROCESSABLE_ENTITY")
}

func TestCustomers_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/customers", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestCustomers_Get_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/customers/00000000-0000-0000-0000-000000000099", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}
