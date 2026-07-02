package attributedef

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound     = apperror.NotFound("product attribute")
	ErrSlugConflict = apperror.ConflictKeyed(i18n.KeyAttributeSlugConflict, "attribute slug already exists")
	ErrNameConflict = apperror.ConflictKeyed(i18n.KeyAttributeNameConflict, "attribute name already exists")
	ErrHasValues    = apperror.UnprocessableKeyed(i18n.KeyAttributeHasValues, "cannot delete attribute with existing values")
)
