package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Error codes used across the application.
const (
	CodeValidation         = "VALIDATION_ERROR"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeUnprocessable       = "UNPROCESSABLE_ENTITY"
	CodeRateLimited        = "RATE_LIMITED"
	CodeInternal           = "INTERNAL_ERROR"
	CodeInvalidCreds       = "INVALID_CREDENTIALS"
	CodeAccountDisabled    = "ACCOUNT_DISABLED"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeTokenRevoked       = "TOKEN_REVOKED"
	CodeInvalidStatus      = "INVALID_STATUS_TRANSITION"
	CodeAccountExistsLogin = "ACCOUNT_EXISTS_LOGIN_REQUIRED"
)

const (
	keyErrorNotFound      = "error.not_found"
	keyErrorInternal      = "error.internal"
	keyErrorValidation    = "error.validation"
	keyErrorUnauthorized  = "error.unauthorized"
	keyErrorForbidden     = "error.forbidden"
	keyErrorConflict      = "error.conflict"
	keyErrorUnprocessable = "error.unprocessable"
)

// AppError represents a structured application error with HTTP status mapping.
type AppError struct {
	Code          string            `json:"code"`
	MessageKey    string            `json:"-"`
	MessageParams map[string]string `json:"-"`
	Message       string            `json:"message"`
	Status        int               `json:"-"`
	Details       map[string]string `json:"details,omitempty"`
	Err           error             `json:"-"`
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

// Keyed creates an AppError with an i18n message key.
func Keyed(code, messageKey, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		MessageKey: messageKey,
		Message:    message,
		Status:     status,
	}
}

// WithDetails adds field-level validation details.
func (e *AppError) WithDetails(details map[string]string) *AppError {
	e.Details = details
	return e
}

// WithParams adds template parameters for i18n lookup.
func (e *AppError) WithParams(params map[string]string) *AppError {
	e.MessageParams = params
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
		Code:       CodeValidation,
		MessageKey: keyErrorValidation,
		Message:    message,
		Status:     http.StatusBadRequest,
		Details:    details,
	}
}

// ValidationKeyed creates a validation error with a specific message key.
func ValidationKeyed(messageKey, message string, details map[string]string) *AppError {
	return &AppError{
		Code:       CodeValidation,
		MessageKey: messageKey,
		Message:    message,
		Status:     http.StatusBadRequest,
		Details:    details,
	}
}

// Unauthorized creates a 401 error.
func Unauthorized(message string) *AppError {
	return Keyed(CodeUnauthorized, keyErrorUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a 403 error.
func Forbidden(message string) *AppError {
	return Keyed(CodeForbidden, keyErrorForbidden, message, http.StatusForbidden)
}

// NotFound creates a 404 error.
func NotFound(resource string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		MessageKey: keyErrorNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		MessageParams: map[string]string{
			"resource": resourceLabelKey(resource),
		},
		Status: http.StatusNotFound,
	}
}

// Conflict creates a 409 error.
func Conflict(message string) *AppError {
	return Keyed(CodeConflict, keyErrorConflict, message, http.StatusConflict)
}

// ConflictKeyed creates a conflict error with a specific message key.
func ConflictKeyed(messageKey, message string) *AppError {
	return Keyed(CodeConflict, messageKey, message, http.StatusConflict)
}

// Unprocessable creates a 422 error.
func Unprocessable(message string) *AppError {
	return Keyed(CodeUnprocessable, keyErrorUnprocessable, message, http.StatusUnprocessableEntity)
}

// UnprocessableKeyed creates an unprocessable error with a specific message key.
func UnprocessableKeyed(messageKey, message string) *AppError {
	return Keyed(CodeUnprocessable, messageKey, message, http.StatusUnprocessableEntity)
}

// Internal creates a 500 error.
func Internal(message string) *AppError {
	return Keyed(CodeInternal, keyErrorInternal, message, http.StatusInternalServerError)
}

func resourceLabelKey(resource string) string {
	switch strings.ToLower(strings.TrimSpace(resource)) {
	case "user":
		return "resource.user"
	case "customer":
		return "resource.customer"
	case "product":
		return "resource.product"
	case "order":
		return "resource.order"
	case "category":
		return "resource.category"
	case "parent category":
		return "resource.parent_category"
	case "brand":
		return "resource.brand"
	case "coupon":
		return "resource.coupon"
	case "cart not found", "cart":
		return "resource.cart"
	case "cart item not found", "cart item":
		return "resource.cart_item"
	case "theme":
		return "resource.theme"
	case "store style":
		return "resource.store_style"
	case "wishlist item":
		return "resource.wishlist_item"
	case "product review":
		return "resource.product_review"
	case "product question":
		return "resource.product_question"
	case "blog post":
		return "resource.blog_post"
	case "blog category":
		return "resource.blog_category"
	case "blog comment":
		return "resource.blog_comment"
	case "contact message":
		return "resource.contact_message"
	case "product attribute":
		return "resource.product_attribute"
	case "attribute value":
		return "resource.attribute_value"
	case "store content":
		return "resource.store_content"
	case "storefront hero":
		return "resource.storefront_hero"
	case "product slide":
		return "resource.product_slide"
	case "pro banner":
		return "resource.pro_banner"
	case "partner brand":
		return "resource.partner_brand"
	case "homepage review":
		return "resource.homepage_review"
	case "faq item":
		return "resource.faq_item"
	case "slide item":
		return "resource.slide_item"
	default:
		return resource
	}
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
