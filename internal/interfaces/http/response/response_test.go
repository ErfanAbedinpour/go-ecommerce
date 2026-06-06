package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"app/pkg/apperror"
)

func TestError_IncludesStatusCodeAndPath(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)

	Error(w, r, nil, apperror.Validation("request validation failed", map[string]string{
		"email": "is required",
	}))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var body apperror.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", body.StatusCode, http.StatusBadRequest)
	}
	if body.Path != "/api/v1/auth/login" {
		t.Errorf("Path = %q, want %q", body.Path, "/api/v1/auth/login")
	}
	if body.Error.Code != apperror.CodeValidation {
		t.Errorf("Error.Code = %q, want %q", body.Error.Code, apperror.CodeValidation)
	}
}
