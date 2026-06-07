package order

import "app/pkg/apperror"

var (
	ErrNotFound = apperror.NotFound("order")
)
