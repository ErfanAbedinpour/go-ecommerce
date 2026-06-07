package brand

import "app/pkg/apperror"

var (
	ErrNotFound      = apperror.NotFound("brand")
	ErrSlugConflict  = apperror.Conflict("brand slug already exists")
	ErrNameConflict  = apperror.Conflict("brand name already exists")
	ErrHasProducts   = apperror.Unprocessable("cannot delete brand assigned to products")
)
