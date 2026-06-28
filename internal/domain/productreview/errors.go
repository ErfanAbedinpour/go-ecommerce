package productreview

import "app/pkg/apperror"

var (
	ErrNotFound       = apperror.NotFound("product review")
	ErrAlreadyReviewed = apperror.Conflict("you have already reviewed this product")
)
