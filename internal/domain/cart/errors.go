package cart

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound    = apperror.NotFound("cart not found")
	ErrEmpty       = apperror.ValidationKeyed(i18n.KeyCartEmpty, "cart is empty", nil)
	ErrItemMissing = apperror.NotFound("cart item not found")
)
