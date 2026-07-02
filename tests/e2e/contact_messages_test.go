//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func submitContactMessage(t *testing.T, source string) string {
	t.Helper()
	store := storeClient(t)
	resp := store.do(http.MethodPost, "/api/v1/store/contact", contactFormBody(source), nil)
	resp.AssertStatus(t, http.StatusCreated)

	var msg struct {
		ID string `json:"id"`
	}
	resp.JSON(t, &msg)
	if msg.ID == "" {
		t.Fatal("expected message id")
	}
	return msg.ID
}

func TestContactMessages_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/contact-messages", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestContactMessages_Lifecycle(t *testing.T) {
	marker := fmt.Sprintf("E2E contact marker %d", testRunID)
	store := storeClient(t)

	submitResp := store.do(http.MethodPost, "/api/v1/store/contact", map[string]any{
		"name":    "E2E Inbox User",
		"email":   uniqueEmail("inbox"),
		"subject": marker,
		"message": "Please follow up about bulk order pricing.",
		"source":  "contact_page",
	}, nil)
	submitResp.AssertStatus(t, http.StatusCreated)

	var submitted struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	submitResp.JSON(t, &submitted)

	c := adminClient(t)

	statsBeforeResp := c.do(http.MethodGet, "/api/v1/admin/contact-messages/stats", nil, nil)
	statsBeforeResp.AssertStatus(t, http.StatusOK)

	var statsBefore struct {
		UnreadCount int64 `json:"unread_count"`
		TotalCount  int64 `json:"total_count"`
	}
	statsBeforeResp.JSON(t, &statsBefore)
	if statsBefore.TotalCount == 0 {
		t.Fatal("expected total_count > 0")
	}

	listResp := c.do(http.MethodGet, "/api/v1/admin/contact-messages?status=unread&source=contact_page&page=1&per_page=20", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Page int `json:"page"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Page != 1 {
		t.Fatalf("page = %d, want 1", list.Meta.Page)
	}

	searchResp := c.do(http.MethodGet, "/api/v1/admin/contact-messages?q="+url.QueryEscape(marker), nil, nil)
	searchResp.AssertStatus(t, http.StatusOK)

	var search struct {
		Data []struct {
			ID      string `json:"id"`
			Subject string `json:"subject"`
			Status  string `json:"status"`
		} `json:"data"`
	}
	searchResp.JSON(t, &search)

	found := false
	for _, row := range search.Data {
		if row.ID == submitted.ID {
			found = true
			if row.Subject != marker {
				t.Fatalf("subject = %q", row.Subject)
			}
			if row.Status != "unread" {
				t.Fatalf("status = %q, want unread", row.Status)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected message in admin search results")
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/contact-messages/"+submitted.ID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var detail struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Email   string `json:"email"`
	}
	getResp.JSON(t, &detail)
	if detail.ID != submitted.ID || detail.Message == "" || detail.Email == "" {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	readResp := c.do(http.MethodPatch, "/api/v1/admin/contact-messages/"+submitted.ID+"/read", nil, nil)
	readResp.AssertStatus(t, http.StatusNoContent)

	getAfterRead := c.do(http.MethodGet, "/api/v1/admin/contact-messages/"+submitted.ID, nil, nil)
	getAfterRead.AssertStatus(t, http.StatusOK)

	var read struct {
		Status string `json:"status"`
	}
	getAfterRead.JSON(t, &read)
	if read.Status != "read" {
		t.Fatalf("status = %q, want read", read.Status)
	}

	archiveResp := c.do(http.MethodPatch, "/api/v1/admin/contact-messages/"+submitted.ID+"/archive", nil, nil)
	archiveResp.AssertStatus(t, http.StatusNoContent)

	statusResp := c.do(http.MethodPatch, "/api/v1/admin/contact-messages/"+submitted.ID+"/status", map[string]string{
		"status": "unread",
	}, nil)
	statusResp.AssertStatus(t, http.StatusNoContent)

	getUnreadAgain := c.do(http.MethodGet, "/api/v1/admin/contact-messages/"+submitted.ID, nil, nil)
	getUnreadAgain.AssertStatus(t, http.StatusOK)

	var unreadAgain struct {
		Status string `json:"status"`
	}
	getUnreadAgain.JSON(t, &unreadAgain)
	if unreadAgain.Status != "unread" {
		t.Fatalf("status = %q, want unread", unreadAgain.Status)
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/contact-messages/"+submitted.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := c.do(http.MethodGet, "/api/v1/admin/contact-messages/"+submitted.ID, nil, nil)
	notFoundResp.AssertStatus(t, http.StatusNotFound)
	notFoundResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestContactMessages_StatusValidation(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodPatch, "/api/v1/admin/contact-messages/00000000-0000-0000-0000-000000000099/status", map[string]string{
		"status": "invalid",
	}, nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestContactMessages_Get_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/contact-messages/00000000-0000-0000-0000-000000000099", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestContactMessages_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/contact-messages", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestContactMessages_StatsReflectsUnread(t *testing.T) {
	msgID := submitContactMessage(t, "about")

	c := adminClient(t)
	statsResp := c.do(http.MethodGet, "/api/v1/admin/contact-messages/stats", nil, nil)
	statsResp.AssertStatus(t, http.StatusOK)

	var stats struct {
		UnreadCount int64 `json:"unread_count"`
		TotalCount  int64 `json:"total_count"`
	}
	statsResp.JSON(t, &stats)
	if stats.UnreadCount < 1 || stats.TotalCount < 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/contact-messages/"+msgID, nil, nil)
	})
}

func TestContactMessages_RateLimitOnPublicForm(t *testing.T) {
	store := storeClient(t)
	admin := adminClient(t)

	for i := 0; i < 11; i++ {
		body := contactFormBody("homepage")
		body["email"] = uniqueEmail(fmt.Sprintf("rate-%d", i))

		resp := store.do(http.MethodPost, "/api/v1/store/contact", body, nil)
		if i < 10 {
			resp.AssertStatus(t, http.StatusCreated)
			var msg struct {
				ID string `json:"id"`
			}
			resp.JSON(t, &msg)
			msgID := msg.ID
			t.Cleanup(func() {
				_ = admin.do(http.MethodDelete, "/api/v1/admin/contact-messages/"+msgID, nil, nil)
			})
			continue
		}
		if resp.Status != http.StatusTooManyRequests {
			t.Fatalf("request %d: status = %d, want 429\nbody: %s", i+1, resp.Status, string(resp.Body))
		}
		resp.AssertErrorCode(t, "RATE_LIMITED")
	}
}
