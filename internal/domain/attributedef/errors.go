package attributedef

import "app/pkg/apperror"

var (
	ErrNotFound     = apperror.NotFound("product attribute")
	ErrSlugConflict = apperror.Conflict("attribute slug already exists")
	ErrNameConflict = apperror.Conflict("attribute name already exists")
	ErrHasValues    = apperror.Unprocessable("cannot delete attribute with existing values")
)
