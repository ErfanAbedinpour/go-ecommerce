package customer

import "app/pkg/apperror"

var (
	ErrNotFound     = apperror.NotFound("customer")
	ErrInvalidType  = apperror.Validation("invalid customer type", map[string]string{
		"type": "must be one of: registered, guest",
	})
)
