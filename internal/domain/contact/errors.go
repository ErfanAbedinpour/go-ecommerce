package contact

import "app/pkg/apperror"

var (
	ErrNotFound       = apperror.NotFound("contact message")
	ErrInvalidSource  = apperror.Validation("invalid contact message source", nil)
	ErrInvalidStatus  = apperror.Validation("invalid contact message status", nil)
)
