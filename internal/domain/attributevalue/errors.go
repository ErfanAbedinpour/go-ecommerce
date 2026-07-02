package attributevalue

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound         = apperror.NotFound("attribute value")
	ErrValueConflict    = apperror.ConflictKeyed(i18n.KeyAttributeValueConflict, "attribute value already exists")
	ErrAttributeMissing = apperror.NotFound("product attribute")
)
