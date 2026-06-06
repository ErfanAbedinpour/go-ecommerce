package handler

import (
	"context"
	"net/http"
	"time"

	"app/internal/interfaces/http/response"
)

// HealthChecker defines the interface for dependency health checks.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles liveness and readiness probes.
type HealthHandler struct {
	db      HealthChecker
	version string
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db HealthChecker, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

type readinessResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version,omitempty"`
	Checks  map[string]string `json:"checks"`
}

// Liveness godoc
// @Summary      Liveness probe
// @Description  Returns 200 if the process is alive. Used by orchestrators for liveness checks.
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.HealthResponse
// @Router       /healthz [get]
func (h *HealthHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, healthResponse{
		Status:  "ok",
		Version: h.version,
	})
}

// Readiness godoc
// @Summary      Readiness probe
// @Description  Returns 200 if all dependencies (database) are healthy, 503 otherwise.
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.ReadinessResponse
// @Failure      503  {object}  response.ReadinessResponse
// @Router       /readyz [get]
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		checks["database"] = "healthy"
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	response.JSON(w, httpStatus, readinessResponse{
		Status:  status,
		Version: h.version,
		Checks:  checks,
	})
}
