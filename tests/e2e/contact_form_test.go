//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func contactFormBody(source string) map[string]any {
	return map[string]any{
		"name":    "E2E Contact User",
		"email":   uniqueEmail("contact"),
		"phone":   "+989121234567",
		"subject": fmt.Sprintf("E2E inquiry %d", testRunID),
		"message": "Hello, I would like more information about your products and shipping options.",
		"source":  source,
	}
}

func TestContactForm_Submit_Success(t *testing.T) {
	store := storeClient(t)

	resp := store.do(http.MethodPost, "/api/v1/store/contact", contactFormBody("contact_page"), nil)
	resp.AssertStatus(t, http.StatusCreated)

	var msg struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Subject string `json:"subject"`
		Message string `json:"message"`
		Source  string `json:"source"`
		Status  string `json:"status"`
	}
	resp.JSON(t, &msg)

	if msg.ID == "" {
		t.Fatal("expected message id")
	}
	if msg.Name != "E2E Contact User" {
		t.Fatalf("name = %q", msg.Name)
	}
	if msg.Source != "contact_page" {
		t.Fatalf("source = %q, want contact_page", msg.Source)
	}
	if msg.Status != "unread" {
		t.Fatalf("status = %q, want unread", msg.Status)
	}
	if msg.Message == "" || msg.Email == "" {
		t.Fatalf("unexpected message payload: %+v", msg)
	}

	admin := adminClient(t)
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/contact-messages/"+msg.ID, nil, nil)
	})
}

func TestContactForm_Submit_DefaultSource(t *testing.T) {
	store := storeClient(t)

	body := contactFormBody("")
	delete(body, "source")

	resp := store.do(http.MethodPost, "/api/v1/store/contact", body, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var msg struct {
		ID     string `json:"id"`
		Source string `json:"source"`
	}
	resp.JSON(t, &msg)
	if msg.Source != "homepage" {
		t.Fatalf("source = %q, want homepage default", msg.Source)
	}

	admin := adminClient(t)
	t.Cleanup(func() {
		_ = admin.do(http.MethodDelete, "/api/v1/admin/contact-messages/"+msg.ID, nil, nil)
	})
}

func TestContactForm_Submit_ValidationErrors(t *testing.T) {
	store := storeClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing name",
			body: map[string]any{
				"email":   "guest@example.com",
				"message": "Hello",
			},
		},
		{
			name: "invalid email",
			body: map[string]any{
				"name":    "Guest",
				"email":   "not-an-email",
				"message": "Hello",
			},
		},
		{
			name: "missing message",
			body: map[string]any{
				"name":  "Guest",
				"email": "guest@example.com",
			},
		},
		{
			name: "invalid source",
			body: map[string]any{
				"name":    "Guest",
				"email":   uniqueEmail("bad-source"),
				"message": "Hello",
				"source":  "newsletter",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := store.do(http.MethodPost, "/api/v1/store/contact", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestContactForm_Submit_AllSources(t *testing.T) {
	store := storeClient(t)
	admin := adminClient(t)

	for _, source := range []string{"homepage", "about", "contact_page"} {
		t.Run(source, func(t *testing.T) {
			resp := store.do(http.MethodPost, "/api/v1/store/contact", contactFormBody(source), nil)
			resp.AssertStatus(t, http.StatusCreated)

			var msg struct {
				ID     string `json:"id"`
				Source string `json:"source"`
			}
			resp.JSON(t, &msg)
			if msg.Source != source {
				t.Fatalf("source = %q, want %q", msg.Source, source)
			}
			msgID := msg.ID
			t.Cleanup(func() {
				_ = admin.do(http.MethodDelete, "/api/v1/admin/contact-messages/"+msgID, nil, nil)
			})
		})
	}
}
