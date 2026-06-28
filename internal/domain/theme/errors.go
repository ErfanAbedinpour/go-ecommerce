package theme

import "app/pkg/apperror"

var (
	ErrNotFound           = apperror.NotFound("theme")
	ErrStyleNotFound      = apperror.NotFound("store style")
	ErrAlreadyPurchased   = apperror.Conflict("theme already purchased")
	ErrNotPurchased       = apperror.Forbidden("theme must be purchased before activation")
	ErrThemeInactive      = apperror.Unprocessable("theme is not available")
)
