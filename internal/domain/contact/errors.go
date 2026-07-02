package contact

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound      = apperror.NotFound("contact message")
	ErrInvalidSource = apperror.ValidationKeyed(i18n.KeyContactInvalidSource, "invalid contact message source", nil)
	ErrInvalidStatus = apperror.ValidationKeyed(i18n.KeyContactInvalidStatus, "invalid contact message status", nil)
)
