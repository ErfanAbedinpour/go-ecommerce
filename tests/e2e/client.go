//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"sync/atomic"
	"testing"
)

const (
	adminEmail    = "admin@shop.com"
	adminPassword = "Admin@123456"
)

type apiResponse struct {
	Status int
	Body   []byte
	Header http.Header
}

func (r apiResponse) JSON(t *testing.T, dst any) {
	t.Helper()
	if err := json.Unmarshal(r.Body, dst); err != nil {
		t.Fatalf("decode json: %v\nbody: %s", err, string(r.Body))
	}
}

func (r apiResponse) AssertStatus(t *testing.T, want int) {
	t.Helper()
	if r.Status != want {
		t.Fatalf("status = %d, want %d\nbody: %s", r.Status, want, string(r.Body))
	}
}

func (r apiResponse) AssertErrorCode(t *testing.T, want string) {
	t.Helper()
	var errBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	r.JSON(t, &errBody)
	if errBody.Error.Code != want {
		t.Fatalf("error code = %q, want %q\nbody: %s", errBody.Error.Code, want, string(r.Body))
	}
}

type client struct {
	baseURL string
	token   string
	http    *http.Client
}

func newClient(baseURL string) *client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(fmt.Sprintf("cookie jar: %v", err))
	}
	return &client{
		baseURL: baseURL,
		http:    &http.Client{Jar: jar},
	}
}

func (c *client) withToken(token string) *client {
	clone := *c
	clone.token = token
	return &clone
}

func (c *client) do(method, path string, body any, extraHeaders map[string]string) apiResponse {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("marshal body: %v", err))
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		panic(fmt.Sprintf("new request: %v", err))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		panic(fmt.Sprintf("%s %s: %v", method, path, err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("read body: %v", err))
	}

	return apiResponse{
		Status: resp.StatusCode,
		Body:   respBody,
		Header: resp.Header,
	}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func (c *client) loginAdmin(t *testing.T) string {
	t.Helper()
	return c.login(t, adminEmail, adminPassword)
}

func (c *client) login(t *testing.T, email, password string) string {
	t.Helper()
	resp := c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, nil)
	resp.AssertStatus(t, http.StatusOK)

	var tokens tokenResponse
	resp.JSON(t, &tokens)
	if tokens.AccessToken == "" {
		t.Fatal("expected access_token in login response")
	}
	if tokens.TokenType != "Bearer" {
		t.Fatalf("token_type = %q, want Bearer", tokens.TokenType)
	}
	if tokens.ExpiresIn <= 0 {
		t.Fatalf("expires_in = %d, want > 0", tokens.ExpiresIn)
	}
	return tokens.AccessToken
}

var uniqueCounter atomic.Int64

func uniqueEmail(prefix string) string {
	n := uniqueCounter.Add(1)
	return fmt.Sprintf("%s-%d-%d@test.local", prefix, testRunID, n)
}

func customerClient(t *testing.T) *client {
	t.Helper()
	return customerStoreClient(t)
}

func customerStoreClient(t *testing.T) *client {
	t.Helper()
	c := newClient(testServer.URL)
	email := uniqueEmail("customer")
	resp := c.do(http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":      email,
		"password":   "CustomerPass1!",
		"first_name": "E2E",
		"last_name":  "Customer",
	}, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var tokens tokenResponse
	resp.JSON(t, &tokens)
	return c.withToken(tokens.AccessToken)
}

func storeClient(t *testing.T) *client {
	t.Helper()
	return newClient(testServer.URL)
}

func standardShippingAddress() map[string]string {
	return map[string]string{
		"street":      "123 E2E Street",
		"city":        "Tehran",
		"postal_code": "1234567890",
		"country":     "IR",
	}
}

func checkoutCustomer(email string) map[string]string {
	return map[string]string{
		"email":      email,
		"first_name": "E2E",
		"last_name":  "Buyer",
		"phone":      "+989121234567",
	}
}

func createCheckoutProduct(t *testing.T, admin *client) string {
	t.Helper()
	n := uniqueCounter.Add(1)
	name := fmt.Sprintf("E2E Checkout Product %d-%d", testRunID, n)
	resp := admin.do(http.MethodPost, "/api/v1/admin/products", map[string]any{
		"name":   name,
		"price":  150000,
		"status": "active",
		"inventory": map[string]any{
			"quantity": 20,
		},
	}, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var product struct {
		ID string `json:"id"`
	}
	resp.JSON(t, &product)
	if product.ID == "" {
		t.Fatal("expected product id")
	}

	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/products/"+product.ID, nil, nil)
	})
	return product.ID
}

func addProductToCart(t *testing.T, store *client, productID string, quantity int) {
	t.Helper()
	body := map[string]any{"product_id": productID}
	if quantity > 0 {
		body["quantity"] = quantity
	}
	resp := store.do(http.MethodPost, "/api/v1/store/cart/items", body, nil)
	resp.AssertStatus(t, http.StatusOK)
}

func placeStoreOrder(t *testing.T, store *client, productID, email, couponCode string) string {
	t.Helper()
	addProductToCart(t, store, productID, 1)

	checkoutBody := map[string]any{
		"shipping_method":  "post",
		"shipping_city":    "Tehran",
		"customer":         checkoutCustomer(email),
		"shipping_address": standardShippingAddress(),
		"payment_method":   "cod",
	}
	if couponCode != "" {
		checkoutBody["coupon_code"] = couponCode
	}

	resp := store.do(http.MethodPost, "/api/v1/store/checkout", checkoutBody, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var placed struct {
		OrderID       string `json:"order_id"`
		OrderNumber   string `json:"order_number"`
		Status        string `json:"status"`
		PaymentStatus string `json:"payment_status"`
		TotalToman    int64  `json:"total_toman"`
	}
	resp.JSON(t, &placed)
	if placed.OrderID == "" || placed.OrderNumber == "" {
		t.Fatalf("expected order id and number: %+v", placed)
	}
	if placed.Status != "pending" {
		t.Fatalf("status = %q, want pending", placed.Status)
	}
	if placed.PaymentStatus != "unpaid" {
		t.Fatalf("payment_status = %q, want unpaid", placed.PaymentStatus)
	}
	if placed.TotalToman <= 0 {
		t.Fatalf("total_toman = %d, want > 0", placed.TotalToman)
	}
	return placed.OrderID
}

func createActiveProduct(t *testing.T, admin *client) string {
	t.Helper()
	n := uniqueCounter.Add(1)
	name := fmt.Sprintf("E2E Product %d-%d", testRunID, n)
	resp := admin.do(http.MethodPost, "/api/v1/admin/products", map[string]any{
		"name":  name,
		"price": 50000,
		"status": "active",
		"inventory": map[string]any{
			"quantity": 10,
		},
	}, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var product struct {
		ID string `json:"id"`
	}
	resp.JSON(t, &product)
	if product.ID == "" {
		t.Fatal("expected product id")
	}

	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/products/"+product.ID, nil, nil)
	})
	return product.ID
}

func (c *client) uploadFile(t *testing.T, filename, contentType string, content []byte) apiResponse {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/admin/uploads", &body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/admin/uploads: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return apiResponse{
		Status: resp.StatusCode,
		Body:   respBody,
		Header: resp.Header,
	}
}

// minimalPNG is a valid 1x1 transparent PNG for upload tests.
var minimalPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}
