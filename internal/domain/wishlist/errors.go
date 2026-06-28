package wishlist

import "app/pkg/apperror"

var (
	ErrNotFound      = apperror.NotFound("wishlist item")
	ErrAlreadyExists = apperror.Conflict("product already in wishlist")
)
