package customer

import "app/pkg/apperror"

var (
	ErrNotFound    = apperror.NotFound("customer")
	ErrEmailTaken  = apperror.Conflict("customer email is already in use")
	ErrHasOrders   = apperror.Unprocessable("cannot delete customer with existing orders")
	ErrInvalidType = apperror.Validation("invalid customer type", map[string]string{
		"type": "must be one of: registered, guest",
	})
)
