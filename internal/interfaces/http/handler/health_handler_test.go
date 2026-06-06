package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHealthChecker struct {
	pingErr error
}

func (m *mockHealthChecker) Ping(_ context.Context) error {
	return m.pingErr
}

func TestHealthHandler_Liveness(t *testing.T) {
	h := NewHealthHandler(&mockHealthChecker{}, "1.0.0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h.Liveness(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp healthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", resp.Version, "1.0.0")
	}
}

func TestHealthHandler_Readiness_Healthy(t *testing.T) {
	h := NewHealthHandler(&mockHealthChecker{}, "1.0.0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	h.Readiness(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp readinessResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "ready" {
		t.Errorf("status = %q, want %q", resp.Status, "ready")
	}
	if resp.Checks["database"] != "healthy" {
		t.Errorf("database check = %q, want %q", resp.Checks["database"], "healthy")
	}
}

func TestHealthHandler_Readiness_Unhealthy(t *testing.T) {
	h := NewHealthHandler(&mockHealthChecker{pingErr: errors.New("connection refused")}, "1.0.0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	h.Readiness(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var resp readinessResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "not_ready" {
		t.Errorf("status = %q, want %q", resp.Status, "not_ready")
	}
}
