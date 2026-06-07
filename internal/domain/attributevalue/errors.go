package attributevalue

import "app/pkg/apperror"

var (
	ErrNotFound         = apperror.NotFound("attribute value")
	ErrValueConflict    = apperror.Conflict("attribute value already exists")
	ErrAttributeMissing = apperror.NotFound("product attribute")
)
