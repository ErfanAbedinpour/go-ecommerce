package category

import "app/pkg/apperror"

var (
	ErrNotFound         = apperror.NotFound("category")
	ErrSlugConflict     = apperror.Conflict("category slug already exists")
	ErrParentNotFound   = apperror.NotFound("parent category")
	ErrInvalidParent    = apperror.Unprocessable("cannot set parent to self or descendant category")
	ErrHasChildren      = apperror.Unprocessable("cannot delete category with child categories")
	ErrHasProducts      = apperror.Unprocessable("cannot delete category with assigned products")
)
