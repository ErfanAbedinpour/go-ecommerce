//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestCheckout_Preview_EmptyCart(t *testing.T) {
	store := storeClient(t)

	resp := store.do(http.MethodPost, "/api/v1/store/checkout/preview", map[string]any{
		"shipping_method": "post",
		"shipping_city":   "Tehran",
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestCheckout_Preview_ValidationErrors(t *testing.T) {
	store := storeClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing shipping method",
			body: map[string]any{"shipping_city": "Tehran"},
		},
		{
			name: "invalid shipping method",
			body: map[string]any{
				"shipping_method": "drone",
				"shipping_city":   "Tehran",
			},
		},
		{
			name: "missing shipping city",
			body: map[string]any{"shipping_method": "post"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := store.do(http.MethodPost, "/api/v1/store/checkout/preview", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestCheckout_Preview_Success(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	addProductToCart(t, store, productID, 1)

	resp := store.do(http.MethodPost, "/api/v1/store/checkout/preview", map[string]any{
		"shipping_method": "post",
		"shipping_city":   "Tehran",
	}, nil)
	resp.AssertStatus(t, http.StatusOK)

	var preview struct {
		Items []struct {
			ProductID      string `json:"product_id"`
			Quantity       int    `json:"quantity"`
			IsAvailable    bool   `json:"is_available"`
			LineTotalToman int64  `json:"line_total_toman"`
		} `json:"items"`
		Summary struct {
			SubtotalToman int64 `json:"subtotal_toman"`
			ShippingToman int64 `json:"shipping_toman"`
			TotalToman    int64 `json:"total_toman"`
			Currency      string `json:"currency"`
		} `json:"summary"`
	}
	resp.JSON(t, &preview)
	if len(preview.Items) != 1 {
		t.Fatalf("expected 1 preview item, got %d", len(preview.Items))
	}
	if preview.Items[0].ProductID != productID || !preview.Items[0].IsAvailable {
		t.Fatalf("unexpected preview item: %+v", preview.Items[0])
	}
	if preview.Summary.SubtotalToman <= 0 {
		t.Fatal("expected positive subtotal")
	}
	if preview.Summary.ShippingToman <= 0 {
		t.Fatal("expected positive shipping")
	}
	if preview.Summary.TotalToman < preview.Summary.SubtotalToman {
		t.Fatal("expected total >= subtotal")
	}
	if preview.Summary.Currency != "IRT" {
		t.Fatalf("currency = %q, want IRT", preview.Summary.Currency)
	}
}

func TestCheckout_PlaceOrder_Guest(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	email := uniqueEmail("guest-checkout")

	orderID := placeStoreOrder(t, store, productID, email, "")

	emptyCartResp := store.do(http.MethodGet, "/api/v1/store/cart", nil, nil)
	emptyCartResp.AssertStatus(t, http.StatusOK)

	var cart struct {
		Items []any `json:"items"`
	}
	emptyCartResp.JSON(t, &cart)
	if len(cart.Items) != 0 {
		t.Fatal("expected cart cleared after checkout")
	}

	c := adminClient(t)
	getResp := c.do(http.MethodGet, "/api/v1/admin/orders/"+orderID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var order struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		PaymentStatus string `json:"payment_status"`
		Items         []struct {
			ProductID string `json:"product_id"`
		} `json:"items"`
	}
	getResp.JSON(t, &order)
	if order.ID != orderID {
		t.Fatalf("order id = %q", order.ID)
	}
	if order.Status != "pending" {
		t.Fatalf("status = %q, want pending", order.Status)
	}
	if len(order.Items) != 1 || order.Items[0].ProductID != productID {
		t.Fatalf("unexpected order items: %+v", order.Items)
	}
}

func TestCheckout_PlaceOrder_AuthenticatedCustomer(t *testing.T) {
	admin := adminClient(t)
	store := customerStoreClient(t)
	productID := createCheckoutProduct(t, admin)

	profileResp := store.do(http.MethodGet, "/api/v1/store/account/profile", nil, nil)
	profileResp.AssertStatus(t, http.StatusOK)

	var profile struct {
		Email string `json:"email"`
	}
	profileResp.JSON(t, &profile)

	orderID := placeStoreOrder(t, store, productID, profile.Email, "")

	ordersResp := store.do(http.MethodGet, "/api/v1/store/account/orders?page=1&per_page=5", nil, nil)
	ordersResp.AssertStatus(t, http.StatusOK)

	var orders struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	ordersResp.JSON(t, &orders)
	if len(orders.Data) == 0 {
		t.Fatal("expected customer order history")
	}

	found := false
	for _, o := range orders.Data {
		if o.ID == orderID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("order %s not found in account orders", orderID)
	}
}

func TestCheckout_Preview_WithCoupon(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	code := uniqueCouponCode("E2E")

	couponResp := admin.do(http.MethodPost, "/api/v1/admin/coupons", map[string]any{
		"code":            code,
		"discount_type":   "percentage",
		"discount_value":  10,
		"min_order_amount": 0,
		"is_active":       true,
	}, nil)
	couponResp.AssertStatus(t, http.StatusCreated)

	var coupon struct {
		ID string `json:"id"`
	}
	couponResp.JSON(t, &coupon)
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/coupons/"+coupon.ID, nil, nil)
	})

	addProductToCart(t, store, productID, 1)

	previewResp := store.do(http.MethodPost, "/api/v1/store/checkout/preview", map[string]any{
		"shipping_method": "post",
		"shipping_city":   "Tehran",
		"coupon_code":     code,
	}, nil)
	previewResp.AssertStatus(t, http.StatusOK)

	var preview struct {
		Summary struct {
			DiscountToman int64 `json:"discount_toman"`
			TotalToman    int64 `json:"total_toman"`
		} `json:"summary"`
		Coupon struct {
			Code    string `json:"code"`
			IsValid bool   `json:"is_valid"`
		} `json:"coupon"`
	}
	previewResp.JSON(t, &preview)
	if !preview.Coupon.IsValid {
		t.Fatal("expected valid coupon in preview")
	}
	if preview.Coupon.Code != code {
		t.Fatalf("coupon code = %q, want %q", preview.Coupon.Code, code)
	}
	if preview.Summary.DiscountToman <= 0 {
		t.Fatal("expected discount applied")
	}
}

func TestCheckout_Preview_InvalidCoupon(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	addProductToCart(t, store, productID, 1)

	resp := store.do(http.MethodPost, "/api/v1/store/checkout/preview", map[string]any{
		"shipping_method": "post",
		"shipping_city":   "Tehran",
		"coupon_code":     "DOESNOTEXIST",
	}, nil)
	resp.AssertStatus(t, http.StatusOK)

	var preview struct {
		Coupon *struct {
			IsValid bool   `json:"is_valid"`
			Message string `json:"message"`
		} `json:"coupon"`
	}
	resp.JSON(t, &preview)
	if preview.Coupon == nil || preview.Coupon.IsValid {
		t.Fatal("expected invalid coupon result")
	}
}

func TestCheckout_PlaceOrder_ValidationErrors(t *testing.T) {
	store := storeClient(t)

	resp := store.do(http.MethodPost, "/api/v1/store/checkout", map[string]any{
		"shipping_method": "post",
		"shipping_city":   "Tehran",
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestCheckout_PaymentCallback(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	email := uniqueEmail("pay-callback")
	orderID := placeStoreOrder(t, store, productID, email, "")

	callbackResp := store.do(http.MethodPost, "/api/v1/store/checkout/payment/callback", map[string]string{
		"order_id":  orderID,
		"authority": "test-authority-123",
		"status":    "OK",
	}, nil)
	callbackResp.AssertStatus(t, http.StatusOK)

	var callback struct {
		OrderID       string `json:"order_id"`
		PaymentStatus string `json:"payment_status"`
	}
	callbackResp.JSON(t, &callback)
	if callback.OrderID != orderID {
		t.Fatalf("order_id = %q", callback.OrderID)
	}
	if callback.PaymentStatus != "paid" {
		t.Fatalf("payment_status = %q, want paid", callback.PaymentStatus)
	}
}

func TestCheckout_ValidateCustomer_BlocksRegisteredEmail(t *testing.T) {
	registered := customerStoreClient(t)

	profileResp := registered.do(http.MethodGet, "/api/v1/store/account/profile", nil, nil)
	profileResp.AssertStatus(t, http.StatusOK)

	var profile struct {
		Email string `json:"email"`
	}
	profileResp.JSON(t, &profile)

	guest := storeClient(t)
	resp := guest.do(http.MethodPost, "/api/v1/store/checkout/validate-customer", map[string]string{
		"email": profile.Email,
		"phone": "+989121234567",
	}, nil)
	resp.AssertStatus(t, http.StatusConflict)
	resp.AssertErrorCode(t, "ACCOUNT_EXISTS_LOGIN_REQUIRED")
}

func TestCheckout_GuestOrderLinkedAfterSignup(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createCheckoutProduct(t, admin)
	email := uniqueEmail("guest-link")

	orderID := placeStoreOrder(t, store, productID, email, "")

	signupResp := store.do(http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":      email,
		"password":   "CustomerPass1!",
		"first_name": "Linked",
		"last_name":  "Guest",
		"phone":      "+989121234567",
	}, nil)
	signupResp.AssertStatus(t, http.StatusCreated)

	var tokens tokenResponse
	signupResp.JSON(t, &tokens)
	store = store.withToken(tokens.AccessToken)

	ordersResp := store.do(http.MethodGet, "/api/v1/store/account/orders?page=1&per_page=10", nil, nil)
	ordersResp.AssertStatus(t, http.StatusOK)

	var orders struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	ordersResp.JSON(t, &orders)

	found := false
	for _, o := range orders.Data {
		if o.ID == orderID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("guest order %s not linked after signup", orderID)
	}
}