//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestDashboard_Unauthorized(t *testing.T) {
	endpoints := []string{
		"/api/v1/admin/dashboard/stats",
		"/api/v1/admin/dashboard/revenue",
		"/api/v1/admin/dashboard/low-stock",
		"/api/v1/admin/dashboard/recent-orders",
		"/api/v1/admin/dashboard/featured-products",
	}

	for _, path := range endpoints {
		t.Run(path, func(t *testing.T) {
			resp := testClient.do(http.MethodGet, path, nil, nil)
			resp.AssertStatus(t, http.StatusUnauthorized)
		})
	}
}

func TestDashboard_Stats(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodGet, "/api/v1/admin/dashboard/stats", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var stats struct {
		TotalRevenue   float64 `json:"total_revenue"`
		TotalOrders    int64   `json:"total_orders"`
		TotalCustomers int64   `json:"total_customers"`
		TotalProducts  int64   `json:"total_products"`
		PendingOrders  int64   `json:"pending_orders"`
		LowStockCount  int64   `json:"low_stock_count"`
		Growth         struct {
			TotalRevenue   float64 `json:"total_revenue"`
			TotalOrders    float64 `json:"total_orders"`
			TotalCustomers float64 `json:"total_customers"`
			TotalProducts  float64 `json:"total_products"`
			PendingOrders  float64 `json:"pending_orders"`
			LowStockCount  float64 `json:"low_stock_count"`
		} `json:"growth"`
	}
	resp.JSON(t, &stats)

	if stats.TotalProducts < 0 {
		t.Fatalf("total_products = %d, want >= 0", stats.TotalProducts)
	}
	if stats.TotalOrders < 0 {
		t.Fatalf("total_orders = %d, want >= 0", stats.TotalOrders)
	}
	if stats.TotalCustomers < 0 {
		t.Fatalf("total_customers = %d, want >= 0", stats.TotalCustomers)
	}
}

func TestDashboard_Revenue(t *testing.T) {
	c := adminClient(t)

	periods := []string{"today", "week", "month", "year"}
	for _, period := range periods {
		t.Run("period="+period, func(t *testing.T) {
			resp := c.do(http.MethodGet, "/api/v1/admin/dashboard/revenue?period="+period, nil, nil)
			resp.AssertStatus(t, http.StatusOK)

			var result struct {
				Data []struct {
					Date    string  `json:"date"`
					Revenue float64 `json:"revenue"`
					Orders  int64   `json:"orders"`
				} `json:"data"`
			}
			resp.JSON(t, &result)
			if result.Data == nil {
				t.Fatal("expected data array in revenue response")
			}
		})
	}

	customResp := c.do(http.MethodGet, "/api/v1/admin/dashboard/revenue?from=2026-01-01&to=2026-01-31", nil, nil)
	customResp.AssertStatus(t, http.StatusOK)

	invalidPeriodResp := c.do(http.MethodGet, "/api/v1/admin/dashboard/revenue?period=invalid", nil, nil)
	invalidPeriodResp.AssertStatus(t, http.StatusBadRequest)
	invalidPeriodResp.AssertErrorCode(t, "VALIDATION_ERROR")

	partialRangeResp := c.do(http.MethodGet, "/api/v1/admin/dashboard/revenue?from=2026-01-01", nil, nil)
	partialRangeResp.AssertStatus(t, http.StatusBadRequest)
	partialRangeResp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestDashboard_LowStock(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodGet, "/api/v1/admin/dashboard/low-stock?page=1&per_page=5", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
		Meta struct {
			Page    int   `json:"page"`
			PerPage int   `json:"per_page"`
			Total   int64 `json:"total"`
		} `json:"meta"`
	}
	resp.JSON(t, &result)
	if result.Meta.Page != 1 || result.Meta.PerPage != 5 {
		t.Fatalf("unexpected pagination meta: %+v", result.Meta)
	}
}

func TestDashboard_RecentOrders(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodGet, "/api/v1/admin/dashboard/recent-orders?limit=5", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var result struct {
		Data []struct {
			ID          string  `json:"id"`
			OrderNumber string  `json:"order_number"`
			Status      string  `json:"status"`
			Total       float64 `json:"total"`
		} `json:"data"`
	}
	resp.JSON(t, &result)
	if result.Data == nil {
		t.Fatal("expected data array in recent orders response")
	}

	invalidLimitResp := c.do(http.MethodGet, "/api/v1/admin/dashboard/recent-orders?limit=abc", nil, nil)
	invalidLimitResp.AssertStatus(t, http.StatusBadRequest)
	invalidLimitResp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestDashboard_FeaturedProducts(t *testing.T) {
	c := adminClient(t)

	resp := c.do(http.MethodGet, "/api/v1/admin/dashboard/featured-products?limit=3", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			IsFeatured bool   `json:"is_featured"`
		} `json:"data"`
	}
	resp.JSON(t, &result)
	if result.Data == nil {
		t.Fatal("expected data array in featured products response")
	}

	invalidLimitResp := c.do(http.MethodGet, "/api/v1/admin/dashboard/featured-products?limit=bad", nil, nil)
	invalidLimitResp.AssertStatus(t, http.StatusBadRequest)
	invalidLimitResp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestDashboard_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/dashboard/stats", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}
