//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestBlog_Admin_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/blog/posts", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestBlog_Category_CRUD(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-blog-cat-%d", testRunID)

	createResp := c.do(http.MethodPost, "/api/v1/admin/blog/categories", map[string]string{
		"name":        "E2E Blog Category",
		"slug":        slug,
		"description": "Category for e2e tests",
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	createResp.JSON(t, &created)
	if created.ID == "" || created.Slug != slug {
		t.Fatalf("unexpected category: %+v", created)
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/blog/categories/"+created.ID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	listResp := c.do(http.MethodGet, "/api/v1/admin/blog/categories", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var categories []struct {
		ID string `json:"id"`
	}
	listResp.JSON(t, &categories)
	if len(categories) == 0 {
		t.Fatal("expected at least one category")
	}

	newSlug := fmt.Sprintf("e2e-blog-cat-updated-%d", testRunID)
	updateResp := c.do(http.MethodPut, "/api/v1/admin/blog/categories/"+created.ID, map[string]string{
		"name":        "Updated Category",
		"slug":        newSlug,
		"description": "Updated description",
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	updateResp.JSON(t, &updated)
	if updated.Name != "Updated Category" || updated.Slug != newSlug {
		t.Fatalf("unexpected updated category: %+v", updated)
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/blog/categories/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)
}

func TestBlog_Category_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]string
	}{
		{
			name: "missing name",
			body: map[string]string{"slug": "missing-name"},
		},
		{
			name: "missing slug",
			body: map[string]string{"name": "Missing Slug"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPost, "/api/v1/admin/blog/categories", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestBlog_Post_LifecycleAndStorefront(t *testing.T) {
	c := adminClient(t)
	catSlug := fmt.Sprintf("e2e-post-cat-%d", testRunID)
	postSlug := fmt.Sprintf("e2e-blog-post-%d", testRunID)

	catResp := c.do(http.MethodPost, "/api/v1/admin/blog/categories", map[string]string{
		"name": "Post Category",
		"slug": catSlug,
	}, nil)
	catResp.AssertStatus(t, http.StatusCreated)

	var category struct {
		ID string `json:"id"`
	}
	catResp.JSON(t, &category)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/blog/categories/"+category.ID, nil, nil)
	})

	draftResp := c.do(http.MethodPost, "/api/v1/admin/blog/posts", map[string]any{
		"title":       "E2E Draft Post",
		"slug":        postSlug,
		"content":     "Draft content for e2e test",
		"summary":     "Draft summary",
		"category_id": category.ID,
		"status":      "draft",
	}, nil)
	draftResp.AssertStatus(t, http.StatusCreated)

	var draft struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Title  string `json:"title"`
	}
	draftResp.JSON(t, &draft)
	if draft.Status != "draft" {
		t.Fatalf("status = %q, want draft", draft.Status)
	}

	storeDraftResp := testClient.do(http.MethodGet, "/api/v1/store/blog/posts/"+postSlug, nil, nil)
	storeDraftResp.AssertStatus(t, http.StatusNotFound)

	publishResp := c.do(http.MethodPut, "/api/v1/admin/blog/posts/"+draft.ID, map[string]any{
		"title":       "E2E Published Post",
		"slug":        postSlug,
		"content":     "Published content for e2e test",
		"summary":     "Published summary",
		"category_id": category.ID,
		"status":      "published",
	}, nil)
	publishResp.AssertStatus(t, http.StatusOK)

	var published struct {
		Status string `json:"status"`
		Title  string `json:"title"`
	}
	publishResp.JSON(t, &published)
	if published.Status != "published" {
		t.Fatalf("status = %q, want published", published.Status)
	}
	if published.Title != "E2E Published Post" {
		t.Fatalf("title = %q", published.Title)
	}

	storeListResp := testClient.do(http.MethodGet, "/api/v1/store/blog/posts?q=E2E+Published", nil, nil)
	storeListResp.AssertStatus(t, http.StatusOK)

	var storeList struct {
		Data []struct {
			Slug  string `json:"slug"`
			Title string `json:"title"`
		} `json:"data"`
	}
	storeListResp.JSON(t, &storeList)
	if len(storeList.Data) == 0 {
		t.Fatal("expected published post in storefront list")
	}

	storeGetResp := testClient.do(http.MethodGet, "/api/v1/store/blog/posts/"+postSlug, nil, nil)
	storeGetResp.AssertStatus(t, http.StatusOK)

	var storePost struct {
		Slug    string `json:"slug"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	storeGetResp.JSON(t, &storePost)
	if storePost.Slug != postSlug {
		t.Fatalf("slug = %q, want %q", storePost.Slug, postSlug)
	}
	if storePost.Title != "E2E Published Post" {
		t.Fatalf("title = %q", storePost.Title)
	}
	if storePost.Content == "" {
		t.Fatal("expected content in storefront post")
	}

	commentResp := testClient.do(http.MethodPost, "/api/v1/store/blog/posts/"+draft.ID+"/comments", map[string]string{
		"author_name":  "E2E Reader",
		"author_email": "reader@test.local",
		"content":      "Great post from e2e test!",
	}, nil)
	commentResp.AssertStatus(t, http.StatusCreated)

	var comment struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	commentResp.JSON(t, &comment)
	if comment.Status != "pending" {
		t.Fatalf("comment status = %q, want pending", comment.Status)
	}

	pendingCommentsResp := testClient.do(http.MethodGet, "/api/v1/store/blog/posts/"+draft.ID+"/comments", nil, nil)
	pendingCommentsResp.AssertStatus(t, http.StatusOK)

	var pendingComments struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	pendingCommentsResp.JSON(t, &pendingComments)
	if len(pendingComments.Data) != 0 {
		t.Fatal("expected no approved comments before moderation")
	}

	adminCommentsResp := c.do(http.MethodGet, "/api/v1/admin/blog/comments?status=pending&post_id="+draft.ID, nil, nil)
	adminCommentsResp.AssertStatus(t, http.StatusOK)

	approveResp := c.do(http.MethodPatch, "/api/v1/admin/blog/comments/"+comment.ID+"/approve", nil, nil)
	approveResp.AssertStatus(t, http.StatusNoContent)

	approvedCommentsResp := testClient.do(http.MethodGet, "/api/v1/store/blog/posts/"+draft.ID+"/comments", nil, nil)
	approvedCommentsResp.AssertStatus(t, http.StatusOK)

	var approvedComments struct {
		Data []struct {
			ID      string `json:"id"`
			Content string `json:"content"`
			Status  string `json:"status"`
		} `json:"data"`
	}
	approvedCommentsResp.JSON(t, &approvedComments)
	if len(approvedComments.Data) != 1 {
		t.Fatalf("expected 1 approved comment, got %d", len(approvedComments.Data))
	}
	if approvedComments.Data[0].Content != "Great post from e2e test!" {
		t.Fatalf("comment content = %q", approvedComments.Data[0].Content)
	}

	slugCommentsResp := testClient.do(http.MethodGet, "/api/v1/store/blog/"+postSlug+"/comments", nil, nil)
	slugCommentsResp.AssertStatus(t, http.StatusOK)

	deletePostResp := c.do(http.MethodDelete, "/api/v1/admin/blog/posts/"+draft.ID, nil, nil)
	deletePostResp.AssertStatus(t, http.StatusNoContent)
}

func TestBlog_Post_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing title",
			body: map[string]any{
				"slug":    "missing-title",
				"content": "content",
			},
		},
		{
			name: "missing content",
			body: map[string]any{
				"title": "Missing Content",
				"slug":  "missing-content",
			},
		},
		{
			name: "invalid status",
			body: map[string]any{
				"title":   "Bad Status",
				"slug":    "bad-status",
				"content": "content",
				"status":  "archived",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPost, "/api/v1/admin/blog/posts", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestBlog_Storefront_ListCategories(t *testing.T) {
	c := adminClient(t)
	slug := fmt.Sprintf("e2e-store-cat-%d", testRunID)

	createResp := c.do(http.MethodPost, "/api/v1/admin/blog/categories", map[string]string{
		"name": "Storefront Category",
		"slug": slug,
	}, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var cat struct {
		ID string `json:"id"`
	}
	createResp.JSON(t, &cat)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/blog/categories/"+cat.ID, nil, nil)
	})

	listResp := testClient.do(http.MethodGet, "/api/v1/store/blog/categories", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var categories []struct {
		Slug string `json:"slug"`
	}
	listResp.JSON(t, &categories)
	if len(categories) == 0 {
		t.Fatal("expected storefront categories")
	}
}

func TestBlog_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/blog/posts", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}
