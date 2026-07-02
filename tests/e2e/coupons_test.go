//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCoupons_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/coupons", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestCoupons_Create_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing code",
			body: map[string]any{
				"discount_type":  "percentage",
				"discount_value": 10,
			},
		},
		{
			name: "code too short",
			body: map[string]any{
				"code":           "AB",
				"discount_type":  "percentage",
				"discount_value": 10,
			},
		},
		{
			name: "invalid discount type",
			body: map[string]any{
				"code":           uniqueCouponCode("BAD"),
				"discount_type":  "free",
				"discount_value": 10,
			},
		},
		{
			name: "zero discount value",
			body: map[string]any{
				"code":           uniqueCouponCode("ZERO"),
				"discount_type":  "percentage",
				"discount_value": 0,
			},
		},
		{
			name: "percentage over 100",
			body: map[string]any{
				"code":           uniqueCouponCode("HIGH"),
				"discount_type":  "percentage",
				"discount_value": 150,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPost, "/api/v1/admin/coupons", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestCoupons_CRUD_Lifecycle(t *testing.T) {
	c := adminClient(t)
	code := uniqueCouponCode("E2ELIFE")
	expires := time.Now().UTC().Add(30 * 24 * time.Hour)

	createResp := c.do(http.MethodPost, "/api/v1/admin/coupons", map[string]any{
		"code":             code,
		"discount_type":    "fixed_amount",
		"discount_value":   25000,
		"min_order_amount": 100000,
		"max_usage":        100,
		"expires_at":       expires.Format(time.RFC3339),
		"note":             "E2E coupon lifecycle test",
		"is_active":        true,
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID             string  `json:"id"`
		Code           string  `json:"code"`
		DiscountType   string  `json:"discount_type"`
		DiscountValue  float64 `json:"discount_value"`
		MinOrderAmount float64 `json:"min_order_amount"`
		IsActive       bool    `json:"is_active"`
		Note           string  `json:"note"`
	}
	createResp.JSON(t, &created)

	if created.ID == "" {
		t.Fatal("expected coupon id")
	}
	if created.Code != code {
		t.Fatalf("code = %q, want %q", created.Code, code)
	}
	if created.DiscountType != "fixed_amount" {
		t.Fatalf("discount_type = %q", created.DiscountType)
	}
	if created.DiscountValue != 25000 {
		t.Fatalf("discount_value = %v", created.DiscountValue)
	}
	if !created.IsActive {
		t.Fatal("expected is_active=true")
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/coupons/"+created.ID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	listResp := c.do(http.MethodGet, "/api/v1/admin/coupons?q="+code+"&is_active=true", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	listResp.JSON(t, &list)
	if len(list.Data) == 0 {
		t.Fatal("expected coupon in list search")
	}

	newCode := uniqueCouponCode("E2EUPD")
	updateResp := c.do(http.MethodPut, "/api/v1/admin/coupons/"+created.ID, map[string]any{
		"code":             newCode,
		"discount_type":    "percentage",
		"discount_value":   15,
		"min_order_amount": 50000,
		"note":             "Updated e2e coupon",
		"is_active":        true,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		Code          string  `json:"code"`
		DiscountType  string  `json:"discount_type"`
		DiscountValue float64 `json:"discount_value"`
		Note          string  `json:"note"`
	}
	updateResp.JSON(t, &updated)
	if updated.Code != newCode {
		t.Fatalf("updated code = %q", updated.Code)
	}
	if updated.DiscountType != "percentage" {
		t.Fatalf("updated discount_type = %q", updated.DiscountType)
	}

	deactivateResp := c.do(http.MethodPatch, "/api/v1/admin/coupons/"+created.ID+"/deactivate", nil, nil)
	deactivateResp.AssertStatus(t, http.StatusOK)

	var deactivated struct {
		IsActive bool `json:"is_active"`
	}
	deactivateResp.JSON(t, &deactivated)
	if deactivated.IsActive {
		t.Fatal("expected is_active=false after deactivate")
	}

	activateResp := c.do(http.MethodPatch, "/api/v1/admin/coupons/"+created.ID+"/activate", nil, nil)
	activateResp.AssertStatus(t, http.StatusOK)

	var activated struct {
		IsActive bool `json:"is_active"`
	}
	activateResp.JSON(t, &activated)
	if !activated.IsActive {
		t.Fatal("expected is_active=true after activate")
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/coupons/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := c.do(http.MethodGet, "/api/v1/admin/coupons/"+created.ID, nil, nil)
	notFoundResp.AssertStatus(t, http.StatusNotFound)
	notFoundResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestCoupons_DuplicateCode(t *testing.T) {
	c := adminClient(t)
	code := uniqueCouponCode("E2EDUP")

	firstResp := c.do(http.MethodPost, "/api/v1/admin/coupons", map[string]any{
		"code":           code,
		"discount_type":  "percentage",
		"discount_value": 10,
		"is_active":      true,
	}, nil)
	firstResp.AssertStatus(t, http.StatusCreated)

	var first struct {
		ID string `json:"id"`
	}
	firstResp.JSON(t, &first)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/coupons/"+first.ID, nil, nil)
	})

	secondResp := c.do(http.MethodPost, "/api/v1/admin/coupons", map[string]any{
		"code":           code,
		"discount_type":  "percentage",
		"discount_value": 20,
	}, nil)
	secondResp.AssertStatus(t, http.StatusConflict)
	secondResp.AssertErrorCode(t, "CONFLICT")
}

func TestCoupons_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/coupons", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestCoupons_Get_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/coupons/00000000-0000-0000-0000-000000000099", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func uniqueCouponCode(prefix string) string {
	n := uniqueCounter.Add(1)
	return fmt.Sprintf("%s%d", prefix, n)
}
