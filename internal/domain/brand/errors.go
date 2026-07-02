package brand

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound      = apperror.NotFound("brand")
	ErrSlugConflict  = apperror.ConflictKeyed(i18n.KeyBrandSlugConflict, "brand slug already exists")
	ErrNameConflict  = apperror.ConflictKeyed(i18n.KeyBrandNameConflict, "brand name already exists")
	ErrHasProducts   = apperror.UnprocessableKeyed(i18n.KeyBrandHasProducts, "cannot delete brand assigned to products")
)
