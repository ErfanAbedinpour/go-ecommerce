package theme

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound           = apperror.NotFound("theme")
	ErrStyleNotFound      = apperror.NotFound("store style")
	ErrAlreadyPurchased   = apperror.ConflictKeyed(i18n.KeyThemeAlreadyPurchased, "theme already purchased")
	ErrNotPurchased       = apperror.Keyed(apperror.CodeForbidden, i18n.KeyThemeNotPurchased, "theme must be purchased before activation", 403)
	ErrThemeInactive      = apperror.UnprocessableKeyed(i18n.KeyThemeInactive, "theme is not available")
)
