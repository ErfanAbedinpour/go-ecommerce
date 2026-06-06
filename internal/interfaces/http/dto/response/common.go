package response

// ErrorBody is the structured error payload.
type ErrorBody struct {
	Code    string            `json:"code" example:"VALIDATION_ERROR"`
	Message string            `json:"message" example:"request validation failed"`
	Details map[string]string `json:"details,omitempty"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// HealthResponse is the liveness probe response.
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Version string `json:"version,omitempty" example:"1.0.0"`
}

// ReadinessResponse is the readiness probe response.
type ReadinessResponse struct {
	Status  string            `json:"status" example:"ready"`
	Version string            `json:"version,omitempty" example:"1.0.0"`
	Checks  map[string]string `json:"checks"`
}

// MessageResponse is a simple message payload.
type MessageResponse struct {
	Message string `json:"message" example:"ecommerce admin API v1"`
}
