package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AppError
		wantSub string
	}{
		{
			name:    "without wrapped error",
			err:     New(CodeNotFound, "product not found", http.StatusNotFound),
			wantSub: "product not found",
		},
		{
			name:    "with wrapped error",
			err:     Internal("db error").WithError(errors.New("connection refused")),
			wantSub: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); !contains(got, tt.wantSub) {
				t.Errorf("Error() = %q, want substring %q", got, tt.wantSub)
			}
		})
	}
}

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"app error", NotFound("product"), http.StatusNotFound},
		{"wrapped app error", Validation("bad", nil), http.StatusBadRequest},
		{"generic error", errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusCode(tt.err); got != tt.want {
				t.Errorf("StatusCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	appErr := Validation("request validation failed", map[string]string{"email": "is required"})
	resp := NewErrorResponse(appErr, "/api/v1/auth/login")

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if resp.Path != "/api/v1/auth/login" {
		t.Errorf("Path = %q, want %q", resp.Path, "/api/v1/auth/login")
	}
	if resp.Error.Code != CodeValidation {
		t.Errorf("Error.Code = %q, want %q", resp.Error.Code, CodeValidation)
	}
}

func TestIsAppError(t *testing.T) {
	if !IsAppError(NotFound("x")) {
		t.Error("expected IsAppError to return true for AppError")
	}
	if IsAppError(errors.New("generic")) {
		t.Error("expected IsAppError to return false for generic error")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || findSub(s, sub))
}

func findSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
