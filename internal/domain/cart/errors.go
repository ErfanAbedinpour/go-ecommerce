package cart

import "app/pkg/apperror"

var (
	ErrNotFound    = apperror.NotFound("cart not found")
	ErrEmpty       = apperror.Validation("cart is empty", nil)
	ErrItemMissing = apperror.NotFound("cart item not found")
)
