package product

import "app/pkg/apperror"

// Status represents the lifecycle state of a product.
type Status string

const (
	StatusDraft    Status = "draft"
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// ParseStatus validates and parses a product status string.
func ParseStatus(value string) (Status, error) {
	switch Status(value) {
	case StatusDraft, StatusActive, StatusArchived:
		return Status(value), nil
	default:
		return "", apperror.Validation("invalid product status", map[string]string{
			"status": "must be one of: draft, active, archived",
		})
	}
}

// IsValid reports whether the status is a known value.
func (s Status) IsValid() bool {
	return s == StatusDraft || s == StatusActive || s == StatusArchived
}

// String returns the status as a plain string.
func (s Status) String() string {
	return string(s)
}
