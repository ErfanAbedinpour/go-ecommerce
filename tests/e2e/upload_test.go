//go:build e2e

package e2e

import (
	"io"
	"net/http"
	"testing"
)

func TestUpload_Unauthorized(t *testing.T) {
	resp := testClient.uploadFile(t, "test.png", "image/png", minimalPNG)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestUpload_Success(t *testing.T) {
	c := adminClient(t)

	resp := c.uploadFile(t, "e2e-test.png", "image/png", minimalPNG)
	resp.AssertStatus(t, http.StatusCreated)

	var uploaded struct {
		URL         string `json:"url"`
		Filename    string `json:"filename"`
		Size        int64  `json:"size"`
		ContentType string `json:"content_type"`
	}
	resp.JSON(t, &uploaded)

	if uploaded.URL == "" {
		t.Fatal("expected url in upload response")
	}
	if uploaded.Filename == "" {
		t.Fatal("expected filename in upload response")
	}
	if uploaded.Size != int64(len(minimalPNG)) {
		t.Fatalf("size = %d, want %d", uploaded.Size, len(minimalPNG))
	}
	if uploaded.ContentType != "image/png" {
		t.Fatalf("content_type = %q, want image/png", uploaded.ContentType)
	}
}

func TestUpload_MissingFile(t *testing.T) {
	c := adminClient(t)

	req, err := http.NewRequest(http.MethodPost, testClient.baseURL+"/api/v1/admin/uploads", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	httpResp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("POST upload: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", httpResp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, httpResp.Body)
}

func TestUpload_InvalidFileType(t *testing.T) {
	c := adminClient(t)

	resp := c.uploadFile(t, "notes.txt", "text/plain", []byte("not an image"))
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestUpload_EmptyFile(t *testing.T) {
	c := adminClient(t)

	resp := c.uploadFile(t, "empty.png", "image/png", nil)
	resp.AssertStatus(t, http.StatusBadRequest)
	resp.AssertErrorCode(t, "VALIDATION_ERROR")
}

func TestUpload_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.uploadFile(t, "test.png", "image/png", minimalPNG)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}
