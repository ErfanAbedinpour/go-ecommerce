package blog

import "app/pkg/apperror"

var (
	ErrPostNotFound     = apperror.NotFound("blog post")
	ErrCategoryNotFound = apperror.NotFound("blog category")
	ErrCommentNotFound  = apperror.NotFound("blog comment")
	ErrDuplicateSlug    = apperror.Conflict("slug is already in use")
)
