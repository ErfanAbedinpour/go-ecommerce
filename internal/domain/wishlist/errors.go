package wishlist

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound      = apperror.NotFound("wishlist item")
	ErrAlreadyExists = apperror.ConflictKeyed(i18n.KeyWishlistAlreadyExists, "product already in wishlist")
)
