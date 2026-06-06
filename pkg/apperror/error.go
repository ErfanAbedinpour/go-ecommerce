package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes used across the application.
const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeUnprocessable    = "UNPROCESSABLE_ENTITY"
	CodeRateLimited      = "RATE_LIMITED"
	CodeInternal         = "INTERNAL_ERROR"
	CodeInvalidCreds     = "INVALID_CREDENTIALS"
	CodeAccountDisabled  = "ACCOUNT_DISABLED"
	CodeInvalidToken     = "INVALID_TOKEN"
	CodeTokenRevoked     = "TOKEN_REVOKED"
	CodeInvalidStatus    = "INVALID_STATUS_TRANSITION"
)

// AppError represents a structured application error with HTTP status mapping.
type AppError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Status  int               `json:"-"`
	Details map[string]string `json:"details,omitempty"`
	Err     error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// ErrorResponse is the JSON envelope for error responses.
type ErrorResponse struct {
	StatusCode int      `json:"statusCode"`
	Path       string   `json:"path"`
	Error      AppError `json:"error"`
}

// NewErrorResponse builds a client-facing error envelope.
func NewErrorResponse(appErr *AppError, path string) ErrorResponse {
	return ErrorResponse{
		StatusCode: appErr.Status,
		Path:       path,
		Error:      *appErr,
	}
}

// New creates a new AppError.
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// WithDetails adds field-level validation details.
func (e *AppError) WithDetails(details map[string]string) *AppError {
	e.Details = details
	return e
}

// WithError wraps an underlying error.
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// Validation creates a 400 validation error.
func Validation(message string, details map[string]string) *AppError {
	return &AppError{
		Code:    CodeValidation,
		Message: message,
		Status:  http.StatusBadRequest,
		Details: details,
	}
}

// Unauthorized creates a 401 error.
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a 403 error.
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

// NotFound creates a 404 error.
func NotFound(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// Conflict creates a 409 error.
func Conflict(message string) *AppError {
	return New(CodeConflict, message, http.StatusConflict)
}

// Unprocessable creates a 422 error.
func Unprocessable(message string) *AppError {
	return New(CodeUnprocessable, message, http.StatusUnprocessableEntity)
}

// Internal creates a 500 error.
func Internal(message string) *AppError {
	return New(CodeInternal, message, http.StatusInternalServerError)
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError extracts an AppError from an error chain.
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal("an unexpected error occurred").WithError(err)
}

// StatusCode returns the HTTP status code for an error.
func StatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Status
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status
	}
	return http.StatusInternalServerError
}
