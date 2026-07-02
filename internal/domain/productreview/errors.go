package productreview

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound        = apperror.NotFound("product review")
	ErrAlreadyReviewed = apperror.ConflictKeyed(i18n.KeyProductReviewAlreadyReviewed, "you have already reviewed this product")
)
