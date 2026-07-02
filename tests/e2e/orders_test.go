//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestOrders_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/orders", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestOrders_StorefrontLifecycle(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	email := uniqueEmail("order-life")
	orderID := placeStoreOrder(t, store, productID, email, "")

	c := adminClient(t)

	getResp := c.do(http.MethodGet, "/api/v1/admin/orders/"+orderID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	listResp := c.do(http.MethodGet, "/api/v1/admin/orders?status=pending&page=1&per_page=10", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Total == 0 {
		t.Fatal("expected at least one pending order")
	}

	notesResp := c.do(http.MethodPatch, "/api/v1/admin/orders/"+orderID+"/notes", map[string]string{
		"notes": "E2E internal order note",
	}, nil)
	notesResp.AssertStatus(t, http.StatusOK)

	var withNotes struct {
		Notes string `json:"notes"`
	}
	notesResp.JSON(t, &withNotes)
	if withNotes.Notes != "E2E internal order note" {
		t.Fatalf("notes = %q", withNotes.Notes)
	}

	invoiceResp := c.do(http.MethodGet, "/api/v1/admin/orders/"+orderID+"/invoice", nil, nil)
	invoiceResp.AssertStatus(t, http.StatusOK)

	var invoice struct {
		InvoiceNumber string `json:"invoice_number"`
		Order         struct {
			ID string `json:"id"`
		} `json:"order"`
	}
	invoiceResp.JSON(t, &invoice)
	if invoice.Order.ID != orderID || invoice.InvoiceNumber == "" {
		t.Fatalf("unexpected invoice: %+v", invoice)
	}

	callbackResp := store.do(http.MethodPost, "/api/v1/store/checkout/payment/callback", map[string]string{
		"order_id":  orderID,
		"authority": "e2e-payment-auth",
		"status":    "OK",
	}, nil)
	callbackResp.AssertStatus(t, http.StatusOK)

	statuses := []string{"processing", "shipped", "delivered"}
	for _, status := range statuses {
		resp := c.do(http.MethodPatch, "/api/v1/admin/orders/"+orderID+"/status", map[string]string{
			"status": status,
			"note":   "E2E status -> " + status,
		}, nil)
		resp.AssertStatus(t, http.StatusOK)

		var updated struct {
			Status string `json:"status"`
		}
		resp.JSON(t, &updated)
		if updated.Status != status {
			t.Fatalf("status = %q, want %q", updated.Status, status)
		}
	}

	refundResp := c.do(http.MethodPost, "/api/v1/admin/orders/"+orderID+"/refund", map[string]any{
		"amount": 50000,
		"reason": "E2E partial refund test",
	}, nil)
	refundResp.AssertStatus(t, http.StatusOK)

	var refunded struct {
		Status        string `json:"status"`
		PaymentStatus string `json:"payment_status"`
	}
	refundResp.JSON(t, &refunded)
	if refunded.Status != "refunded" {
		t.Fatalf("status = %q, want refunded", refunded.Status)
	}
}

func TestOrders_Cancel(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	orderID := placeStoreOrder(t, store, productID, uniqueEmail("order-cancel"), "")

	c := adminClient(t)
	cancelResp := c.do(http.MethodPost, "/api/v1/admin/orders/"+orderID+"/cancel", nil, nil)
	cancelResp.AssertStatus(t, http.StatusOK)

	var cancelled struct {
		Status string `json:"status"`
	}
	cancelResp.JSON(t, &cancelled)
	if cancelled.Status != "cancelled" {
		t.Fatalf("status = %q, want cancelled", cancelled.Status)
	}

	invalidTransitionResp := c.do(http.MethodPatch, "/api/v1/admin/orders/"+orderID+"/status", map[string]string{
		"status": "processing",
	}, nil)
	invalidTransitionResp.AssertStatus(t, http.StatusUnprocessableEntity)
	invalidTransitionResp.AssertErrorCode(t, "INVALID_STATUS_TRANSITION")
}

func TestOrders_Create_Manual(t *testing.T) {
	admin := adminClient(t)
	customer := customerStoreClient(t)
	productID := createCheckoutProduct(t, admin)

	profileResp := customer.do(http.MethodGet, "/api/v1/store/account/profile", nil, nil)
	profileResp.AssertStatus(t, http.StatusOK)

	var profile struct {
		ID string `json:"id"`
	}
	profileResp.JSON(t, &profile)

	createResp := admin.do(http.MethodPost, "/api/v1/admin/orders", map[string]any{
		"customer_id": profile.ID,
		"items": []map[string]any{
			{"product_id": productID, "quantity": 1},
		},
		"shipping_amount": 85000,
		"tax_amount":      0,
		"billing_address": standardShippingAddress(),
		"shipping_address": standardShippingAddress(),
		"payment_method":  "manual",
		"payment_status":  "paid",
		"notes":           "E2E manual admin order",
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		PaymentStatus string `json:"payment_status"`
		Items         []struct {
			ProductID string  `json:"product_id"`
			Quantity  int     `json:"quantity"`
			UnitPrice float64 `json:"unit_price"`
		} `json:"items"`
	}
	createResp.JSON(t, &created)
	if created.ID == "" {
		t.Fatal("expected order id")
	}
	if created.Status != "pending" {
		t.Fatalf("status = %q, want pending", created.Status)
	}
	if created.PaymentStatus != "paid" {
		t.Fatalf("payment_status = %q, want paid", created.PaymentStatus)
	}
	if len(created.Items) != 1 || created.Items[0].ProductID != productID {
		t.Fatalf("unexpected items: %+v", created.Items)
	}
}

func TestOrders_Create_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodPost, "/api/v1/admin/orders", map[string]any{
		"customer_id": "not-a-uuid",
		"items": []map[string]any{
			{"product_id": "00000000-0000-0000-0000-000000000001", "quantity": 1},
		},
		"billing_address":  standardShippingAddress(),
		"shipping_address": standardShippingAddress(),
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestOrders_UpdateStatus_InvalidTransition(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	orderID := placeStoreOrder(t, store, productID, uniqueEmail("order-bad-status"), "")

	c := adminClient(t)
	resp := c.do(http.MethodPatch, "/api/v1/admin/orders/"+orderID+"/status", map[string]string{
		"status": "delivered",
	}, nil)
	resp.AssertStatus(t, http.StatusUnprocessableEntity)
	resp.AssertErrorCode(t, "INVALID_STATUS_TRANSITION")
}

func TestOrders_Get_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/orders/00000000-0000-0000-0000-000000000099", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestOrders_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/orders", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestOrders_List_InvalidDateRange(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/orders?from=2026-02-01&to=2026-01-01", nil, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}
