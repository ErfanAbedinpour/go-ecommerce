//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func reviewSubmitBody(author string, rating int) map[string]any {
	return map[string]any{
		"author_name": author,
		"rating":      rating,
		"title":       "E2E review title",
		"content":     "E2E review content describing the product experience.",
	}
}

func TestProductReviews_Submit_ValidationErrors(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createActiveProduct(t, admin)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing author_name",
			body: map[string]any{
				"rating":  5,
				"content": "Content without author",
			},
		},
		{
			name: "rating too low",
			body: map[string]any{
				"author_name": "Guest",
				"rating":      0,
				"content":     "Invalid rating",
			},
		},
		{
			name: "rating too high",
			body: map[string]any{
				"author_name": "Guest",
				"rating":      6,
				"content":     "Invalid rating",
			},
		},
		{
			name: "missing content",
			body: map[string]any{
				"author_name": "Guest",
				"rating":      4,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/reviews", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestProductReviews_Lifecycle_GuestSubmitAndModeration(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createActiveProduct(t, admin)

	submitResp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/reviews",
		reviewSubmitBody("E2E Guest Reviewer", 5), nil)
	submitResp.AssertStatus(t, http.StatusCreated)

	var submitted struct {
		ID         string `json:"id"`
		ProductID  string `json:"product_id"`
		AuthorName string `json:"author_name"`
		Rating     int    `json:"rating"`
		Status     string `json:"status"`
	}
	submitResp.JSON(t, &submitted)
	if submitted.ID == "" || submitted.ProductID != productID {
		t.Fatalf("unexpected review: %+v", submitted)
	}
	if submitted.Status != "pending" {
		t.Fatalf("status = %q, want pending", submitted.Status)
	}
	if submitted.Rating != 5 {
		t.Fatalf("rating = %d, want 5", submitted.Rating)
	}

	pendingListResp := store.do(http.MethodGet, "/api/v1/store/products/"+productID+"/reviews", nil, nil)
	pendingListResp.AssertStatus(t, http.StatusOK)

	var pendingList struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	pendingListResp.JSON(t, &pendingList)
	for _, row := range pendingList.Data {
		if row.ID == submitted.ID {
			t.Fatal("pending review should not appear in public list")
		}
	}

	summaryBeforeResp := store.do(http.MethodGet, "/api/v1/store/products/"+productID+"/reviews/summary", nil, nil)
	summaryBeforeResp.AssertStatus(t, http.StatusOK)

	var summaryBefore struct {
		TotalCount int64 `json:"total_count"`
	}
	summaryBeforeResp.JSON(t, &summaryBefore)
	if summaryBefore.TotalCount != 0 {
		t.Fatalf("total_count = %d, want 0 before approval", summaryBefore.TotalCount)
	}

	c := adminClient(t)
	adminListResp := c.do(http.MethodGet, "/api/v1/admin/reviews?status=pending&product_id="+productID, nil, nil)
	adminListResp.AssertStatus(t, http.StatusOK)

	var adminList struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	adminListResp.JSON(t, &adminList)

	found := false
	for _, row := range adminList.Data {
		if row.ID == submitted.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected pending review in admin list")
	}

	approveResp := c.do(http.MethodPatch, "/api/v1/admin/reviews/"+submitted.ID+"/status", map[string]string{
		"status": "approved",
	}, nil)
	approveResp.AssertStatus(t, http.StatusNoContent)

	publicListResp := store.do(http.MethodGet, "/api/v1/store/products/"+productID+"/reviews?sort=newest", nil, nil)
	publicListResp.AssertStatus(t, http.StatusOK)

	var publicList struct {
		Data []struct {
			ID         string `json:"id"`
			AuthorName string `json:"author_name"`
			Status     string `json:"status"`
		} `json:"data"`
	}
	publicListResp.JSON(t, &publicList)

	foundPublic := false
	for _, row := range publicList.Data {
		if row.ID == submitted.ID {
			foundPublic = true
			if row.Status != "approved" {
				t.Fatalf("status = %q, want approved", row.Status)
			}
			if row.AuthorName != "E2E Guest Reviewer" {
				t.Fatalf("author_name = %q", row.AuthorName)
			}
			break
		}
	}
	if !foundPublic {
		t.Fatal("approved review should appear in public list")
	}

	summaryAfterResp := store.do(http.MethodGet, "/api/v1/store/products/"+productID+"/reviews/summary", nil, nil)
	summaryAfterResp.AssertStatus(t, http.StatusOK)

	var summaryAfter struct {
		AverageRating float64       `json:"average_rating"`
		TotalCount    int64         `json:"total_count"`
		Distribution  map[int]int64 `json:"distribution"`
	}
	summaryAfterResp.JSON(t, &summaryAfter)
	if summaryAfter.TotalCount < 1 {
		t.Fatalf("total_count = %d, want >= 1", summaryAfter.TotalCount)
	}
	if summaryAfter.AverageRating < 1 {
		t.Fatalf("average_rating = %v", summaryAfter.AverageRating)
	}

	rejectOtherResp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/reviews",
		reviewSubmitBody("Another Guest", 2), nil)
	rejectOtherResp.AssertStatus(t, http.StatusCreated)

	var otherReview struct {
		ID string `json:"id"`
	}
	rejectOtherResp.JSON(t, &otherReview)

	rejectResp := c.do(http.MethodPatch, "/api/v1/admin/reviews/"+otherReview.ID+"/status", map[string]string{
		"status": "rejected",
	}, nil)
	rejectResp.AssertStatus(t, http.StatusNoContent)

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/reviews/"+submitted.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	deleteOtherResp := c.do(http.MethodDelete, "/api/v1/admin/reviews/"+otherReview.ID, nil, nil)
	deleteOtherResp.AssertStatus(t, http.StatusNoContent)
}

func TestProductReviews_Submit_DuplicateAuthenticatedCustomer(t *testing.T) {
	admin := adminClient(t)
	store := customerStoreClient(t)
	productID := createActiveProduct(t, admin)

	firstResp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/reviews",
		reviewSubmitBody("E2E Customer", 4), nil)
	firstResp.AssertStatus(t, http.StatusCreated)

	secondResp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/reviews",
		reviewSubmitBody("E2E Customer", 3), nil)
	secondResp.AssertStatus(t, http.StatusConflict)
	secondResp.AssertErrorCode(t, "CONFLICT")

	c := adminClient(t)
	var first struct {
		ID string `json:"id"`
	}
	firstResp.JSON(t, &first)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/reviews/"+first.ID, nil, nil)
	})
}

func TestProductReviews_ProductNotFound(t *testing.T) {
	store := storeClient(t)
	missingID := "00000000-0000-0000-0000-000000000099"

	submitResp := store.do(http.MethodPost, "/api/v1/store/products/"+missingID+"/reviews",
		reviewSubmitBody("Guest", 5), nil)
	submitResp.AssertStatus(t, http.StatusNotFound)
	submitResp.AssertErrorCode(t, "NOT_FOUND")

	listResp := store.do(http.MethodGet, "/api/v1/store/products/"+missingID+"/reviews", nil, nil)
	listResp.AssertStatus(t, http.StatusNotFound)
}

func TestProductReviews_Admin_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/reviews", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestProductReviews_Admin_StatusValidation(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodPatch, "/api/v1/admin/reviews/00000000-0000-0000-0000-000000000099/status", map[string]string{
		"status": "invalid",
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestProductReviews_Admin_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/reviews", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestProductReviews_Admin_ListWithFilters(t *testing.T) {
	admin := adminClient(t)
	store := storeClient(t)
	productID := createActiveProduct(t, admin)

	submitResp := store.do(http.MethodPost, "/api/v1/store/products/"+productID+"/reviews",
		reviewSubmitBody(fmt.Sprintf("Filter Test %d", testRunID), 3), nil)
	submitResp.AssertStatus(t, http.StatusCreated)

	var review struct {
		ID string `json:"id"`
	}
	submitResp.JSON(t, &review)

	c := adminClient(t)
	filterResp := c.do(http.MethodGet, fmt.Sprintf("/api/v1/admin/reviews?product_id=%s&rating=3&status=pending&q=Filter", productID), nil, nil)
	filterResp.AssertStatus(t, http.StatusOK)

	var filtered struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Page int `json:"page"`
		} `json:"meta"`
	}
	filterResp.JSON(t, &filtered)
	if filtered.Meta.Page != 1 {
		t.Fatalf("page = %d, want 1", filtered.Meta.Page)
	}

	found := false
	for _, row := range filtered.Data {
		if row.ID == review.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected review in filtered admin list")
	}

	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/reviews/"+review.ID, nil, nil)
	})
}
