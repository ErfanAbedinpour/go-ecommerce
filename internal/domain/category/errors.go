package category

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound         = apperror.NotFound("category")
	ErrSlugConflict     = apperror.ConflictKeyed(i18n.KeyCategorySlugConflict, "category slug already exists")
	ErrParentNotFound   = apperror.NotFound("parent category")
	ErrInvalidParent    = apperror.UnprocessableKeyed(i18n.KeyCategoryInvalidParent, "cannot set parent to self or descendant category")
	ErrHasChildren      = apperror.UnprocessableKeyed(i18n.KeyCategoryHasChildren, "cannot delete category with child categories")
	ErrHasProducts      = apperror.UnprocessableKeyed(i18n.KeyCategoryHasProducts, "cannot delete category with assigned products")
)
